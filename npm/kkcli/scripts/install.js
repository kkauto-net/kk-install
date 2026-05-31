"use strict";

const crypto = require("node:crypto");
const fs = require("node:fs");
const https = require("node:https");
const os = require("node:os");
const path = require("node:path");
const { execFile } = require("node:child_process");
const { promisify } = require("node:util");

const execFileAsync = promisify(execFile);

const GITHUB_OWNER = "kkauto-net";
const GITHUB_REPO = "kk-install";
const PACKAGE_ROOT = path.resolve(__dirname, "..");
const VENDOR_DIR = path.join(PACKAGE_ROOT, "vendor");
const DEFAULT_TIMEOUT_MS = 30000;
const MAX_CHECKSUM_BYTES = 1024 * 1024;
const MAX_ARCHIVE_BYTES = 128 * 1024 * 1024;

function packageVersion(packageJSONPath = path.join(PACKAGE_ROOT, "package.json")) {
  const pkg = JSON.parse(fs.readFileSync(packageJSONPath, "utf8"));
  if (!pkg.version) {
    throw new Error("package.json version is missing");
  }
  return pkg.version;
}

function mapPlatform(platform = process.platform, arch = process.arch) {
  if (platform !== "linux") {
    throw new Error(`Unsupported platform: ${platform}. @kkauto/kkcli currently supports Linux only.`);
  }

  if (arch === "x64") {
    return { os: "linux", arch: "amd64" };
  }
  if (arch === "arm64") {
    return { os: "linux", arch: "arm64" };
  }
  throw new Error(`Unsupported CPU architecture: ${arch}. Supported architectures: x64, arm64.`);
}

function artifactName(version, target = mapPlatform()) {
  return `kkcli_${version}_${target.os}_${target.arch}.tar.gz`;
}

function releaseBaseURL(version) {
  return `https://github.com/${GITHUB_OWNER}/${GITHUB_REPO}/releases/download/v${version}`;
}

function parseChecksum(checksumsText, assetName) {
  for (const line of checksumsText.split(/\r?\n/)) {
    const fields = line.trim().split(/\s+/).filter(Boolean);
    if (fields.length < 2 || fields[1] !== assetName) {
      continue;
    }
    const checksum = fields[0];
    if (!/^[a-fA-F0-9]{64}$/.test(checksum)) {
      throw new Error(`Malformed SHA256 checksum for ${assetName}`);
    }
    return checksum.toLowerCase();
  }
  throw new Error(`Checksum entry not found for ${assetName}`);
}

function sha256File(filePath) {
  return crypto.createHash("sha256").update(fs.readFileSync(filePath)).digest("hex");
}

function verifyChecksum(checksumsText, assetName, archivePath) {
  const expected = parseChecksum(checksumsText, assetName);
  const actual = sha256File(archivePath);
  if (actual !== expected) {
    throw new Error(`Checksum mismatch for ${assetName}`);
  }
  return actual;
}

function download(url, destination, options = {}, redirectCount = 0) {
  if (redirectCount > 5) {
    return Promise.reject(new Error(`Too many redirects while downloading ${url}`));
  }

  const timeoutMs = options.timeoutMs || DEFAULT_TIMEOUT_MS;
  const maxBytes = options.maxBytes || MAX_ARCHIVE_BYTES;
  fs.mkdirSync(path.dirname(destination), { recursive: true });
  const temporaryDestination = `${destination}.tmp-${process.pid}`;

  return new Promise((resolve, reject) => {
    let settled = false;
    let bytes = 0;
    let request;
    const deadline = setTimeout(() => {
      if (request) {
        request.destroy(new Error(`Download timed out for ${url}`));
      }
    }, timeoutMs);

    function fail(error) {
      if (settled) {
        return;
      }
      settled = true;
      clearTimeout(deadline);
      fs.rmSync(temporaryDestination, { force: true });
      reject(error);
    }

    request = https.get(url, (response) => {
      const statusCode = response.statusCode || 0;
      if (statusCode >= 300 && statusCode < 400 && response.headers.location) {
        response.resume();
        const redirectedURL = new URL(response.headers.location, url).toString();
        settled = true;
        clearTimeout(deadline);
        download(redirectedURL, destination, options, redirectCount + 1).then(resolve, reject);
        return;
      }

      if (statusCode !== 200) {
        response.resume();
        fail(new Error(`Download failed for ${url}: HTTP ${statusCode}`));
        return;
      }

      const contentLength = Number(response.headers["content-length"] || 0);
      if (contentLength > maxBytes) {
        response.resume();
        fail(new Error(`Download exceeded size limit for ${url}`));
        return;
      }

      const file = fs.createWriteStream(temporaryDestination, { mode: 0o600 });
      response.on("data", (chunk) => {
        bytes += chunk.length;
        if (bytes > maxBytes) {
          request.destroy(new Error(`Download exceeded size limit for ${url}`));
        }
      });
      response.pipe(file);
      file.on("finish", () => {
        file.close((error) => {
          if (error) {
            fail(error);
            return;
          }
          if (settled) {
            return;
          }
          settled = true;
          clearTimeout(deadline);
          fs.renameSync(temporaryDestination, destination);
          resolve(destination);
        });
      });
      file.on("error", (error) => {
        fail(error);
      });
    });

    request.setTimeout(timeoutMs, () => {
      request.destroy(new Error(`Download timed out for ${url}`));
    });
    request.on("error", (error) => {
      fail(error);
    });
  });
}

async function listArchiveEntries(archivePath, exec = execFileAsync) {
  const { stdout } = await exec("tar", ["-tzf", archivePath], { maxBuffer: 1024 * 1024 });
  return stdout.split(/\r?\n/).filter(Boolean);
}

async function listArchiveMetadata(archivePath, exec = execFileAsync) {
  const { stdout } = await exec("tar", ["-tvzf", archivePath], { maxBuffer: 1024 * 1024 });
  return stdout.split(/\r?\n/).filter(Boolean);
}

function validateArchiveEntries(entries) {
  for (const entry of entries) {
    const normalized = entry.replace(/\\/g, "/");
    const parts = normalized.split("/");
    if (normalized.startsWith("/") || parts.includes("..")) {
      throw new Error(`Unsafe archive entry: ${entry}`);
    }
  }
}

function validateArchiveEntryTypes(metadataLines) {
  for (const line of metadataLines) {
    const type = line[0];
    if (type !== "-" && type !== "d") {
      throw new Error("Archive contains unsupported link or special-file entry");
    }
  }
}

function assertPathInside(childPath, parentPath) {
  const realChild = fs.realpathSync(childPath);
  const realParent = fs.realpathSync(parentPath);
  const relative = path.relative(realParent, realChild);
  if (relative === "" || (!relative.startsWith("..") && !path.isAbsolute(relative))) {
    return;
  }
  throw new Error(`Extracted file escaped archive directory: ${childPath}`);
}

function findExtractedBinary(directory) {
  for (const entry of fs.readdirSync(directory, { withFileTypes: true })) {
    const fullPath = path.join(directory, entry.name);
    if (entry.isDirectory()) {
      const found = findExtractedBinary(fullPath);
      if (found) {
        return found;
      }
      continue;
    }
    if (entry.isFile() && entry.name === "kk") {
      return fullPath;
    }
  }
  return "";
}

async function extractBinary(archivePath, destinationDir = VENDOR_DIR, exec = execFileAsync) {
  const temporaryDirectory = fs.mkdtempSync(path.join(os.tmpdir(), "kkcli-extract-"));
  try {
    const metadata = await listArchiveMetadata(archivePath, exec);
    validateArchiveEntryTypes(metadata);
    const entries = await listArchiveEntries(archivePath, exec);
    validateArchiveEntries(entries);
    await exec("tar", ["-xzf", archivePath, "-C", temporaryDirectory], { maxBuffer: 1024 * 1024 });

    const extractedBinary = findExtractedBinary(temporaryDirectory);
    if (!extractedBinary) {
      throw new Error("Archive did not contain kk binary");
    }
    assertPathInside(extractedBinary, temporaryDirectory);

    fs.mkdirSync(destinationDir, { recursive: true });
    const destination = path.join(destinationDir, "kk");
    fs.copyFileSync(extractedBinary, destination);
    fs.chmodSync(destination, 0o755);
    return destination;
  } finally {
    fs.rmSync(temporaryDirectory, { recursive: true, force: true });
  }
}

async function install(options = {}) {
  const version = options.version || packageVersion();
  const target = options.target || mapPlatform();
  const asset = artifactName(version, target);
  const baseURL = options.baseURL || releaseBaseURL(version);
  const workDirectory = fs.mkdtempSync(path.join(os.tmpdir(), "kkcli-install-"));

  try {
    const checksumsPath = path.join(workDirectory, "checksums.txt");
    const archivePath = path.join(workDirectory, asset);

    console.log(`Downloading kkcli ${version} for ${target.os}/${target.arch}...`);
    await download(`${baseURL}/checksums.txt`, checksumsPath, { maxBytes: MAX_CHECKSUM_BYTES });
    await download(`${baseURL}/${asset}`, archivePath, { maxBytes: MAX_ARCHIVE_BYTES });

    verifyChecksum(fs.readFileSync(checksumsPath, "utf8"), asset, archivePath);
    await extractBinary(archivePath, options.destinationDir || VENDOR_DIR);
    console.log("kkcli binary installed.");
  } finally {
    fs.rmSync(workDirectory, { recursive: true, force: true });
  }
}

async function main() {
  await install();
}

if (require.main === module) {
  main().catch((error) => {
    console.error(`kkcli npm install failed: ${error.message}`);
    process.exit(1);
  });
}

module.exports = {
  artifactName,
  extractBinary,
  install,
  listArchiveMetadata,
  listArchiveEntries,
  mapPlatform,
  packageVersion,
  parseChecksum,
  releaseBaseURL,
  sha256File,
  validateArchiveEntryTypes,
  validateArchiveEntries,
  verifyChecksum,
};

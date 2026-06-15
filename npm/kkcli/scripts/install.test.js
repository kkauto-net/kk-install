"use strict";

const assert = require("node:assert/strict");
const crypto = require("node:crypto");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const { execFileSync, spawnSync } = require("node:child_process");
const test = require("node:test");

const {
  artifactName,
  extractBinary,
  mapPlatform,
  parseChecksum,
  releaseBaseURL,
  sha256File,
  validateArchiveEntryTypes,
  validateArchiveEntries,
  verifyChecksum,
} = require("./install");

function tempDir() {
  return fs.mkdtempSync(path.join(os.tmpdir(), "kkcli-test-"));
}

function sha256(data) {
  return crypto.createHash("sha256").update(data).digest("hex");
}

function hasTar() {
  return spawnSync("tar", ["--version"], { stdio: "ignore" }).status === 0;
}

test("mapPlatform supports current Linux release targets", () => {
  assert.deepEqual(mapPlatform("linux", "x64"), { os: "linux", arch: "amd64" });
  assert.deepEqual(mapPlatform("linux", "arm64"), { os: "linux", arch: "arm64" });
});

test("mapPlatform rejects unsupported OS and CPU before network calls", () => {
  assert.throws(() => mapPlatform("darwin", "x64"), /supports Linux only/);
  assert.throws(() => mapPlatform("linux", "arm"), /Unsupported CPU architecture/);
});

test("artifactName and releaseBaseURL match GoReleaser release contract", () => {
  assert.equal(artifactName("0.3.3", { os: "linux", arch: "amd64" }), "kkcli_0.3.3_linux_amd64.tar.gz");
  assert.equal(
    releaseBaseURL("0.3.3"),
    "https://github.com/kkauto-net/kk-install/releases/download/v0.3.3",
  );
});

test("parseChecksum requires exact filename match", () => {
  const asset = "kkcli_0.3.3_linux_amd64.tar.gz";
  const wrong = "a".repeat(64);
  const right = "b".repeat(64);
  const checksums = `${wrong} ${asset}.old\n${right} ${asset}\n`;

  assert.equal(parseChecksum(checksums, asset), right);
});

test("parseChecksum lowercases uppercase SHA256 and tolerates trailing fields", () => {
  const asset = "kkcli_0.3.3_linux_arm64.tar.gz";
  const uppercase = "A".repeat(64);

  assert.equal(parseChecksum(`${uppercase} ${asset} ignored-field\n`, asset), uppercase.toLowerCase());
});

test("parseChecksum fails closed on missing and malformed entries", () => {
  const asset = "kkcli_0.3.3_linux_amd64.tar.gz";

  assert.throws(() => parseChecksum("", asset), /not found/);
  assert.throws(() => parseChecksum(`not-a-sha ${asset}\n`, asset), /Malformed SHA256/);
});

test("sha256File and verifyChecksum detect matches and mismatches", () => {
  const directory = tempDir();
  try {
    const archive = path.join(directory, "kkcli_0.3.3_linux_amd64.tar.gz");
    fs.writeFileSync(archive, "release archive");
    const expected = sha256("release archive");

    assert.equal(sha256File(archive), expected);
    assert.equal(verifyChecksum(`${expected} ${path.basename(archive)}\n`, path.basename(archive), archive), expected);
    assert.throws(() => verifyChecksum(`${"0".repeat(64)} ${path.basename(archive)}\n`, path.basename(archive), archive), /Checksum mismatch/);
  } finally {
    fs.rmSync(directory, { recursive: true, force: true });
  }
});

test("validateArchiveEntries rejects traversal entries", () => {
  assert.doesNotThrow(() => validateArchiveEntries(["kk", "docs/readme.txt"]));
  assert.throws(() => validateArchiveEntries(["../kk"]), /Unsafe archive entry/);
  assert.throws(() => validateArchiveEntries(["/tmp/kk"]), /Unsafe archive entry/);
});

test("validateArchiveEntryTypes rejects links and special files", () => {
  assert.doesNotThrow(() => validateArchiveEntryTypes(["-rwxr-xr-x user group 1 2026-01-01 00:00 kk", "drwxr-xr-x user group 0 2026-01-01 00:00 docs/"]));
  assert.throws(() => validateArchiveEntryTypes(["lrwxrwxrwx user group 1 2026-01-01 00:00 kk -> /tmp/kk"]), /unsupported link/);
  assert.throws(() => validateArchiveEntryTypes(["hrwxrwxrwx user group 1 2026-01-01 00:00 kk link to /tmp/kk"]), /unsupported link/);
});

test(
  "extractBinary extracts kk from a local fixture archive",
  { skip: hasTar() ? false : "tar is not available" },
  async () => {
    const directory = tempDir();
    try {
      const source = path.join(directory, "source");
      const destination = path.join(directory, "vendor");
      const archive = path.join(directory, "fixture.tar.gz");
      fs.mkdirSync(source);
      fs.writeFileSync(path.join(source, "kk"), "#!/bin/sh\necho kk\n", { mode: 0o755 });
      execFileSync("tar", ["-czf", archive, "-C", source, "kk"]);

      const installed = await extractBinary(archive, destination);
      assert.equal(installed, path.join(destination, "kk"));
      assert.equal(fs.readFileSync(installed, "utf8"), "#!/bin/sh\necho kk\n");
      assert.equal(fs.statSync(installed).mode & 0o111, 0o111);
    } finally {
      fs.rmSync(directory, { recursive: true, force: true });
    }
  },
);

test("package files includes bootstrap.js required by install.js", () => {
  const pkg = JSON.parse(fs.readFileSync(path.join(__dirname, "..", "package.json"), "utf8"));
  assert.ok(pkg.files.includes("scripts/bootstrap.js"));
});

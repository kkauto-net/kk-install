"use strict";

const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const { spawnSync } = require("node:child_process");

function bootstrapWorkingDirectory() {
  return process.env.INIT_CWD || process.cwd();
}

function hasUnattendedEnv() {
  return Boolean(process.env.KKAUTO_LICENSE && process.env.KK_DOMAIN && process.env.KK_LANGUAGE);
}

function shouldBootstrap() {
  return process.env.KK_SKIP_BOOTSTRAP !== "1";
}

function runKK(kkPath, args, cwd) {
  return spawnSync(kkPath, args, {
    cwd,
    stdio: "inherit",
    env: process.env,
  });
}

function createTempLicenseFile(license) {
  const directory = fs.mkdtempSync(path.join(os.tmpdir(), "kkcli-license-"));
  const licenseFile = path.join(directory, "license");
  fs.writeFileSync(licenseFile, `${license}\n`, { mode: 0o600 });
  return { directory, licenseFile };
}

function bootstrap(kkPath, options = {}) {
  if (!shouldBootstrap()) {
    console.log("Bootstrap skipped (KK_SKIP_BOOTSTRAP=1).");
    return { skipped: true, ok: true };
  }

  const cwd = options.cwd || bootstrapWorkingDirectory();
  const spawn = options.spawn || runKK;

  if (process.stdin.isTTY) {
    console.log("Starting interactive setup...");
    const initResult = spawn(kkPath, ["init"], cwd);
    if (initResult.status !== 0) {
      return { ok: false, step: "init", status: initResult.status ?? 1 };
    }
    const startResult = spawn(kkPath, ["start"], cwd);
    if (startResult.status !== 0) {
      return { ok: false, step: "start", status: startResult.status ?? 1 };
    }
    return { ok: true };
  }

  if (hasUnattendedEnv()) {
    console.log("Starting unattended setup...");
    const { directory, licenseFile } = createTempLicenseFile(process.env.KKAUTO_LICENSE);
    try {
      const initResult = spawn(
        kkPath,
        [
          "init",
          "--yes",
          "--install-docker",
          "--license-file",
          licenseFile,
          "--domain",
          process.env.KK_DOMAIN,
          "--language",
          process.env.KK_LANGUAGE,
        ],
        cwd,
      );
      if (initResult.status !== 0) {
        return { ok: false, step: "init", status: initResult.status ?? 1 };
      }
      const startResult = spawn(kkPath, ["start"], cwd);
      if (startResult.status !== 0) {
        return { ok: false, step: "start", status: startResult.status ?? 1 };
      }
      return { ok: true };
    } finally {
      fs.rmSync(directory, { recursive: true, force: true });
    }
  }

  console.warn("Non-interactive shell: set KKAUTO_LICENSE, KK_DOMAIN, and KK_LANGUAGE for auto setup.");
  console.warn("Or run manually: kk init && kk start");
  return { ok: true, warned: true };
}

module.exports = {
  bootstrap,
  bootstrapWorkingDirectory,
  createTempLicenseFile,
  hasUnattendedEnv,
  shouldBootstrap,
};

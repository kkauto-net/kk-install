#!/usr/bin/env node
"use strict";

const fs = require("node:fs");
const path = require("node:path");
const { spawnSync } = require("node:child_process");

function binaryPath(packageRoot = path.resolve(__dirname, "..")) {
  return path.join(packageRoot, "vendor", "kk");
}

function run(argv = process.argv.slice(2), env = process.env) {
  const kkPath = env.KKCLI_BINARY_PATH || binaryPath();
  if (!fs.existsSync(kkPath)) {
    console.error(
      "kk binary is missing. Reinstall with `npm install -g @kkauto/kkcli` or check the npm postinstall logs.",
    );
    return 1;
  }

  const result = spawnSync(kkPath, argv, { stdio: "inherit" });
  if (result.error) {
    console.error(`failed to execute kk binary: ${result.error.message}`);
    return 1;
  }
  if (result.signal) {
    process.kill(process.pid, result.signal);
    return 1;
  }
  return result.status === null ? 1 : result.status;
}

if (require.main === module) {
  process.exit(run());
}

module.exports = { binaryPath, run };

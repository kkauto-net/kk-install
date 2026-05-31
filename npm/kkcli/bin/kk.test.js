"use strict";

const assert = require("node:assert/strict");
const path = require("node:path");
const { spawnSync } = require("node:child_process");
const test = require("node:test");

const { binaryPath } = require("./kk");

test("binaryPath resolves vendor kk path", () => {
  assert.equal(binaryPath("/tmp/pkg"), path.join("/tmp/pkg", "vendor", "kk"));
});

test("bin wrapper fails clearly when kk binary is missing", () => {
  const result = spawnSync(process.execPath, [path.join(__dirname, "kk.js")], {
    env: { ...process.env, KKCLI_BINARY_PATH: path.join(__dirname, "missing-kk") },
    encoding: "utf8",
  });

  assert.equal(result.status, 1);
  assert.match(result.stderr, /kk binary is missing/);
});

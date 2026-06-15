"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const test = require("node:test");

const {
  bootstrap,
  bootstrapWorkingDirectory,
  createTempLicenseFile,
  hasUnattendedEnv,
  shouldBootstrap,
} = require("./bootstrap");

test("shouldBootstrap respects KK_SKIP_BOOTSTRAP", () => {
  const previous = process.env.KK_SKIP_BOOTSTRAP;
  try {
    delete process.env.KK_SKIP_BOOTSTRAP;
    assert.equal(shouldBootstrap(), true);
    process.env.KK_SKIP_BOOTSTRAP = "1";
    assert.equal(shouldBootstrap(), false);
  } finally {
    if (previous === undefined) {
      delete process.env.KK_SKIP_BOOTSTRAP;
    } else {
      process.env.KK_SKIP_BOOTSTRAP = previous;
    }
  }
});

test("hasUnattendedEnv requires all bootstrap variables", () => {
  const previous = {
    license: process.env.KKAUTO_LICENSE,
    domain: process.env.KK_DOMAIN,
    language: process.env.KK_LANGUAGE,
  };
  try {
    delete process.env.KKAUTO_LICENSE;
    delete process.env.KK_DOMAIN;
    delete process.env.KK_LANGUAGE;
    assert.equal(hasUnattendedEnv(), false);

    process.env.KKAUTO_LICENSE = "LICENSE-ABCDEF0123456789";
    process.env.KK_DOMAIN = "example.com";
    process.env.KK_LANGUAGE = "en";
    assert.equal(hasUnattendedEnv(), true);
  } finally {
    for (const [key, value] of Object.entries(previous)) {
      const envKey = key === "license" ? "KKAUTO_LICENSE" : key === "domain" ? "KK_DOMAIN" : "KK_LANGUAGE";
      if (value === undefined) {
        delete process.env[envKey];
      } else {
        process.env[envKey] = value;
      }
    }
  }
});

test("bootstrapWorkingDirectory prefers INIT_CWD", () => {
  const previous = process.env.INIT_CWD;
  try {
    process.env.INIT_CWD = "/srv/kkengine";
    assert.equal(bootstrapWorkingDirectory(), "/srv/kkengine");
  } finally {
    if (previous === undefined) {
      delete process.env.INIT_CWD;
    } else {
      process.env.INIT_CWD = previous;
    }
  }
});

test("createTempLicenseFile writes secure temporary license file", () => {
  const { directory, licenseFile } = createTempLicenseFile("LICENSE-ABCDEF0123456789");
  try {
    assert.equal(fs.readFileSync(licenseFile, "utf8"), "LICENSE-ABCDEF0123456789\n");
    assert.equal(fs.statSync(licenseFile).mode & 0o777, 0o600);
  } finally {
    fs.rmSync(directory, { recursive: true, force: true });
  }
});

test("bootstrap skips when opted out", () => {
  const previous = process.env.KK_SKIP_BOOTSTRAP;
  const calls = [];
  try {
    process.env.KK_SKIP_BOOTSTRAP = "1";
    const result = bootstrap("/tmp/kk", {
      spawn: (...args) => {
        calls.push(args);
        return { status: 0 };
      },
    });
    assert.equal(result.skipped, true);
    assert.equal(result.ok, true);
    assert.equal(calls.length, 0);
  } finally {
    if (previous === undefined) {
      delete process.env.KK_SKIP_BOOTSTRAP;
    } else {
      process.env.KK_SKIP_BOOTSTRAP = previous;
    }
  }
});

test("bootstrap unattended uses INIT_CWD and install-docker flags", () => {
  const previous = {
    skip: process.env.KK_SKIP_BOOTSTRAP,
    license: process.env.KKAUTO_LICENSE,
    domain: process.env.KK_DOMAIN,
    language: process.env.KK_LANGUAGE,
    initCwd: process.env.INIT_CWD,
    isTTY: process.stdin.isTTY,
  };
  const calls = [];
  const cwd = fs.mkdtempSync(path.join(os.tmpdir(), "kkcli-bootstrap-"));
  try {
    delete process.env.KK_SKIP_BOOTSTRAP;
    process.env.KKAUTO_LICENSE = "LICENSE-ABCDEF0123456789";
    process.env.KK_DOMAIN = "example.com";
    process.env.KK_LANGUAGE = "vi";
    process.env.INIT_CWD = cwd;
    Object.defineProperty(process.stdin, "isTTY", { configurable: true, value: false });

    const result = bootstrap("/tmp/kk", {
      spawn: (kkPath, args, workingDirectory) => {
        calls.push({ kkPath, args, workingDirectory });
        return { status: 0 };
      },
    });

    assert.equal(result.ok, true);
    assert.equal(calls.length, 2);
    assert.equal(calls[0].workingDirectory, cwd);
    assert.deepEqual(calls[0].args.slice(0, 4), ["init", "--yes", "--install-docker", "--license-file"]);
    assert.equal(calls[0].args[calls[0].args.indexOf("--domain") + 1], "example.com");
    assert.equal(calls[0].args[calls[0].args.indexOf("--language") + 1], "vi");
    assert.deepEqual(calls[1].args, ["start"]);
  } finally {
    fs.rmSync(cwd, { recursive: true, force: true });
    Object.defineProperty(process.stdin, "isTTY", { configurable: true, value: previous.isTTY });
    for (const [key, value] of Object.entries(previous)) {
      const envKey =
        key === "skip"
          ? "KK_SKIP_BOOTSTRAP"
          : key === "license"
            ? "KKAUTO_LICENSE"
            : key === "domain"
              ? "KK_DOMAIN"
              : key === "language"
                ? "KK_LANGUAGE"
                : "INIT_CWD";
      if (value === undefined) {
        delete process.env[envKey];
      } else {
        process.env[envKey] = value;
      }
    }
  }
});

test("bootstrap warns without env in non-interactive mode", () => {
  const previous = {
    skip: process.env.KK_SKIP_BOOTSTRAP,
    license: process.env.KKAUTO_LICENSE,
    domain: process.env.KK_DOMAIN,
    language: process.env.KK_LANGUAGE,
    isTTY: process.stdin.isTTY,
  };
  const warnings = [];
  const originalWarn = console.warn;
  try {
    delete process.env.KK_SKIP_BOOTSTRAP;
    delete process.env.KKAUTO_LICENSE;
    delete process.env.KK_DOMAIN;
    delete process.env.KK_LANGUAGE;
    Object.defineProperty(process.stdin, "isTTY", { configurable: true, value: false });
    console.warn = (...args) => warnings.push(args.join(" "));

    const result = bootstrap("/tmp/kk", {
      spawn: () => ({ status: 0 }),
    });

    assert.equal(result.ok, true);
    assert.equal(result.warned, true);
    assert.match(warnings.join("\n"), /KKAUTO_LICENSE/);
  } finally {
    console.warn = originalWarn;
    Object.defineProperty(process.stdin, "isTTY", { configurable: true, value: previous.isTTY });
    for (const [key, value] of Object.entries(previous)) {
      const envKey =
        key === "skip"
          ? "KK_SKIP_BOOTSTRAP"
          : key === "license"
            ? "KKAUTO_LICENSE"
            : key === "domain"
              ? "KK_DOMAIN"
              : "KK_LANGUAGE";
      if (value === undefined) {
        delete process.env[envKey];
      } else {
        process.env[envKey] = value;
      }
    }
  }
});

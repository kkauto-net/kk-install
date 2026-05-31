# @kkauto/kkcli

`kkcli` is the command-line installer for the kkauto automation social system from [kkauto.net](https://kkauto.net). It installs the `kk` command, which helps provision and operate kkengine Docker Compose stacks on Linux servers.

Use this package when you want to install `kk` through npm instead of the shell installer. The package is a thin distribution wrapper: it downloads the official Linux release binary, verifies its SHA256 checksum, and exposes the `kk` command globally.

## Install

```bash
npm install -g @kkauto/kkcli
kk --version
```

## What kkcli Does

- Installs and initializes the kkauto automation social stack.
- Renders kkengine Docker Compose configuration for VPS/server deployments.
- Starts, stops, restarts, updates, and checks status for the kkengine stack.
- Supports unattended provisioning for automation scripts.

For product information, visit [kkauto.net](https://kkauto.net). For source code and releases, visit [github.com/kkauto-net/kk-install](https://github.com/kkauto-net/kk-install).

## Platform Support

This package currently supports Linux `x64` and Linux `arm64`, matching the GitHub Release artifacts produced by GoReleaser.

## Integrity Model

During `postinstall`, the package downloads the GitHub Release archive that matches this npm package version, downloads the release `checksums.txt`, verifies the archive SHA256 by exact filename, then extracts the `kk` binary into `vendor/kk`.

Install fails closed when the platform is unsupported, release metadata is missing, checksum metadata is malformed, the artifact entry is absent, checksum verification fails, or extraction fails.

Release signature verification is not implemented.

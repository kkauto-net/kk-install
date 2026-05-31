# @kkauto/kkcli

Linux-only npm distribution wrapper for `kkcli`. The installed command is `kk`.

## Install

```bash
npm install -g @kkauto/kkcli
kk --version
```

## Platform Support

This package currently supports Linux `x64` and Linux `arm64`, matching the GitHub Release artifacts produced by GoReleaser.

## Integrity Model

During `postinstall`, the package downloads the GitHub Release archive that matches this npm package version, downloads the release `checksums.txt`, verifies the archive SHA256 by exact filename, then extracts the `kk` binary into `vendor/kk`.

Install fails closed when the platform is unsupported, release metadata is missing, checksum metadata is malformed, the artifact entry is absent, checksum verification fails, or extraction fails.

Release signature verification is not implemented.

# Project Roadmap

## Overview

This roadmap tracks known product, reliability, security, and documentation work for `kkcli`. It reflects current code and workflow findings only; it is not a release commitment.

## Completed

| Item | Evidence |
|---|---|
| Unattended init | `kk init --yes --license-file ... --domain ... --language ...` exists. |
| Secret-safe license source | `--license-file` and `--license-stdin` avoid argv license exposure. |
| Deterministic init exit codes | `cmd/exit_error.go` and `cmd/root.go`. |
| kkengine lifecycle commands | `start`, `stop`, `restart`, `remove`, `status`, `update`. |
| n8n command group | `cmd/n8n*.go` and `pkg/n8n/*`. |
| Installer fail-closed checksum support | `scripts/install.sh` requires matching `checksums.txt` SHA256 verification before install. |
| Self-update fail-closed checksum support | `pkg/selfupdate` requires matching release `checksums.txt` SHA256 verification before binary replacement. |
| Draft-release changelog outputs | `.github/workflows/draft-release.yml` sets `previous_tag`, `compare_url`, and `changelog` outputs before creating the draft release. |
| MariaDB port contract aligned | `pkg/validator/ports.go`, `pkg/ui/table.go`, and generated templates use `3306`; template contract tests cover drift. |
| Release workflow test scope aligned | `release.yml` and `draft-release.yml` run `go test -v ./...`. |
| Template workflow Go version aligned | `validate-templates.yml` uses `go-version-file: go.mod`. |
| Docker-free binary smoke gate | `make test-smoke` and CI verify core command wiring without Docker. |
| Nightly/manual Compose e2e | `.github/workflows/e2e-compose.yml` runs full lifecycle with cleanup and redacted diagnostics. |
| License host identity template | Generated kkengine Compose mounts `/etc/machine-id` read-only for v2 license hardware identity. |
| npm distribution wrapper | `npm/kkcli` provides a Linux-only npm install channel for verified release binaries. |

## Near-Term Priorities

| Priority | Item | Reason |
|---:|---|---|
| P1 | Decide published platform matrix. | GoReleaser currently publishes Linux `amd64`/`arm64` only. |
| P1 | Keep release integrity guidance explicit. | Current releases use SHA256 checksums only; no signature verification is implemented. |
| P1 | Complete first npm publish setup. | Confirm `@kkauto` scope ownership and configure trusted publishing or `NPM_TOKEN`. |

## Product Enhancements

| Item | Notes |
|---|---|
| Document migration and rollback paths | Especially for `kk update`, `kk selfupdate`, and volume removal. |
| Add stronger n8n automation contract | n8n `install -f` exists, but no `--yes`/domain/language contract like `kk init`. |
| Improve config security posture if secrets are added | Current config is `0644` and should remain non-secret unless migrated. |
| Release platform expansion | Add macOS artifacts only if GoReleaser and installer support are updated and tested. |

## Maintenance Backlog

| Item | Trigger |
|---|---|
| Keep README command list synced | Any `cmd/*.go` flag or command change. |
| Refresh codebase summary | Significant package, command, or workflow changes. |
| Keep docs links valid | Any docs file rename or split. |
| Revisit generated file permissions | Any template or config write-path change. |

## References

- [Project Overview and PDR](./project-overview-pdr.md)
- [System Architecture](./system-architecture.md)
- [Deployment Guide](./deployment-guide.md)

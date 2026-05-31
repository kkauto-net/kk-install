# Project Overview and PDR

## Overview

`kkcli` is a Go command-line installer and operator for kkengine Docker Compose deployments. It installs as the `kk` binary and manages stack initialization, lifecycle commands, image updates, health/status output, CLI self-updates, shell completion, local config display, and an optional n8n automation stack.

## Product Goals

| Goal | Requirement |
|---|---|
| Fast setup | Render kkengine Compose and config files from embedded templates. |
| Safe operation | Validate Docker, Compose, ports, env files, disk space, and config before starting services. |
| Automation-ready installs | Support non-interactive provisioning with deterministic exit codes. |
| Secret-safe defaults | Generate crypto-random secrets and write generated `.env` files as `0600`. |
| Operator clarity | Provide concise terminal UI, status tables, health checks, and actionable suggestions. |
| Extensible operations | Keep kkengine and n8n stack management separated but consistent. |

## Primary Users

| User | Needs |
|---|---|
| VPS operator | Install kkengine, start/stop services, inspect status, and update containers. |
| Provisioning automation | Run unattended `kk init` over SSH without leaking license keys in argv. |
| Maintainer | Add commands, validations, templates, tests, and release assets safely. |

## Functional Requirements

| ID | Requirement | Status | Evidence |
|---|---|---|---|
| PDR-001 | Initialize kkengine stack with `kk init`. | Implemented | `cmd/init.go` |
| PDR-002 | Support `--force`, `--yes`, license, domain, and language init flags. | Implemented | `cmd/init.go`, `cmd/init_options.go` |
| PDR-003 | Validate licenses through `https://kkauto.net/api/license/config`. | Implemented | `pkg/license/license.go` |
| PDR-004 | Start, stop, restart, remove, update, and show status for kkengine. | Implemented | `cmd/start.go`, `cmd/stop.go`, `cmd/restart.go`, `cmd/remove.go`, `cmd/update.go`, `cmd/status.go` |
| PDR-005 | Manage local CLI config with `kk config show`. | Implemented | `cmd/config.go`, `pkg/config/config.go` |
| PDR-006 | Self-update from GitHub releases with `kk selfupdate`. | Implemented | `cmd/selfupdate.go`, `pkg/selfupdate/selfupdate.go` |
| PDR-007 | Generate shell completions for Bash, Zsh, and Fish. | Implemented | `cmd/completion.go` |
| PDR-008 | Manage n8n install/lifecycle/logs/update/remove commands. | Implemented | `cmd/n8n*.go`, `pkg/n8n/*` |
| PDR-009 | Generate kkengine stack files: `docker-compose.yml`, `.env`, `kkphp.conf`, optional `Caddyfile`, optional `kkfiler.toml`. | Implemented | `pkg/templates/embed.go` |
| PDR-010 | Generate n8n `docker-compose.yml` and `.env` under the configured n8n directory. | Implemented | `pkg/n8n/templates.go` |
| PDR-011 | Provide npm install channel for Linux release binaries. | Implemented | `npm/kkcli` |

## Unattended Init Contract

Use this automation-safe form:

```bash
kk init --yes --license-file /path/to/license --domain example.com --language en
```

| Rule | Status |
|---|---|
| Exactly one license source is required: `--license-file`, `--license-stdin`, or legacy `--license`. | Implemented |
| `--license-file` is recommended for automation; `--license` is kept for compatibility only. | Documented |
| `--domain` is required in unattended mode. | Implemented |
| `--language` must be `en` or `vi`. | Implemented |
| License/file/stdin input is capped at 4096 bytes. | Implemented |
| Non-interactive validation uses deterministic exit codes. | Implemented |

## Exit Codes

| Code | Meaning |
|---:|---|
| `0` | Success |
| `1` | Legacy or untyped error |
| `2` | Input or flag validation failure |
| `3` | License validation/API failure |
| `4` | Docker preflight failure |
| `5` | Template render or file write failure |

## Non-Functional Requirements

| Area | Requirement |
|---|---|
| Security | Never print full license keys or generated service secrets in errors. |
| File permissions | Generated kkengine and n8n `.env` files must be `0600`; config file is currently `0644` and should hold non-secret values only. |
| Compatibility | Runtime requires Docker and Docker Compose v2-compatible commands. |
| Release scope | Published GoReleaser artifacts currently target Linux `amd64` and `arm64`. |
| npm scope | `@kkauto/kkcli` is a Linux-only wrapper for release binaries; npm publish requires scope ownership and registry auth/trusted publishing. |
| Maintainability | Commands should stay thin and delegate domain behavior to `pkg/*`. |
| Testability | CI and local development should keep `go test ./...` green. |

## Completed Roadmap Items

| Item | Notes |
|---|---|
| Secret-safe unattended license input | `--license-file` and `--license-stdin` avoid argv exposure. |
| Deterministic unattended init exits | Typed `ExitError` codes are mapped in `cmd/root.go`. |
| n8n command group | `install`, lifecycle, logs, update, remove, and status commands exist. |
| npm distribution wrapper | `npm/kkcli` exposes `kk` and verifies GitHub Release SHA256 before extraction. |

## Success Metrics

| Metric | Target |
|---|---|
| Unattended provisioning | Valid `kk init --yes ...` runs without prompts. |
| Test suite | `go test ./...` passes before release handoff. |
| Secret exposure | No full license or generated secret appears in CLI errors/docs examples. |
| Documentation | README and evergreen docs describe only verified commands and flags. |

## References

- [Codebase Summary](./codebase-summary.md)
- [System Architecture](./system-architecture.md)
- [Code Standards](./code-standards.md)
- [Deployment Guide](./deployment-guide.md)
- [Project Roadmap](./project-roadmap.md)

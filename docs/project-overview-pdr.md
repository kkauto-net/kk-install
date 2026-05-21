# Project Overview and PDR

## Project Overview

`kkcli` is a command-line installer and operator for kkengine Docker Compose deployments. It installs as `kk` and helps users initialize configuration, start services, check status, update images, and manage optional n8n automation components.

## Product Goals

| Goal | Description |
|---|---|
| Fast stack setup | Generate all required Compose/config files from embedded templates |
| Safe operations | Run Docker and Compose preflight checks before starting services |
| Automation support | Allow backend provisioning systems to run unattended installs over SSH |
| Secure defaults | Generate strong secrets and restrict `.env` permissions |
| Operator clarity | Present status, health, and actionable error suggestions in the terminal |

## Primary Users

| User | Need |
|---|---|
| Individual operator | Install and manage kkengine locally or on a VPS |
| Backend provisioning worker | Run non-interactive setup for paid VPS service delivery |
| Maintainer | Extend CLI commands, templates, and validations safely |

## Functional Requirements

| ID | Requirement | Current Status |
|---|---|---|
| PDR-001 | CLI must initialize a kkengine Docker Compose stack with `kk init` | Implemented |
| PDR-002 | CLI must start services with preflight checks through `kk start` | Implemented |
| PDR-003 | CLI must show container status through `kk status` | Implemented |
| PDR-004 | CLI must update images and recreate containers through `kk update` | Implemented |
| PDR-005 | CLI must validate license keys against the kk license API during init | Implemented |
| PDR-006 | CLI must support English and Vietnamese language selection | Implemented |
| PDR-007 | CLI must generate strong secrets for rendered configuration | Implemented |
| PDR-008 | CLI must support unattended VPS provisioning with `kk init --yes --license <key> --domain <domain> --language <en|vi>` | Current uncommitted issue #3 implementation |
| PDR-009 | Unattended init must avoid interactive prompts when required flags are valid | Current uncommitted issue #3 implementation |
| PDR-010 | Unattended init must return deterministic exit codes for automation | Current uncommitted issue #3 implementation |

## Non-Functional Requirements

| Area | Requirement |
|---|---|
| Security | Do not expose full license keys or generated secrets in errors/logs |
| File permissions | Generated `.env` and `.env` backups must use owner-only permissions (`0600`) |
| Compatibility | Docker Compose v2+ is required for stack operation |
| Reliability | Existing untyped command errors retain legacy exit code `1` |
| Maintainability | CLI behavior must be covered by focused Go tests where feasible |

## Issue #3 Acceptance Criteria

Verified against current uncommitted code and tests:

| Criterion | Evidence |
|---|---|
| `kk init --yes --license ... --domain ... --language ...` exists | Flags registered in `cmd/init.go` |
| Missing unattended flags fail before prompts | `validateInitOptions` in `cmd/init_options.go` |
| Invalid license format fails before API call | `pkg/license.ValidateFormat` used by `validateInitOptions` |
| Invalid domain/language fail deterministically | `validateInitOptions` tests in `cmd/init_test.go` |
| License API failures return exit code `3` | `NewExitError(exitCodeLicenseValidation, ...)` in `cmd/init.go` |
| Docker validation failures return exit code `4` unless `--force` bypasses | Non-interactive Docker branches in `cmd/init.go` |
| Render/write failures return exit code `5` | `templates.RenderAll` error wrapping in `cmd/init.go` |
| `.env` remains private | `pkg/templates.RenderAll` chmods `.env`; `backupExistingConfigs` uses `0600` for `.env` backups |
| README contains unattended install example | `README.md` current uncommitted update |

## Constraints

- The license API endpoint and response handling are owned by `pkg/license` and should not be duplicated in command code.
- Template output is owned by `pkg/templates`; command code should pass `templates.Config` rather than writing stack files directly.
- Unattended mode must not add service-disable flags unless a separate requirement introduces them.
- Documentation examples must use fake license keys only.

## Success Metrics

| Metric | Target |
|---|---|
| Automated install command | Runs without TTY prompts when valid flags are supplied |
| Test suite | `go test ./...` passes |
| Secret exposure | No full license or generated secret in validation errors |
| Operator docs | README and docs describe unattended flow and exit codes consistently |

## References

- [Codebase Summary](./codebase-summary.md)
- [System Architecture](./system-architecture.md)
- [Code Standards](./code-standards.md)

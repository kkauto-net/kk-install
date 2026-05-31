# Codebase Summary

Generated from `./repomix-output.xml` on 2026-05-31 with:

```bash
repomix --style xml -o repomix-output.xml
```

Repomix packed 151 files and reported no suspicious files. This summary was then checked against current source files.

## Overview

`kkcli` is a Go CLI module at `github.com/kkauto-net/kk-install`. The distributed binary is `kk`. The CLI provisions and operates kkengine Docker Compose stacks and includes a separate command group for n8n workflow automation.

## Runtime Stack

| Area | Implementation |
|---|---|
| Language | Go `1.24.2` |
| CLI framework | Cobra |
| Interactive prompts | Charmbracelet `huh` |
| Terminal UI | `pterm` and internal `pkg/ui` helpers |
| Docker access | Docker CLI/Compose wrappers plus Docker SDK use in monitor/validator packages |
| Config formats | YAML and TOML |
| Tests | Go tests with `testify` and `go-cmp` |

## Repository Map

| Path | Purpose |
|---|---|
| `main.go` | Process entry point; calls `cmd.Execute()`. |
| `cmd/` | Cobra command definitions, flags, orchestration, exit-code mapping. |
| `pkg/config/` | User config under `~/.kk/config.yaml` and project directory helpers. |
| `pkg/license/` | License format validation and remote license API client. |
| `pkg/templates/` | Embedded kkengine templates and render/write logic. |
| `pkg/compose/` | Docker Compose command execution and Compose YAML parsing. |
| `pkg/validator/` | Docker, Compose, ports, env, config, disk, and preflight validation. |
| `pkg/monitor/` | Container status and Docker health monitoring. |
| `pkg/ui/` | i18n messages, banners, progress, tables, errors, password generation. |
| `pkg/updater/` | Docker image identity snapshot/diff logic, running-container comparison, and legacy pull output parsing. |
| `pkg/selfupdate/` | GitHub release lookup, archive download, binary replacement. |
| `pkg/n8n/` | n8n directories, config validation, and templates. |
| `scripts/` | Installer script and local installer checksum test harness. |
| `npm/kkcli/` | Linux-only npm wrapper that downloads verified GitHub Release artifacts and exposes `kk`. |
| `.github/workflows/` | CI, reviewdog, template validation, e2e compose, auto-version, release, draft-release, cleanup-artifacts. |

## Verified Command Surface

| Command | Verified flags/subcommands |
|---|---|
| `kk init` | `--force/-f`, `--yes`, `--license`, `--license-file`, `--license-stdin`, `--domain`, `--language` |
| `kk start` | Starts configured kkengine stack after preflight. |
| `kk stop` | Stops configured kkengine stack. |
| `kk restart` | Restarts configured kkengine stack. |
| `kk remove` | `--volumes/-v` also removes data volumes. |
| `kk status` | Shows container status. |
| `kk update` | Pulls images, compares image identities, optionally force-recreates containers; `--force/-f` skips confirmation. |
| `kk selfupdate` | `--check/-c`, `--force/-f` |
| `kk config show` | Shows language, project dir, config path. |
| `kk completion` | `bash`, `zsh`, `fish` |
| `kk n8n install` | `--force/-f` |
| `kk n8n logs` | `--follow/-f`, `--tail/-n`, `--all/-a` |
| `kk n8n remove` | `--volumes/-v` |
| `kk n8n update` | `--force/-f` |
| `kk n8n start/stop/restart/status` | n8n lifecycle/status commands. |

## Generated Files

| Stack | Files |
|---|---|
| kkengine | `docker-compose.yml`, `.env`, `kkphp.conf`, optional `Caddyfile`, optional `kkfiler.toml` |
| n8n | `docker-compose.yml`, `.env` under `pkg/n8n.N8nDir()` |

`pkg/templates.RenderAll` and `pkg/n8n.RenderAll` chmod generated `.env` files to `0600`.

Generated kkengine Compose includes `/etc/machine-id:/etc/machine-id:ro` so the host runtime can derive v2 license hardware identity. The template intentionally does not generate `LICENSE_STATE_DIR`, a dedicated license-state bind mount, `/sys/class/dmi/id`, or offline-token key env vars by default.

## Security Snapshot

| Area | Current behavior |
|---|---|
| License format | `LICENSE-[A-F0-9]{16}` before API call. |
| License API | POST to `https://kkauto.net/api/license/config`. |
| License input | `--license-file` and `--license-stdin` avoid argv exposure; input cap is 4096 bytes. |
| Secrets | Generated with `crypto/rand` in `cmd/init.go` and `pkg/ui/passwords.go`. |
| Config file | `~/.kk/config.yaml` is written `0644` and currently stores non-secret project/language data. |
| Installer checksum | `scripts/install.sh` requires `checksums.txt` SHA256 verification before installing the binary. |
| Self-update integrity | `pkg/selfupdate` requires matching release `checksums.txt` SHA256 verification before extracting and replacing the binary. |
| npm wrapper integrity | `npm/kkcli` verifies the downloaded release archive SHA256 by exact filename before extraction. |
| License host identity | Generated Compose mounts `/etc/machine-id` read-only. This is stable identity input, not a secret; backend heartbeat/offline-token policy enforces runtime access. |

## CI and Release Snapshot

| Workflow/tool | Verified behavior |
|---|---|
| `make test` | `go test -v ./...` |
| `make test-smoke` | Builds `kk` and verifies Docker-free command wiring. |
| `make build` | `CGO_ENABLED=0 go build` to `build/kk` |
| CI | Tests `./...`, builds `kk`, runs binary smoke, runs golangci-lint on push and PR, and runs race/shuffle outside PRs. |
| Installer shell tests | `scripts/install_test.sh` runs 7 offline tests for checksum branches, no-checksum-tool failure, and piped-installer guard behavior in CI. |
| npm wrapper tests | `npm/kkcli` runs offline Node tests and `npm pack --dry-run` in CI. |
| Scheduled security scan | Pinned `govulncheck` runs only on scheduled CI as a staged vulnerability check and reports findings as warnings. |
| Reviewdog | Runs golangci-lint and shellcheck on PRs to `main`. |
| Template validation | Uses Go from `go.mod`, checks template content, runs template tests, validates golden YAML. |
| Release | GoReleaser publishes Linux `amd64`/`arm64` tarballs and checksums. |
| npm publish | `@kkauto/kkcli` is public on npm. `release.yml` uses npm Trusted Publisher and calls `publish-npm.yml` after matching release assets are available; release-triggered publish is enabled by `NPM_PUBLISH_ENABLED=true`. |
| E2E Compose | Nightly/manual workflow runs unattended init, Compose config, full lifecycle, cleanup, and fail-closed redacted diagnostics. |

## Known Inconsistencies to Track

- Draft-release and release integrity docs should stay synced with checksum asset naming if GoReleaser config changes.
- Release integrity currently uses SHA256 checksums only; no release signature verification is implemented.

## Related Docs

- [Project Overview and PDR](./project-overview-pdr.md)
- [System Architecture](./system-architecture.md)
- [Code Standards](./code-standards.md)
- [Deployment Guide](./deployment-guide.md)
- [Project Roadmap](./project-roadmap.md)

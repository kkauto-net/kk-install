# System Architecture

## Overview

`kkcli` is a single-binary Go CLI. `main.go` calls the Cobra root command in `cmd/`; command handlers orchestrate domain packages in `pkg/`; domain packages perform file, HTTP, Docker, template, config, and terminal UI work.

```text
main.go
  -> cmd.Execute()
    -> Cobra command in cmd/
      -> package in pkg/
        -> Docker / HTTP / filesystem / template / terminal operation
```

## Component Responsibilities

| Component | Responsibility |
|---|---|
| `cmd` | Command tree, flags, prompt flow, typed exit errors, command orchestration. |
| `pkg/config` | User config load/save and project directory checks. |
| `pkg/license` | License regex validation and kk license API calls. |
| `pkg/templates` | kkengine template rendering and `.env` permissions. |
| `pkg/compose` | Docker Compose execution and YAML parsing. |
| `pkg/validator` | Docker/Compose/preflight/ports/env/config/disk validation. |
| `pkg/monitor` | Docker health and service status. |
| `pkg/ui` | i18n, progress, tables, banners, suggestions, password generation. |
| `pkg/updater` | Parse legacy Docker pull output and compare Docker image identities for update detection. |
| `pkg/selfupdate` | GitHub release check, tarball download/extract, binary replacement. |
| `pkg/n8n` | n8n stack paths, config validation, template rendering. |

## kkengine Init Flow

### Interactive

```text
kk init
  -> collect current working directory
  -> load existing .env values when present
  -> prompt license/language/services/domain/timezone/secrets
  -> validate license through pkg/license
  -> check Docker installation, daemon, and Compose
  -> build templates.Config
  -> backup existing generated files when needed
  -> pkg/templates.RenderAll(targetDir)
  -> save ~/.kk/config.yaml
```

### Unattended

```text
kk init --yes --license-file <path> --domain <domain> --language <en|vi>
  -> collectInitOptions()
  -> resolveInitLicenseSource()
  -> validateInitOptions()
  -> validate license through pkg/license
  -> run Docker checks without prompts unless --force bypasses
  -> generate defaults and secrets
  -> pkg/templates.RenderAll(targetDir)
  -> save ~/.kk/config.yaml
```

`--license-file` is the recommended automation source. `--license-stdin` is supported. `--license` exists but should not be used in provisioning scripts because argv can leak.

## Stack Architecture

### kkengine Stack

| Service | Template evidence |
|---|---|
| `kkengine` | `kkauto/kkengine:latest`, published on `8019:8019`. |
| `db` | MariaDB `10.6`, currently published as `3306:3306`. |
| `redis` | Redis Alpine with password from `.env`. |
| `seaweedfs` | Optional object/file storage service. |
| `caddy` | Optional reverse proxy on ports `80` and `443`. |

Generated files: `docker-compose.yml`, `.env`, `kkphp.conf`, optional `Caddyfile`, optional `kkfiler.toml`.

The generated kkengine Compose template mounts `/etc/machine-id:/etc/machine-id:ro` by default. The host runtime hashes this host-level identifier as part of v2 license hardware identity. The mount is read-only and is not a secret; backend heartbeat leases and offline-token expiry remain the enforcement boundary. The installer does not generate `LICENSE_STATE_DIR`, a separate license-state bind mount, or offline-token key environment variables.

### n8n Stack

`kk n8n install` renders n8n `docker-compose.yml` and `.env` through `pkg/n8n.RenderAll`. The n8n directory is `ProjectDir/n8n` when `~/.kk/config.yaml` has `ProjectDir`; otherwise it falls back to `~/.kk/n8n` and then `/tmp/.kk/n8n` for edge cases.

## Operation Flows

### Start

```text
kk start
  -> config.EnsureProjectDir()
  -> compose.ParseComposeFile()
  -> validator.RunPreflight()
  -> compose.Executor.Up()
  -> monitor health/status
```

### Update

```text
kk update [-f]
  -> config.EnsureProjectDir()
  -> compose.ParseComposeFile()
  -> updater.SnapshotImages() before pull
  -> compose.Executor.Pull()
  -> updater.SnapshotImages() after pull
  -> updater.CompareSnapshots()
  -> updater.CompareRunningContainers()
  -> optional confirmation
  -> compose.Executor.ForceRecreate()
  -> monitor health/status
```

Image update detection uses repo digest when Docker exposes it, otherwise image ID. It also compares running container image IDs with the desired local image IDs so a prior pull-without-recreate is still detected as pending work. Pulling an image only updates the local image cache; `ForceRecreate` is the apply step that recreates containers so the running services actually use the pulled image.

### Self-update

```text
kk selfupdate [--check] [--force]
  -> GitHub latest release API
  -> pick asset kkcli_<version>_<goos>_<goarch>.tar.gz
  -> pick checksums.txt from the same release
  -> download archive
  -> download checksums.txt
  -> verify archive SHA256 by exact artifact filename
  -> extract kk binary
  -> replace current executable, with sudo when needed
```

Security note: installer and self-update paths require successful SHA256 verification before installing or replacing the `kk` binary.

## Test And CI Architecture

| Gate | Scope |
|---|---|
| PR and main CI | `go test -v ./...`, static lint, build, and Docker-free binary smoke. |
| Installer shell tests | Source `scripts/install.sh` safely and run 7 offline tests for checksum success/failure, missing checksum tooling, and piped execution. |
| Main and scheduled CI | Shuffled tests plus selected race tests for command, license, template, compose, and validator packages. |
| Scheduled security scan | Pinned `govulncheck` runs outside PR gates as a low-noise vulnerability scan trial and reports findings as warnings. |
| Release and draft release | Full repository tests before release build steps. |
| Template validation | Uses Go from `go.mod` and validates template tests/golden YAML. |
| Nightly/manual e2e | Builds `kk`, runs unattended init, validates Compose config, runs start/status/stop/remove, collects redacted `compose ps`/log diagnostics, deletes compose diagnostics if redaction fails, and cleans up resources. |

Real Docker Compose lifecycle tests are kept out of PR gates to avoid network, image-pull, license, and runner flake blocking deterministic changes.

## Data and Configuration

| File/location | Owner | Notes |
|---|---|---|
| `~/.kk/config.yaml` | `pkg/config` | Written `0644`; stores language/project directory. |
| kkengine `.env` | `pkg/templates` | Written and chmodded `0600`. |
| n8n `.env` | `pkg/n8n` | Written and chmodded `0600`. |
| `repomix-output.xml` | Documentation workflow | Generated codebase compaction, not runtime input. |

## License Runtime Template Policy

| Item | Behavior |
|---|---|
| Host identity mount | `/etc/machine-id:/etc/machine-id:ro` in `pkg/templates/docker-compose.yml.tmpl`. |
| Optional DMI input | Not generated by default; may be added manually by support when needed. |
| Offline state env | `LICENSE_STATE_DIR` is not generated; host runtime uses its default path. |
| Offline token keys | No `LICENSE_OFFLINE_TOKEN_*` env vars; verification reuses `SERVER_PUBLIC_KEY_ENCRYPTED`. |
| Anti-abuse boundary | Backend binding, heartbeat lease, and offline-token expiry; not machine-id secrecy. |

## Error Model

`cmd/root.go` maps command errors through `ExitCode(err)`. Untyped errors exit `1`; typed init errors provide deterministic automation codes `2` through `5`.

## Known Architecture Risks

- Release artifacts are Linux-only while source may build elsewhere.
- Release artifact integrity depends on GitHub release `checksums.txt`; missing or invalid checksum metadata fails closed.

## References

- [Project Overview and PDR](./project-overview-pdr.md)
- [Code Standards](./code-standards.md)
- [Codebase Summary](./codebase-summary.md)
- [Deployment Guide](./deployment-guide.md)

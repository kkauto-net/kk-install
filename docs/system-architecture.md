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
| `pkg/updater` | Parse Docker pull output into image update information. |
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
  -> compose.Executor.Pull()
  -> updater.ParsePullOutput()
  -> optional confirmation
  -> compose.Executor.ForceRecreate()
  -> monitor health/status
```

### Self-update

```text
kk selfupdate [--check] [--force]
  -> GitHub latest release API
  -> pick asset kkcli_<version>_<goos>_<goarch>.tar.gz
  -> download archive
  -> extract kk binary
  -> replace current executable, with sudo when needed
```

Security note: installer script verifies checksums when available, but `pkg/selfupdate` has no visible checksum/signature verification.

## Test And CI Architecture

| Gate | Scope |
|---|---|
| PR and main CI | `go test -v ./...`, static lint, build, and Docker-free binary smoke. |
| Main and scheduled CI | Shuffled tests plus selected race tests for command, license, template, compose, and validator packages. |
| Release and draft release | Full repository tests before release build steps. |
| Template validation | Uses Go from `go.mod` and validates template tests/golden YAML. |
| Nightly/manual e2e | Builds `kk`, runs unattended init, validates Compose config, runs start/status/stop/remove, collects redacted diagnostics, and cleans up resources. |

Real Docker Compose lifecycle tests are kept out of PR gates to avoid network, image-pull, license, and runner flake blocking deterministic changes.

## Data and Configuration

| File/location | Owner | Notes |
|---|---|---|
| `~/.kk/config.yaml` | `pkg/config` | Written `0644`; stores language/project directory. |
| kkengine `.env` | `pkg/templates` | Written and chmodded `0600`. |
| n8n `.env` | `pkg/n8n` | Written and chmodded `0600`. |
| `repomix-output.xml` | Documentation workflow | Generated codebase compaction, not runtime input. |

## Error Model

`cmd/root.go` maps command errors through `ExitCode(err)`. Untyped errors exit `1`; typed init errors provide deterministic automation codes `2` through `5`.

## Known Architecture Risks

- Release artifacts are Linux-only while source may build elsewhere.
- Self-update lacks visible checksum/signature verification.

## References

- [Project Overview and PDR](./project-overview-pdr.md)
- [Code Standards](./code-standards.md)
- [Codebase Summary](./codebase-summary.md)
- [Deployment Guide](./deployment-guide.md)

# System Architecture

## Overview

`kkcli` is a single-binary Go CLI. The root command delegates to Cobra subcommands in `cmd/`, which orchestrate domain packages under `pkg/`. The CLI renders Docker Compose stack files, validates Docker readiness, starts containers, and reports health/status.

## High-Level Flow

```text
main.go
  -> cmd.Execute()
    -> Cobra command
      -> pkg/* domain package
        -> Docker/HTTP/filesystem/template operation
```

## Package Responsibilities

| Package | Role |
|---|---|
| `cmd` | CLI command orchestration, flag parsing, prompt flow, exit-code mapping |
| `pkg/license` | License key format validation and kk license API calls |
| `pkg/templates` | Embedded template rendering and secret length validation |
| `pkg/validator` | Docker, Compose, file, port, and preflight validations |
| `pkg/compose` | Docker Compose command execution and Compose file parsing |
| `pkg/monitor` | Container status and health checks |
| `pkg/ui` | Terminal output, messages, colors, tables, and suggestions |
| `pkg/config` | Project directory/config helpers |
| `pkg/updater` | Docker pull output parsing |
| `pkg/n8n` | n8n-specific template/config support |

## `kk init` Architecture

### Interactive Mode

```text
kk init
  -> load existing .env values
  -> prompt for license with huh
  -> validate license through pkg/license
  -> prompt or assist Docker setup
  -> prompt language/services/domain/timezone/secrets
  -> build templates.Config
  -> pkg/templates.RenderAll
  -> save local project config
```

### Unattended Mode

Unattended init supports non-interactive provisioning with a secret-safe license source:

```text
kk init --yes --license-file <path> --domain <domain> --language <en|vi>
  -> collectInitOptions()
  -> resolveInitLicenseSource()
  -> validateInitOptions()
  -> validate license through pkg/license
  -> run Docker checks without prompts
  -> use default service selection and generated secrets
  -> build templates.Config
  -> pkg/templates.RenderAll
  -> save local project config
```

`--license-file` is the recommended automation path because license content does not appear in process arguments. `--license-stdin` is available for explicit stdin input. Legacy `--license <key>` remains supported for compatibility, but should not be used in provisioning scripts.

Provisioning scripts should create temporary license files with owner-only permissions (`0600`) and register a cleanup trap before running `kk init` so failed runs do not leave license material on disk.

`--force` can be combined with `--yes` to preserve existing Docker preflight bypass behavior.

## License Validation

`pkg/license` owns license validation:

| Step | Code |
|---|---|
| Format check | `ValidateFormat` expects `LICENSE-[A-F0-9]{16}` |
| API call | `LicenseClient.Validate` posts to `/api/license/config` on `https://kkauto.net` |
| Success condition | API response `status` must equal `success` |

Command code must not expose raw license values in error messages. License source errors name the source (`--license-file` or `--license-stdin`) instead of the value.

## Template Rendering

`pkg/templates.RenderAll` validates secret lengths, renders embedded templates, and applies `.env` permissions.

| File | Rendered When |
|---|---|
| `docker-compose.yml` | Always |
| `.env` | Always |
| `kkphp.conf` | Always |
| `Caddyfile` | Caddy enabled |
| `kkfiler.toml` | SeaweedFS enabled |

## Docker Operation Flow

`kk start` verifies the project directory, parses the Compose file, runs preflight checks, calls `compose.Executor.Up`, monitors health, then prints status and access information.

```text
kk start
  -> config.EnsureProjectDir
  -> compose.ParseComposeFile
  -> validator.RunPreflight
  -> compose.Executor.Up
  -> monitor.NewHealthMonitor / MonitorAll
  -> monitor.GetStatusWithServices
```

## Error and Exit-Code Model

`cmd/root.go` exits with `ExitCode(err)`. Untyped errors continue to exit with `1`; typed `ExitError` values carry deterministic automation codes.

| Code | Source |
|---:|---|
| `1` | Legacy/untyped errors |
| `2` | `validateInitOptions` input failures |
| `3` | License API validation failures in `runInit` |
| `4` | Non-interactive Docker validation failures in `runInit` |
| `5` | Template render/write failures in `runInit` |

## Security Boundaries

- License validation is remote, but license format is checked locally before API calls.
- Unattended automation should pass license content through `--license-file` or explicit stdin, not argv.
- Temporary license files used for `--license-file` should be chmodded to `0600` and removed by a shell trap.
- Generated stack secrets are held in memory long enough to build `templates.Config` and render files.
- `.env` output is chmodded to `0600` by `pkg/templates.RenderAll`.
- `.env` backups are written with `0600` by `backupExistingConfigs`.

## Validation

Architecture and issue #3 behavior were reviewed against current files and `go test ./...` passed during this documentation update.

## References

- [Project Overview and PDR](./project-overview-pdr.md)
- [Code Standards](./code-standards.md)
- [Codebase Summary](./codebase-summary.md)

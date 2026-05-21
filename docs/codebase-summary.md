# Codebase Summary

Generated from `./repomix-output.xml` on 2026-05-22 and verified against current source files.

## Overview

`kkcli` is a Go CLI for provisioning and operating kkengine Docker Compose stacks. The binary is exposed as `kk` and is implemented with Cobra commands under `cmd/`.

## Runtime Stack

| Area | Implementation |
|---|---|
| Language | Go `1.24.2` (`go.mod`) |
| CLI framework | `github.com/spf13/cobra` |
| Interactive prompts | `github.com/charmbracelet/huh` |
| Terminal UI | `github.com/pterm/pterm` plus internal `pkg/ui` helpers |
| Compose operations | Shells out through `pkg/compose.Executor` |
| Templates | Embedded files in `pkg/templates/*.tmpl` |
| Tests | Go tests across `cmd` and `pkg/*` |

## Main Entry Points

| File | Purpose |
|---|---|
| `main.go` | Calls `cmd.Execute()` |
| `cmd/root.go` | Defines root `kk` command and maps command errors to process exit codes |
| `cmd/init.go` | Initializes stack config and renders Compose/template files |
| `cmd/init_options.go` | Collects and validates unattended `kk init` flags |
| `cmd/exit_error.go` | Typed exit-code wrapper for deterministic CLI failures |

## Command Surface

Core commands verified in `cmd/*.go`:

| Command | Purpose |
|---|---|
| `kk init` | Create Docker Compose stack config using interactive prompts, `--force`, or unattended `--yes` flags |
| `kk start` | Run preflight checks, start services, and report health/status |
| `kk stop` | Stop services |
| `kk restart` | Restart services |
| `kk remove` | Remove services; supports volume removal through command flags |
| `kk status` | Show container status |
| `kk update` | Pull images and recreate containers |
| `kk selfupdate` | Update the CLI binary |
| `kk completion` | Generate shell completion |
| `kk n8n ...` | Manage n8n workflow automation commands |

## Unattended VPS Install Impact

Current implementation supports true non-interactive initialization with a non-argv license source:

```bash
install -d -m 700 /root/.kk
license_file="$(mktemp /root/.kk/license.XXXXXX)"
cleanup_license_file() { rm -f "$license_file"; }
trap cleanup_license_file EXIT
printf '%s\n' "$KKAUTO_LICENSE" > "$license_file"
chmod 600 "$license_file"
kk init --yes --license-file "$license_file" --domain example.com --language en
```

Verified behavior from `cmd/init.go`, `cmd/init_options.go`, and tests:

- `--yes` enables unattended mode and skips `huh` prompt forms.
- Exactly one license source is required when `--yes` is set: recommended `--license-file`, explicit `--license-stdin`, or legacy `--license`.
- `--license` remains compatible but is discouraged for automation because argv can leak through process listings or logs.
- Automation examples should create the temporary license file with `0600` permissions and remove it with a shell trap.
- `--domain` and `--language` are required when `--yes` is set.
- `--language` accepts only `en` or `vi`.
- License format is validated before the API call by `pkg/license.ValidateFormat`.
- Domain validation reuses `validateDomain` in `cmd/init.go`.
- Existing config files are backed up before overwrite; `.env` backups use `0600` permissions.
- Template rendering still uses `pkg/templates.RenderAll`, which chmods generated `.env` to `0600`.
- `--force` keeps Docker preflight bypass behavior when used with unattended mode.

## Deterministic Exit Codes

Verified in `cmd/exit_error.go` and `cmd/root.go`:

| Code | Meaning |
|---:|---|
| `0` | Success |
| `1` | Legacy/untyped error |
| `2` | Input/flag validation failure |
| `3` | License validation/API failure |
| `4` | Docker validation failure |
| `5` | Template render or file write failure |

## Generated Stack Files

`pkg/templates.RenderAll` writes these files from embedded templates:

| Output | Condition |
|---|---|
| `docker-compose.yml` | Always |
| `.env` | Always; chmod `0600` |
| `kkphp.conf` | Always |
| `Caddyfile` | When Caddy is enabled |
| `kkfiler.toml` | When SeaweedFS is enabled |

## Validation Snapshot

Commands run during this documentation review:

```bash
repomix --style xml -o repomix-output.xml
```

## Related Docs

- [Project Overview and PDR](./project-overview-pdr.md)
- [System Architecture](./system-architecture.md)
- [Code Standards](./code-standards.md)

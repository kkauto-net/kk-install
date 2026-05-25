# Code Standards

## Scope

These standards describe the current Go CLI codebase. They are documentation-only guidance for maintainers; code changes must still be verified against implementation and tests.

## Repository Structure

| Path | Responsibility |
|---|---|
| `main.go` | Minimal process entry point. |
| `cmd/` | Cobra commands, flags, command orchestration, exit-code mapping. |
| `pkg/compose/` | Docker Compose execution and Compose YAML parsing. |
| `pkg/config/` | `~/.kk/config.yaml` and project directory helpers. |
| `pkg/license/` | License regex and kk license API client. |
| `pkg/monitor/` | Container status and health checks. |
| `pkg/n8n/` | n8n config, directories, embedded templates. |
| `pkg/selfupdate/` | GitHub release check, archive download, binary replacement. |
| `pkg/templates/` | kkengine embedded templates and render validation. |
| `pkg/ui/` | i18n, terminal UI, tables, progress, passwords. |
| `pkg/updater/` | Docker image identity snapshot/diff logic, running-container comparison, and legacy pull output parsing. |
| `pkg/validator/` | Docker, Compose, ports, env, config, disk, preflight checks. |
| `scripts/` | Installer and operational scripts. |
| `docs/` | Evergreen project documentation. |

## Go Standards

- Keep `main.go` thin; all behavior should flow through `cmd.Execute()`.
- Keep command files focused on CLI orchestration; move reusable behavior into `pkg/*`.
- Use `RunE` for commands that can fail.
- Return errors instead of calling `os.Exit` outside the root command boundary.
- Preserve existing command names and flags unless a migration requirement exists.
- Keep validation ownership clear: license validation in `pkg/license`, template validation in `pkg/templates`, Docker/preflight validation in `pkg/validator`.
- Prefer small helper functions for command option resolution, especially when automation behavior needs tests.

## CLI Standards

| Area | Standard |
|---|---|
| Flags | Register flags near command definitions in `init()`. |
| Automation | Keep non-interactive paths free of prompt calls. |
| Output | Use `pkg/ui` helpers for user-facing terminal messages where practical. |
| Errors | Use typed `ExitError` only where stable automation exits are part of the contract. |
| Backward compatibility | Keep legacy behavior unless explicitly replaced. |

## `kk init` Standards

- Interactive mode may use `huh` prompts.
- `--yes` is unattended mode and requires exactly one license source, `--domain`, and `--language`.
- Recommended automation license source is `--license-file`; `--license-stdin` is also non-argv; `--license` is compatibility only.
- `--language` accepts only `en` and `vi`.
- License format is `LICENSE-[A-F0-9]{16}`.
- File/stdin license input is capped at 4096 bytes.
- Existing generated config files should be backed up before overwrite where current init behavior does so.
- kkengine template writes must continue through `pkg/templates.RenderAll`.

## Exit-Code Standards

| Code | Use |
|---:|---|
| `0` | Success |
| `1` | Legacy or untyped command errors |
| `2` | Input/flag validation failures |
| `3` | License validation/API failures |
| `4` | Docker/Compose preflight failures |
| `5` | Template render or file write failures |

Do not convert every legacy error to a typed error without a product requirement; stable codes are part of the unattended init contract.

## Security Standards

- Never print full license keys in command errors.
- Mask licenses as `LICENSE-************6789` when display is unavoidable.
- Do not recommend argv license input for automation.
- Do not print generated values for `JWT_SECRET`, `DB_PASSWORD`, `DB_ROOT_PASSWORD`, `REDIS_PASSWORD`, `S3_ACCESS_KEY`, `S3_SECRET_KEY`, or n8n encryption keys in logs/errors.
- Generated kkengine and n8n `.env` files must remain `0600`.
- `~/.kk/config.yaml` is currently `0644`; keep it non-secret unless permissions and migration are redesigned.
- Installer checksum verification should remain enabled when release checksums are available.
- Treat self-update checksum/signature verification as a roadmap security item.

## Testing Standards

Run these local checks before release handoff:

```bash
make fmt
make lint
make test
make test-smoke
make build
```

`make test` runs `go test -v ./...`. `make test-smoke` builds the CLI and verifies root command wiring without a Docker daemon.

Run race and shuffle checks before promoting them to required PR gates:

```bash
go test -shuffle=on ./...
go test -race ./cmd ./pkg/license ./pkg/templates ./pkg/compose ./pkg/validator
```

Nightly/manual Docker Compose validation lives in `.github/workflows/e2e-compose.yml`. It is intentionally not a PR requirement because it depends on Docker runtime, image pulls, network, and `KKAUTO_E2E_LICENSE`.

Add focused tests for command option validation, exit-code mapping, render permissions, secret masking, port/template contracts, and parser behavior.

## Documentation Standards

- Public quick-start and command lists belong in `README.md`.
- Architecture, requirements, deployment, roadmap, and design guidance belong in `docs/`.
- Keep evergreen documentation filenames kebab-case.
- Verify command names, flags, paths, and function references before documenting them.
- Do not document `.env` values from real environments.
- Keep README under 300 lines and docs under the project line target.

## References

- [Project Overview and PDR](./project-overview-pdr.md)
- [System Architecture](./system-architecture.md)
- [Codebase Summary](./codebase-summary.md)
- [Deployment Guide](./deployment-guide.md)

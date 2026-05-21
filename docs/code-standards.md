# Code Standards

## Scope

These standards reflect the current Go CLI codebase and the issue #3 unattended init implementation.

## Repository Structure

| Path | Responsibility |
|---|---|
| `main.go` | Minimal process entry point |
| `cmd/` | Cobra command definitions and command-level orchestration |
| `pkg/compose/` | Docker Compose execution and parsing |
| `pkg/config/` | Local project configuration helpers |
| `pkg/license/` | License format and remote API validation |
| `pkg/monitor/` | Container status and health monitoring |
| `pkg/n8n/` | n8n template/config helpers |
| `pkg/templates/` | Embedded kkengine templates and render validation |
| `pkg/ui/` | Terminal UI, i18n messages, tables, and error formatting |
| `pkg/updater/` | Docker image update parsing |
| `pkg/validator/` | Docker, Compose, port, env, and preflight checks |
| `scripts/` | Install and utility scripts |
| `docs/` | Evergreen project documentation |
| `plans/` | Time-scoped implementation plans and reports |

## Go Code Guidelines

- Keep `main.go` thin; command behavior belongs in `cmd/`.
- Keep package APIs focused by domain (`license`, `templates`, `validator`, etc.).
- Prefer small helpers for command option collection/validation when command files grow large.
- Return errors instead of calling `os.Exit` outside the root command execution boundary.
- Avoid duplicating validation already owned by a package.

## CLI Command Standards

- Define commands with Cobra in `cmd/*.go`.
- Use `RunE` for commands that can fail.
- Register flags in each command's `init()` function.
- Preserve legacy behavior unless a requirement explicitly changes it.
- Keep interactive prompts behind explicit interactive branches so automation paths remain prompt-free.

## `kk init` Standards

Verified current behavior:

- Interactive mode uses `huh` forms for license, language, service selection, domain, timezone, and secret edits.
- `--force` bypasses selected prompts and Docker validation failures with defaults.
- `--yes` is unattended mode and requires:
  - `--license`
  - `--domain`
  - `--language`
- `--language` accepts only `en` and `vi`.
- Existing config files are backed up before overwrite when unattended or force mode overwrites them.
- Template rendering must continue through `pkg/templates.RenderAll`.

## Exit Error Standards

Use typed errors from `cmd/exit_error.go` for deterministic automation outcomes.

| Code | Use For |
|---:|---|
| `1` | Legacy/untyped command errors |
| `2` | Flag or user input validation failures |
| `3` | License validation/API failures |
| `4` | Docker/Compose preflight failures |
| `5` | Template render or file write failures |

Do not wrap every legacy error just to change exit behavior. Use typed errors where automation contracts require stable codes.

## Security Standards

- Never print full license keys in command errors.
- Mask license keys as `LICENSE-************6789` when display is unavoidable.
- Never print generated secret values such as `JWT_SECRET`, `DB_PASSWORD`, `DB_ROOT_PASSWORD`, `REDIS_PASSWORD`, `S3_ACCESS_KEY`, or `S3_SECRET_KEY`.
- Generated `.env` files must remain `0600`.
- `.env` backups must remain `0600`.
- Documentation examples must use fake license keys.

## Testing Standards

- Run the full suite before documentation or release handoff:

```bash
go test ./...
```

- Add focused tests for command option validation, secret masking, file permissions, and exit-code mapping.
- Prefer tests against helpers (`validateInitOptions`, `ExitCode`, render helpers) over brittle terminal prompt tests.

## Documentation Standards

- Public usage changes belong in `README.md`.
- Architecture, product requirements, and implementation standards belong in `docs/`.
- Keep docs evidence-based: verify flags, function names, file paths, and exit codes in code before documenting.
- Keep generated docs under 800 lines per file; split by topic when needed.

## References

- [Project Overview and PDR](./project-overview-pdr.md)
- [System Architecture](./system-architecture.md)
- [Codebase Summary](./codebase-summary.md)

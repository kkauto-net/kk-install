# Design Guidelines

## Overview

`kkcli` design is terminal-first: short commands, safe defaults, clear prompts, and concise status output. The current implementation uses Cobra for command structure, `huh` for interactive forms, `pterm` for terminal presentation, and `pkg/ui` for shared messages and styling.

## CLI Experience Principles

| Principle | Guidance |
|---|---|
| Safety first | Make destructive actions explicit, especially volume removal and binary replacement. |
| Automation safe | Provide non-interactive flows that avoid prompts and secrets in argv. |
| Clear recovery | Pair failures with suggestions and commands where possible. |
| Consistency | Use the same flag meanings across kkengine and n8n commands. |
| Bilingual support | Keep English and Vietnamese messages aligned in `pkg/ui`. |

## Command Design

- Prefer verbs users already know: `init`, `start`, `stop`, `restart`, `status`, `update`, `remove`.
- Use `--force/-f` only for skipping confirmations or bypassing preflight behavior already supported by the command.
- Use `--volumes/-v` only for data-volume deletion semantics.
- Keep n8n commands under `kk n8n` to avoid mixing stack domains.
- Keep `kk config show` read-only unless new config mutation commands are explicitly designed.

## Prompt Design

- Interactive prompts may use `huh`.
- Do not call prompt forms from unattended paths.
- Confirm destructive actions, especially `remove -v`.
- Do not display full secret values after generation except where current n8n behavior explicitly requires the user to preserve the encryption key.
- Prefer concise labels and one action per prompt.

## Output Design

| Output type | Guideline |
|---|---|
| Banners | Use for major command starts only. |
| Steps | Use numbered steps for multi-stage operations like init/update/selfupdate. |
| Tables | Use for status, config summaries, and image updates. |
| Errors | Include title, safe message, suggestion, and command when available. |
| Secrets | Mask or omit; never print full licenses in errors. |

## Language and Terminology

| Term | Use |
|---|---|
| `kkcli` | Project/package/tool name in docs. |
| `kk` | Installed binary and command prefix. |
| `kkengine stack` | Main Docker Compose deployment. |
| `n8n stack` | Optional workflow automation deployment. |
| `unattended init` | `kk init --yes ...` automation mode. |

## Security UX

- Recommend `--license-file` in automation docs.
- Treat `--license` as compatibility-only for scripts.
- Explain temporary license file cleanup with `trap` in examples.
- Warn before volume deletion.
- Keep generated `.env` file permission behavior visible in docs.
- Describe installer and self-update integrity as fail-closed SHA256 verification against `checksums.txt`.
- Do not imply release signature verification; no GPG, cosign, or equivalent signature check is implemented.

## Documentation Design

- README stays short: install, quick start, unattended mode, command tables, docs links.
- Evergreen docs live in `docs/` with kebab-case names.
- Keep docs evidence-based and update them when command flags or workflows change.
- Cross-link related docs instead of duplicating long explanations.

## References

- [Code Standards](./code-standards.md)
- [Project Overview and PDR](./project-overview-pdr.md)
- [System Architecture](./system-architecture.md)

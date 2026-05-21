---
title: "Issue 3 unattended init completion"
date: 2026-05-21
issue: 3
type: journal
---

# Issue 3 Unattended Init Completion

## Summary

Implemented true non-interactive `kk init` for backend VPS provisioning.

Superseded by issue #4 for license input guidance: current automation docs recommend `--license-file` with `0600` temporary files and trap-based cleanup. Legacy `--license <key>` remains compatibility only and is discouraged for provisioning scripts.

## Completed Work

- Added initial unattended `kk init --yes --license <key> --domain <domain> --language <en|vi>` support. This argv license form is now legacy compatibility after issue #4.
- Added deterministic exit-code mapping for unattended validation, license, Docker, and render failures.
- Preserved interactive behavior while skipping all prompts in unattended mode.
- Added license masking for validation errors.
- Secured `.env` backups with owner-only file mode.
- Added command option, sanitizer, backup permission, and exit-code tests.
- Updated README and evergreen project docs.

## Validation

- `go test ./cmd` passed.
- `go test ./pkg/license ./pkg/templates` passed.
- `go test ./...` passed.
- Smoke-tested `kk init --yes` missing flags exits `2`.
- Smoke-tested skipped-license unattended init renders expected files without prompting.

## Unresolved Questions

None blocking.

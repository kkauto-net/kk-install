---
title: Secret-safe license input implementation
date: 2026-05-22
type: implementation-journal
plan: plans/260521-2232-secret-safe-license-input/plan.md
issue: 4
---

# Secret-safe License Input Implementation

## Context

Implemented GitHub issue #4 for unattended `kk init`: backend provisioning can now avoid passing KKAuto license keys through process arguments.

## What Happened

- Added `--license-file` and `--license-stdin` for unattended init.
- Kept legacy `--license` compatible, but docs now discourage it for automation.
- Added source resolution before option validation.
- Enforced exactly one license source in unattended mode.
- Added source trimming, empty-source rejection, non-regular file rejection, and a 4096-byte source-size guard.
- Added fail-fast behavior for `--license-stdin` when stdin is a TTY.
- Added user-visible diagnostics for early input validation failures while preserving deterministic exit code `2`.
- Updated README and evergreen docs to recommend owner-only temp license files with cleanup traps.
- Marked related plan and phase files completed.

## Verification

- `go test ./cmd` passed.
- `go test ./...` passed.
- Code review agent found no remaining critical/high findings after fixes.
- Tester and debugger agents passed CLI smoke and failure-mode checks.

## Decisions

- `--license-file` is the recommended automation path.
- `--license-stdin` is supported only when explicitly requested and piped/redirected.
- `--license-file` must point to a regular file to avoid FIFO/device blocking risk.
- Missing, multiple, unreadable, empty, non-regular, oversized, or invalid-format sources return exit code `2`.
- Remote license API rejection still returns exit code `3`.

## Next

- Ask whether to commit with `feat(init): add secret-safe unattended license input`.

## Unresolved Questions

None.

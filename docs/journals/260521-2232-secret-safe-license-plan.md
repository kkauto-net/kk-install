---
title: Secret-safe license input planning
date: 2026-05-21
type: planning-journal
plan: plans/260521-2232-secret-safe-license-input/plan.md
issue: 4
---

# Secret-safe License Input Planning

## Context

Created implementation plan for GitHub issue #4: add non-argv license input for unattended `kk init` so backend VPS provisioning does not expose KKAuto license keys through process arguments.

## What Happened

- Scanned unfinished plans under `plans/` before creating a new plan.
- Found one pending unrelated plan: `plans/260105-0930-reviewdog-pr-workflow/plan.md`.
- Confirmed completed issue #3 plan is a prerequisite but not an active blocker.
- Researched secret-safe CLI input patterns and selected `--license-file` as primary API.
- Kept legacy `--license` in scope for compatibility, but marked it as discouraged for automation.
- Added optional `--license-stdin` as a small extension if implementation remains simple.
- Wrote plan and phase files under `plans/260521-2232-secret-safe-license-input/`.

## Decisions

- Use `--license-file /path/to/license.tmp` as the required production-safe path, with the temporary file chmodded to `0600` and removed through a cleanup trap.
- Require exactly one license source in unattended mode.
- Classify missing, unreadable, empty, or invalid license-source errors as exit code `2`.
- Preserve remote license validation failures as exit code `3`.
- Add tests and docs to ensure no full license appears in output, docs examples, or failures.

## Next

- Implement via `/ck:cook /home/kkdev/kkinstall/plans/260521-2232-secret-safe-license-input/plan.md`.
- Start with phase 01 resolver changes, then phase 02 tests/smoke, then phase 03 docs handoff.

## Unresolved Questions

None blocking.

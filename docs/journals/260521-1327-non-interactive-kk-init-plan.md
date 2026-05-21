---
title: "Non-interactive kk init planning"
date: 2026-05-21
plan: "plans/260521-1327-non-interactive-kk-init/plan.md"
issue: 3
type: journal
---

# Non-interactive kk init Planning

## Context

Created an implementation plan for GitHub issue #3: true non-interactive `kk init` for backend VPS provisioning.

## What Happened

- Read project README and global development rules.
- Fetched issue #3 requirements from GitHub.
- Scanned existing plans for unfinished dependency overlap.
- Found one pending plan, `260105-0930-reviewdog-pr-workflow`, with no scope overlap.
- Analyzed current `cmd/init.go`, license validation, templates, root command exit behavior, and README docs targets.
- Created a self-contained plan with three phases under `plans/260521-1327-non-interactive-kk-init/`.

## Decisions

- Use `--yes` as the explicit unattended mode.
- Keep `--force` as legacy bypass/default behavior.
- Require `--license`, `--domain`, and `--language` when `--yes` is used.
- Reuse existing license API and template render flow.
- Add deterministic typed exit codes with fallback exit code `1` for legacy errors.
- Avoid new service-disable flags for Caddy or SeaweedFS in this issue.

## Next

Run cook against the plan:

```bash
/ck:cook /home/kkdev/kkinstall/plans/260521-1327-non-interactive-kk-init/plan.md
```

## Unresolved Questions

None blocking.

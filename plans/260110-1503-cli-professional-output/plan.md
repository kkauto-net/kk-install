---
title: "CLI Professional Output Improvements"
description: "Replace hand-rolled ASCII tables with pterm tables, add step wizard to init, improve i18n"
status: pending
priority: P2
effort: 3h
branch: main
tags: [ui, i18n, refactor]
created: 2026-01-10
---

# CLI Professional Output Improvements

## Overview

Improve kkcli output professionalism by:
- Using pterm tables instead of hand-rolled ASCII
- Adding step-by-step wizard UI to `kk init`
- Summary table at init completion
- Proper i18n for all UI strings
- Default English, proper Vietnamese diacritics

## Phases

| # | Phase | Status | Effort | Link |
|---|-------|--------|--------|------|
| 1 | UI Components | Done | 1h | [phase-01](./phase-01-ui-components.md) |
| 2 | i18n Messages | Done | 30m | [phase-02](./phase-02-i18n-messages.md) |
| 3 | Command Updates | Pending | 1h | [phase-03](./phase-03-command-updates.md) |
| 4 | Testing | Pending | 30m | [phase-04](./phase-04-testing.md) |

## Dependencies

- pterm library (already in project)
- Existing i18n infrastructure in `pkg/ui`

## Key Files

- `pkg/ui/table.go` - Table display functions
- `pkg/ui/progress.go` - New: step header, summary helpers
- `pkg/ui/lang_en.go` - English messages
- `pkg/ui/lang_vi.go` - Vietnamese messages
- `cmd/init.go` - Init command
- `cmd/status.go` - Status command
- `cmd/start.go` - Start command

## Success Criteria

- [ ] All tables use pterm with consistent styling
- [ ] `kk init` shows Step 1/5...Step 5/5 progress
- [ ] Summary table displayed at init completion
- [ ] Default language is English
- [ ] Vietnamese has proper diacritics
- [ ] Build passes with no lint errors

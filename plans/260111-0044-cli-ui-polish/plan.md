---
title: "CLI UI Polish - Boxed Tables & Error Improvements"
description: "Polish CLI output with consistent boxed tables, integrate ShowBoxedError, improve status icons"
status: pending
priority: P1
effort: 1.5h
branch: main
tags: [ui, pterm, cli, ux, polish]
created: 2026-01-11
---

# CLI UI Polish - Boxed Tables & Error Improvements

## Overview

Polish kkcli output với consistent boxed tables, integrate ShowBoxedError thay ShowError, cải thiện icons/colors cho status states.

## Context

- **Brainstorm Report:** `../reports/brainstormer-260111-0040-cli-status-error-ui.md`
- **Previous Work:** cli-professional-output-v2 (completed)
- **Library:** pterm (already in use)

## Requirements Summary

| ID | Requirement | Priority |
|----|-------------|----------|
| R1 | Box `PrintAccessInfo` table | P0 |
| R2 | Integrate `ShowBoxedError` trong commands | P0 |
| R3 | Add "starting" state icon (blue ◐) | P1 |
| R4 | Error grouping cho preflight checks | P2 |

## Implementation Phases

| Phase | Description | Status | Effort |
|-------|-------------|--------|--------|
| [Phase 01](./phase-01-quick-wins.md) | Box AccessInfo + Integrate ShowBoxedError | complete | 30m |
| [Phase 02](./phase-02-icons-colors.md) | Starting icon + Color refinements | complete | 30m |
| [Phase 03](./phase-03-error-grouping.md) | Preflight error grouping | pending | 30m |

## Files to Modify

### Phase 01
- `pkg/ui/table.go` - Box PrintAccessInfo
- `cmd/status.go` - Use ShowBoxedError
- `cmd/start.go` - Use ShowBoxedError
- `cmd/restart.go` - Use ShowBoxedError
- `cmd/update.go` - Use ShowBoxedError
- `cmd/init.go` - Use ShowBoxedError

### Phase 02
- `pkg/ui/table.go` - Add starting icon
- `pkg/ui/progress.go` - Update ShowServiceProgress với starting icon

### Phase 03
- `pkg/validator/preflight.go` - Error grouping
- `pkg/ui/errors.go` - ShowBoxedErrors (multiple)

## Success Criteria

1. All tables consistently boxed
2. All errors use ShowBoxedError with suggestions
3. Status icons: ● running, ○ stopped, ◐ starting
4. Preflight errors grouped in single box
5. Tests pass

## Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Breaking existing output format | Medium | Only style changes, structure same |
| Icon rendering on older terminals | Low | pterm fallback to ASCII |

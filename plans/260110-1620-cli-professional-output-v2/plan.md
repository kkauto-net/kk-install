---
title: "CLI Professional Output Enhancement v2"
description: "Upgrade kkcli command outputs with professional pterm UI - boxed tables, spinners, command banners"
status: pending
priority: P2
effort: 3h
branch: main
tags: [ui, pterm, cli, ux]
created: 2026-01-10
---

# CLI Professional Output Enhancement v2

## Overview

Enhance kkcli terminal output for professional, beginner-friendly appearance using pterm boxed tables, command banners, animated spinners, and structured error displays.

## Context

- **Brainstorm Report:** `../reports/brainstorm-260110-1620-cli-professional-output-v2.md`
- **Approach:** Solution A - Incremental Enhancement
- **Library:** pterm (already in use)

## Requirements Summary

| ID | Requirement |
|----|-------------|
| R1 | Boxed tables for all status/info displays |
| R2 | Verbose mode with step-by-step + summary |
| R3 | Professional animations (spinners, progress bars) |
| R4 | Standard color scheme |
| R5 | Boxed errors with fix suggestions |
| R6 | Default English, Vietnamese với dấu |

## Implementation Phases

| Phase | Description | Status | Effort |
|-------|-------------|--------|--------|
| [Phase 01](./phase-01-core-ui-components.md) | Core UI Components - banners, errors, spinners | DONE | 1h |
| [Phase 02](./phase-02-command-updates.md) | Update all commands with new UI | pending | 1.5h |
| [Phase 03](./phase-03-i18n-polish.md) | I18n updates and final polish | pending | 0.5h |

## Files to Modify

### New Files
- `pkg/ui/banner.go` - Command headers/footers
- `pkg/ui/errors.go` - Boxed errors with suggestions

### Modified Files
- `pkg/ui/table.go` - Add PrintUpdatesTable, box existing
- `pkg/ui/progress.go` - Replace SimpleSpinner with pterm
- `cmd/init.go`, `cmd/start.go`, `cmd/status.go`, `cmd/restart.go`, `cmd/update.go`
- `pkg/ui/lang_en.go`, `pkg/ui/lang_vi.go`

## Success Criteria

1. All commands show consistent header
2. All status/info uses boxed tables
3. All progress uses pterm spinners
4. Errors show suggestions for common issues
5. Output readable for CLI beginners
6. Tests pass

## Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| pterm version incompatibility | High | Pin version in go.mod |
| Terminal without color support | Medium | pterm handles gracefully |
| Breaking existing output parsing | Medium | Maintain structure, only style |

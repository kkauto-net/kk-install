---
title: "Professional CLI Output Format"
description: "Refactor kk CLI to use GitHub CLI style output with Vietnamese diacritics support"
status: pending
priority: P2
effort: 2h
branch: main
tags: [cli, ui, i18n, cobra]
created: 2026-01-10
---

# Professional CLI Output Format

## Overview

Refactor `kk` CLI output to match professional standards like GitHub CLI (`gh`). Key changes:
- GitHub CLI style help format with UPPERCASE headers and grouped commands
- Vietnamese messages with full diacritics
- Persistent language preference storage (`~/.kk/config.yaml`)
- Consistent messaging across all commands

## Problem Statement

Current issues:
1. Vietnamese without diacritics ("Khoi tao" → "Khởi tạo")
2. Flat command list, no logical grouping
3. No persistent language preference
4. Inconsistent messaging style

## Implementation Phases

| Phase | Description | Effort | Status |
|-------|-------------|--------|--------|
| [Phase 01](phase-01-config-storage.md) | Config storage for language preference | 30m | done |
| [Phase 02](phase-02-help-templates.md) | Custom Cobra help templates (GitHub CLI style) | 45m | done |
| [Phase 03](phase-03-language-files.md) | Update language files with diacritics | 30m | pending |
| [Phase 04](phase-04-command-annotations.md) | Add group annotations to commands | 15m | **REQUIRED** |

## Architecture

```
~/.kk/
└── config.yaml          # language: vi | en

pkg/
├── config/
│   └── config.go        # NEW: Config management
└── ui/
    ├── help.go          # NEW: Custom templates
    ├── lang_vi.go       # MODIFY: Add diacritics
    └── lang_en.go       # MODIFY: Polish messages

cmd/
├── root.go              # MODIFY: Apply templates, load config
├── init.go              # MODIFY: Add group annotation
├── start.go             # MODIFY: Add group annotation
├── status.go            # MODIFY: Add group annotation
├── restart.go           # MODIFY: Add group annotation
├── update.go            # MODIFY: Add group annotation
└── completion.go        # MODIFY: Add group annotation
```

## Command Grouping

| Group | Commands | Description |
|-------|----------|-------------|
| core | init, start, status | Primary workflows |
| management | restart, update | Operational commands |
| additional | completion | Utilities |

## Success Criteria

- [ ] `kk --help` shows GitHub CLI style output with grouped commands
- [ ] Vietnamese displays with full diacritics when selected
- [ ] Language preference persists in `~/.kk/config.yaml`
- [ ] All runtime messages use i18n system

## Related Files

- Brainstorm: `plans/reports/brainstorm-260110-1004-cli-professional-format.md`
- Codebase summary: `docs/codebase-summary.md`
- Code standards: `docs/code-standards.md`

## Risks

| Risk | Mitigation |
|------|------------|
| Template syntax errors | Unit tests for help output |
| Unicode issues in terminals | Test on common terminals |
| Config file permissions | Graceful error handling |

# Phase 01: Core UI Components

## Context
- **Parent Plan:** [plan.md](./plan.md)
- **Brainstorm:** [brainstorm report](../reports/brainstorm-260110-1620-cli-professional-output-v2.md)
- **Dependencies:** None (foundational phase)

## Overview

| Field | Value |
|-------|-------|
| Date | 2026-01-10 |
| Priority | P2 |
| Effort | 1h |
| Implementation Status | DONE |
| Review Status | DONE (9.5/10) |
| Review Report | [code-reviewer-260110-2305-phase01-ui-components.md](../reports/code-reviewer-260110-2305-phase01-ui-components.md) |

**Description:** Create core UI components that will be used across all commands - command banners, boxed error displays, and replace custom spinner with pterm.

## Key Insights

1. pterm already in use → no new dependencies
2. SimpleSpinner in progress.go is custom → replace with pterm.DefaultSpinner
3. Error messages scattered → centralize in errors.go
4. No command headers currently → add consistent banners

## Requirements

- R1: Boxed tables for all displays
- R3: Professional animations (pterm spinners)
- R5: Boxed errors with fix suggestions

## Architecture

```
pkg/ui/
├── banner.go       # NEW - ShowCommandBanner, ShowCompletionBanner
├── errors.go       # NEW - ShowBoxedError with suggestions
├── progress.go     # MODIFY - Replace SimpleSpinner with pterm
├── table.go        # MODIFY - Add PrintUpdatesTable
└── ...existing...
```

## Related Code Files

| File | Action | Description |
|------|--------|-------------|
| `pkg/ui/banner.go` | CREATE | Command header/footer functions |
| `pkg/ui/errors.go` | CREATE | Boxed error display with suggestions |
| `pkg/ui/progress.go` | MODIFY | Replace SimpleSpinner, add pterm spinner wrapper |
| `pkg/ui/table.go` | MODIFY | Add PrintUpdatesTable for update command |

## Implementation Steps

### 1. Create banner.go

```go
package ui

import "github.com/pterm/pterm"

// ShowCommandBanner displays command header box
func ShowCommandBanner(cmd, description string) {
    pterm.DefaultBox.
        WithTitle(pterm.Cyan(cmd)).
        WithTitleTopCenter().
        Println(description)
    pterm.Println() // spacing
}

// ShowCompletionBanner displays success/failure footer
func ShowCompletionBanner(success bool, title, content string) {
    style := pterm.NewStyle(pterm.FgGreen)
    if !success {
        style = pterm.NewStyle(pterm.FgRed)
    }
    pterm.DefaultBox.
        WithTitle(title).
        WithTitleTopCenter().
        WithBoxStyle(style).
        Println(content)
}
```

### 2. Create errors.go

```go
package ui

import "github.com/pterm/pterm"

// ErrorSuggestion contains error info and fix suggestion
type ErrorSuggestion struct {
    Title      string
    Message    string
    Suggestion string
    Command    string // optional command to run
}

// ShowBoxedError displays error in red box with suggestions
func ShowBoxedError(err ErrorSuggestion) {
    content := err.Message
    if err.Suggestion != "" {
        content += "\n\n" + Msg("to_fix") + ":\n  " + err.Suggestion
    }
    if err.Command != "" {
        content += "\n\n" + Msg("then_run") + ": " + err.Command
    }

    pterm.DefaultBox.
        WithTitle(pterm.Red("❌ " + err.Title)).
        WithTitleTopLeft().
        WithBoxStyle(pterm.NewStyle(pterm.FgRed)).
        Println(content)
}
```

### 3. Update progress.go - Replace SimpleSpinner

Keep `SimpleSpinner` for backward compatibility but add pterm wrappers:

```go
// StartPtermSpinner creates and starts a pterm spinner
func StartPtermSpinner(msg string) *pterm.SpinnerPrinter {
    spinner, _ := pterm.DefaultSpinner.Start(msg)
    return spinner
}

// Deprecate: SimpleSpinner (keep for now, mark deprecated)
```

### 4. Update table.go - Add PrintUpdatesTable

```go
// ImageUpdate represents an image update info
type ImageUpdate struct {
    Image     string
    OldDigest string
    NewDigest string
}

// PrintUpdatesTable displays available updates as boxed table
func PrintUpdatesTable(updates []ImageUpdate) {
    if len(updates) == 0 {
        return
    }

    tableData := pterm.TableData{
        {Msg("col_image"), Msg("col_current"), Msg("col_new")},
    }

    for _, u := range updates {
        old := truncateDigest(u.OldDigest, 12)
        new := truncateDigest(u.NewDigest, 12)
        tableData = append(tableData, []string{u.Image, old, new})
    }

    pterm.DefaultSection.Println(Msg("updates_available"))
    pterm.DefaultTable.
        WithHasHeader(true).
        WithBoxed(true).
        WithData(tableData).
        Render()
}

func truncateDigest(digest string, maxLen int) string {
    if len(digest) > maxLen {
        return digest[:maxLen] + "..."
    }
    return digest
}
```

## Todo List

- [x] Create `pkg/ui/banner.go` with ShowCommandBanner, ShowCompletionBanner
- [x] Create `pkg/ui/errors.go` with ShowBoxedError
- [x] Update `pkg/ui/progress.go` - add StartPtermSpinner wrapper
- [x] Update `pkg/ui/table.go` - add PrintUpdatesTable, ImageUpdate struct
- [x] Add new i18n keys: `to_fix`, `then_run`, `col_image`, `col_current`, `col_new`
- [x] Run tests: `go test ./pkg/ui/...` ✅ PASS


## Success Criteria

1. `ShowCommandBanner("kk init", "Docker Stack Initialization")` renders boxed header
2. `ShowBoxedError(...)` renders red box with suggestion
3. `StartPtermSpinner("msg")` returns working pterm spinner
4. `PrintUpdatesTable(updates)` renders boxed table
5. All existing tests pass

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| pterm API changes | Low | Medium | Use stable API methods |
| Color not supported | Medium | Low | pterm handles gracefully |

## Security Considerations

- No security impact - UI-only changes

## Next Steps

After completion → proceed to [Phase 02](./phase-02-command-updates.md)

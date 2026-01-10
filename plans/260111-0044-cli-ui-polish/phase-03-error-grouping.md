# Phase 03: Error Grouping - Preflight Error Display

## Context
- **Parent Plan:** [plan.md](./plan.md)
- **Dependencies:** [Phase 01](./phase-01-quick-wins.md), [Phase 02](./phase-02-icons-colors.md)
- **Effort:** 30 minutes

## Overview

| Field | Value |
|-------|-------|
| Date | 2026-01-11 |
| Priority | P2 |
| Effort | 30m |
| Implementation Status | complete |

**Description:** Group multiple preflight errors trong single boxed display thay vì multiple separate messages.

## Current Behavior

```
❌ Docker not running
❌ docker-compose.yml not found
❌ .env file missing
```

## Target Behavior

```
┌─ ❌ Preflight Checks Failed ────────────────────────┐
│                                                      │
│  1. Docker not running                               │
│     → Start Docker: systemctl start docker           │
│                                                      │
│  2. docker-compose.yml not found                     │
│     → Run: kk init                                   │
│                                                      │
│  3. .env file missing                                │
│     → Run: kk init                                   │
│                                                      │
└──────────────────────────────────────────────────────┘
```

## Tasks

### 1. Add ShowBoxedErrors Function

**File:** `pkg/ui/errors.go`

```go
// ShowBoxedErrors displays multiple errors in a single box.
func ShowBoxedErrors(title string, errors []ErrorSuggestion) {
    if len(errors) == 0 {
        return
    }

    var content strings.Builder
    for i, err := range errors {
        content.WriteString(fmt.Sprintf("%d. %s\n", i+1, err.Message))
        if err.Suggestion != "" {
            content.WriteString(fmt.Sprintf("   → %s\n", err.Suggestion))
        }
        if err.Command != "" {
            content.WriteString(fmt.Sprintf("   → Run: %s\n", err.Command))
        }
        if i < len(errors)-1 {
            content.WriteString("\n")
        }
    }

    pterm.DefaultBox.
        WithTitle(pterm.Red("❌ " + title)).
        WithTitleTopLeft().
        WithBoxStyle(pterm.NewStyle(pterm.FgRed)).
        Println(content.String())
}
```

### 2. Update Preflight Display

**File:** `pkg/validator/preflight.go` or `cmd/start.go`

```go
// Collect all errors
var errors []ui.ErrorSuggestion
for _, result := range results {
    if !result.Passed {
        errors = append(errors, ui.ErrorSuggestion{
            Message:    result.Name + ": " + result.Error,
            Suggestion: result.Fix,
            Command:    result.FixCommand,
        })
    }
}

// Display grouped
if len(errors) > 0 {
    ui.ShowBoxedErrors(ui.Msg("preflight_failed"), errors)
}
```

### 3. Add Fix Suggestions to Preflight Results

**File:** `pkg/validator/preflight.go`

Add `Fix` and `FixCommand` fields to `PreflightResult` struct if not exists.

## Todo List

- [x] Add `ShowBoxedErrors` function to `pkg/ui/errors.go`
- [x] Add `Fix`, `FixCommand` fields to `PreflightResult` if needed
- [x] Update preflight display logic in `cmd/start.go`
- [x] Add fix suggestions for each preflight check
- [x] Run `go build ./...`
- [x] Test with missing Docker, compose file, etc.

## Success Criteria

1. Multiple preflight errors display in single box
2. Each error shows numbered with fix suggestion
3. Clear visual hierarchy
4. Build passes

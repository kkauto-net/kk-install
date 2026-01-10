# Phase 01: Quick Wins - Box AccessInfo + Integrate ShowBoxedError

## Context
- **Parent Plan:** [plan.md](./plan.md)
- **Dependencies:** None
- **Effort:** 30 minutes

## Overview

| Field | Value |
|-------|-------|
| Date | 2026-01-11 |
| Priority | P0 |
| Effort | 30m |
| Implementation Status | complete |

**Description:** Box PrintAccessInfo table và integrate ShowBoxedError thay ShowError trong tất cả commands.

## Tasks

### 1. Box PrintAccessInfo Table

**File:** `pkg/ui/table.go`

```go
// Change line 127
- pterm.DefaultTable.WithHasHeader(true).WithData(tableData).Render()
+ pterm.DefaultTable.WithHasHeader(true).WithBoxed(true).WithData(tableData).Render()
```

### 2. Replace ShowError với ShowBoxedError

**Common error patterns to convert:**

#### cmd/status.go
```go
// Line ~43
- return fmt.Errorf("%s: %w", ui.Msg("get_status_failed"), err)
+ ui.ShowBoxedError(ui.ErrorSuggestion{
+     Title:      ui.Msg("get_status_failed"),
+     Message:    err.Error(),
+     Suggestion: "Check if Docker is running",
+     Command:    "docker ps",
+ })
+ return err
```

#### cmd/start.go
```go
// Preflight failed
ui.ShowBoxedError(ui.ErrorSuggestion{
    Title:      ui.Msg("preflight_failed"),
    Message:    "One or more preflight checks failed",
    Suggestion: "Fix the issues above and try again",
})

// Start failed
ui.ShowBoxedError(ui.ErrorSuggestion{
    Title:      ui.Msg("start_failed"),
    Message:    err.Error(),
    Suggestion: "Check Docker logs for details",
    Command:    "docker compose logs",
})
```

#### cmd/restart.go
```go
ui.ShowBoxedError(ui.ErrorSuggestion{
    Title:      ui.Msg("restart_failed"),
    Message:    err.Error(),
    Suggestion: "Check if services are running",
    Command:    "kk status",
})
```

#### cmd/update.go
```go
ui.ShowBoxedError(ui.ErrorSuggestion{
    Title:      ui.Msg("pull_failed"),
    Message:    err.Error(),
    Suggestion: "Check internet connection or Docker Hub status",
})
```

#### cmd/init.go
```go
// Docker check errors
ui.ShowBoxedError(ui.ErrorSuggestion{
    Title:      "Docker Not Ready",
    Message:    err.Error(),
    Suggestion: "Install Docker or start Docker daemon",
    Command:    "systemctl start docker",
})
```

## Todo List

- [x] Box `PrintAccessInfo` table (1 line change)
- [x] Update `cmd/status.go` - ShowBoxedError
- [x] Update `cmd/start.go` - ShowBoxedError for preflight/start errors
- [x] Update `cmd/restart.go` - ShowBoxedError
- [x] Update `cmd/update.go` - ShowBoxedError
- [x] Update `cmd/init.go` - ShowBoxedError for Docker errors
- [x] Add i18n keys for suggestions if needed
- [x] Run `go build ./...`
- [x] Test each command

## Success Criteria

1. `PrintAccessInfo` displays boxed table
2. All error messages use boxed format with suggestions
3. Build passes
4. Commands work correctly

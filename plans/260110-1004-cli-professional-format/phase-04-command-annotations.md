---
phase: 04
title: Command Group Annotations
status: pending
effort: 15m
---

# Phase 04: Command Annotations

## Context

- Parent plan: [plan.md](plan.md)
- Dependencies: Phase 02 (templates use annotations)

## Overview

Add `Annotations` map to each command defining its group for help template grouping.

## Requirements

1. Each command has `Annotations: map[string]string{"group": "..."}`
2. Groups: "core", "management", "additional"

## Command Classification

| Command | Group | Rationale |
|---------|-------|-----------|
| init | core | Primary setup workflow |
| start | core | Primary run workflow |
| status | core | Primary monitoring |
| restart | management | Operational task |
| update | management | Operational task |
| completion | additional | Utility |

## Implementation Steps

### 1. Update cmd/init.go

```go
var initCmd = &cobra.Command{
    Use:         "init",
    Short:       "Initialize Docker stack with interactive setup",
    Long:        `Create docker-compose.yml and required config files.`,
    Annotations: map[string]string{"group": "core"},
    RunE:        runInit,
}
```

### 2. Update cmd/start.go

```go
var startCmd = &cobra.Command{
    Use:         "start",
    Short:       "Start all services with preflight checks",
    Long:        `Run preflight checks, then start all services.`,
    Annotations: map[string]string{"group": "core"},
    RunE:        runStart,
}
```

### 3. Update cmd/status.go

```go
var statusCmd = &cobra.Command{
    Use:         "status",
    Short:       "View service status and health",
    Long:        `Display status of all containers in the stack.`,
    Annotations: map[string]string{"group": "core"},
    RunE:        runStatus,
}
```

### 4. Update cmd/restart.go

```go
var restartCmd = &cobra.Command{
    Use:         "restart",
    Short:       "Restart all services",
    Long:        `Restart all containers in the stack.`,
    Annotations: map[string]string{"group": "management"},
    RunE:        runRestart,
}
```

### 5. Update cmd/update.go

```go
var updateCmd = &cobra.Command{
    Use:         "update",
    Short:       "Pull latest images and recreate containers",
    Long:        `Check and download new images from Docker Hub, then restart services.`,
    Annotations: map[string]string{"group": "management"},
    RunE:        runUpdate,
}
```

### 6. Update cmd/completion.go

```go
var completionCmd = &cobra.Command{
    Use:         "completion [bash|zsh|fish]",
    Short:       "Generate shell completion scripts",
    Annotations: map[string]string{"group": "additional"},
    // ... rest unchanged
}
```

## Todo List

- [ ] Add annotation to init.go
- [ ] Add annotation to start.go
- [ ] Add annotation to status.go
- [ ] Add annotation to restart.go
- [ ] Add annotation to update.go
- [ ] Add annotation to completion.go
- [ ] Update Short descriptions to English

## Success Criteria

- [ ] All commands have group annotation
- [ ] `kk --help` shows grouped commands
- [ ] Short descriptions are in English (default language)

## Files Changed

| File | Action |
|------|--------|
| `cmd/init.go` | MODIFY |
| `cmd/start.go` | MODIFY |
| `cmd/status.go` | MODIFY |
| `cmd/restart.go` | MODIFY |
| `cmd/update.go` | MODIFY |
| `cmd/completion.go` | MODIFY |

## Notes

- Short descriptions should be in English since they're used in help template
- Long descriptions can use i18n if needed for runtime help
- The help template will handle translation separately

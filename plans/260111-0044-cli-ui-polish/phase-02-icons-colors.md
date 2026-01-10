# Phase 02: Icons & Colors - Starting State + Refinements

## Context
- **Parent Plan:** [plan.md](./plan.md)
- **Dependencies:** [Phase 01](./phase-01-quick-wins.md)
- **Effort:** 30 minutes

## Overview

| Field | Value |
|-------|-------|
| Date | 2026-01-11 |
| Priority | P1 |
| Effort | 30m |
| Implementation Status | pending |

**Description:** Add "starting" state với blue icon, refine color scheme cho consistency.

## Icon Set

| State | Icon | Color | pterm Function |
|-------|------|-------|----------------|
| Running | ● | Green | `pterm.Green()` |
| Stopped | ○ | Red | `pterm.Red()` |
| Starting | ◐ | Blue | `pterm.Blue()` |
| Healthy | ✓ | Green | `pterm.Green()` |
| Unhealthy | ✗ | Red | `pterm.Red()` |
| Warning | ⚠ | Yellow | `pterm.Yellow()` |
| Unknown | ? | Gray | `pterm.Gray()` |

## Tasks

### 1. Add Constants for Icons

**File:** `pkg/ui/icons.go` (new)

```go
package ui

// Status icons for CLI output
const (
    IconRunning   = "●"  // Green - service running
    IconStopped   = "○"  // Red - service stopped
    IconStarting  = "◐"  // Blue - service starting
    IconHealthy   = "✓"  // Green - health check passed
    IconUnhealthy = "✗"  // Red - health check failed
    IconWarning   = "⚠"  // Yellow - warning state
    IconUnknown   = "?"  // Gray - unknown state
)
```

### 2. Update PrintStatusTable

**File:** `pkg/ui/table.go`

```go
func PrintStatusTable(statuses []monitor.ServiceStatus) {
    // ... existing code ...

    for _, s := range statuses {
        var statusText string
        switch {
        case s.Running:
            statusText = pterm.Green(IconRunning + " " + Msg("status_running"))
        case s.Starting: // if field exists
            statusText = pterm.Blue(IconStarting + " " + Msg("status_starting"))
        default:
            statusText = pterm.Red(IconStopped + " " + Msg("status_stopped"))
        }
        // ...
    }
}
```

### 3. Update ShowServiceProgress

**File:** `pkg/ui/progress.go`

```go
func ShowServiceProgress(serviceName, status string) {
    switch status {
    case "starting":
        pterm.Info.Printfln("%s %s %s", IconStarting, serviceName, Msg("starting"))
    case "healthy", "running":
        pterm.Success.Printfln("%s %s %s", IconHealthy, serviceName, Msg("ready"))
    case "unhealthy":
        pterm.Error.Printfln("%s %s %s", IconUnhealthy, serviceName, Msg("unhealthy"))
    default:
        pterm.Warning.Printfln("%s %s: %s", IconWarning, serviceName, status)
    }
}
```

### 4. Update formatHealth

**File:** `pkg/ui/table.go`

```go
func formatHealth(health string) string {
    switch health {
    case "":
        return pterm.Gray("-")
    case "healthy":
        return pterm.Green(IconHealthy + " healthy")
    case "unhealthy":
        return pterm.Red(IconUnhealthy + " unhealthy")
    case "starting":
        return pterm.Blue(IconStarting + " starting")
    default:
        return pterm.Yellow(IconWarning + " " + health)
    }
}
```

## Todo List

- [ ] Create `pkg/ui/icons.go` with icon constants
- [ ] Update `PrintStatusTable` to use new icons
- [ ] Update `ShowServiceProgress` to use new icons
- [ ] Update `formatHealth` to use new icons
- [ ] Add i18n key: `status_starting`
- [ ] Run `go build ./...`
- [ ] Test visually

## Success Criteria

1. Starting services show blue ◐ icon
2. Health status shows appropriate icons
3. Colors consistent across all outputs
4. Build passes

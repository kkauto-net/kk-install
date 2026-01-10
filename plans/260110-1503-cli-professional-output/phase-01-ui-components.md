# Phase 01: UI Components

## Context Links
- Parent: [plan.md](./plan.md)
- Brainstorm: [brainstorm report](../reports/brainstorm-260110-1503-cli-professional-output.md)

## Overview
- **Priority**: High
- **Status**: Done
- **Description**: Create/update UI helper functions for pterm tables and step progress
- **Review**: [code-reviewer-260110-1525-phase01-ui-components.md](../reports/code-reviewer-260110-1525-phase01-ui-components.md)

## Key Insights
- pterm already in project - use `pterm.DefaultTable` for consistent styling
- Need boxed tables for status display
- Step headers should use `pterm.DefaultSection` for visual hierarchy
- Keep existing icons (`IconCheck`, `IconDocker`, etc.)

## Requirements

### Functional
1. `PrintStatusTable()` - pterm table with colored status indicators
2. `PrintAccessInfo()` - pterm table for service URLs
3. `ShowStepHeader()` - step progress indicator (Step 1/5: Title)
4. `PrintInitSummary()` - configuration summary table + created files list
5. `boolToStatus()` - helper for "✓ Enabled" / "○ Disabled" display

### Non-Functional
- Consistent styling across all tables
- Support narrow terminals (auto-wrap)
- Use i18n keys for all user-facing strings

## Architecture

```
pkg/ui/
├── table.go      # PrintStatusTable, PrintAccessInfo (MODIFY)
├── progress.go   # ShowStepHeader, PrintInitSummary, boolToStatus (NEW)
├── messages.go   # Keep existing icon constants
└── i18n.go       # Existing i18n infrastructure
```

## Related Code Files

| File | Action | Description |
|------|--------|-------------|
| `pkg/ui/table.go` | Modify | Rewrite with pterm tables |
| `pkg/ui/progress.go` | Create | New step/summary helpers |

## Implementation Steps

### 1. Update `pkg/ui/table.go`

```go
package ui

import (
    "github.com/pterm/pterm"
    "github.com/kkauto-net/kk-install/pkg/monitor"
)

// PrintStatusTable displays service status using pterm table
func PrintStatusTable(statuses []monitor.ServiceStatus) {
    pterm.DefaultSection.Println(Msg("service_status"))

    tableData := pterm.TableData{
        {Msg("col_service"), Msg("col_status"), Msg("col_health"), Msg("col_ports")},
    }

    for _, s := range statuses {
        statusText := pterm.Green("● Running")
        if !s.Running {
            statusText = pterm.Red("○ Stopped")
        }

        health := formatHealth(s.Health)
        ports := truncatePorts(s.Ports, 30)

        tableData = append(tableData, []string{
            s.Name,
            statusText,
            health,
            ports,
        })
    }

    pterm.DefaultTable.
        WithHasHeader(true).
        WithBoxed(true).
        WithData(tableData).
        Render()
}

func formatHealth(health string) string {
    if health == "" {
        return pterm.Gray("-")
    }
    if health == "healthy" {
        return pterm.Green("healthy")
    }
    if health == "unhealthy" {
        return pterm.Red("unhealthy")
    }
    return pterm.Yellow(health)
}

func truncatePorts(ports string, maxLen int) string {
    if ports == "" {
        return "-"
    }
    if len(ports) > maxLen {
        return ports[:maxLen-3] + "..."
    }
    return ports
}

// PrintAccessInfo shows access URLs for services
func PrintAccessInfo(statuses []monitor.ServiceStatus) {
    pterm.DefaultSection.Println(Msg("access_info"))

    tableData := pterm.TableData{
        {Msg("col_service"), Msg("col_url")},
    }

    for _, s := range statuses {
        if !s.Running {
            continue
        }
        url := getServiceURL(s.Name, s.Ports)
        if url != "" {
            tableData = append(tableData, []string{s.Name, url})
        }
    }

    if len(tableData) > 1 {
        pterm.DefaultTable.WithHasHeader(true).WithData(tableData).Render()
    }
}

func getServiceURL(name, ports string) string {
    switch name {
    case "kkengine":
        return "http://localhost:8019"
    case "db":
        return "localhost:3307"
    case "caddy":
        return "http://localhost (HTTPS: https://localhost)"
    default:
        return ""
    }
}
```

### 2. Create `pkg/ui/progress.go`

```go
package ui

import (
    "fmt"
    "github.com/pterm/pterm"
)

// ShowStepHeader displays step progress indicator
func ShowStepHeader(current, total int, title string) {
    stepText := fmt.Sprintf("Step %d/%d", current, total)
    pterm.DefaultSection.
        WithLevel(2).
        Println(fmt.Sprintf("%s: %s", stepText, title))
}

// PrintInitSummary shows configuration summary and created files
func PrintInitSummary(enableSeaweedFS, enableCaddy bool, domain string, createdFiles []string) {
    // Configuration Summary
    pterm.DefaultSection.Println(Msg("config_summary"))

    configData := pterm.TableData{
        {Msg("col_setting"), Msg("col_value")},
        {"SeaweedFS", boolToStatus(enableSeaweedFS)},
        {"Caddy", boolToStatus(enableCaddy)},
    }
    if enableCaddy && domain != "" {
        configData = append(configData, []string{Msg("domain"), domain})
    }

    pterm.DefaultTable.WithHasHeader(true).WithData(configData).Render()

    // Created Files
    fmt.Println()
    pterm.DefaultSection.Println(Msg("created_files"))

    for _, f := range createdFiles {
        pterm.Success.Println(f)
    }
}

// boolToStatus returns colored enabled/disabled status
func boolToStatus(b bool) string {
    if b {
        return pterm.Green("✓ " + Msg("enabled"))
    }
    return pterm.Gray("○ " + Msg("disabled"))
}
```

## Todo List

- [x] Update `pkg/ui/table.go` with pterm implementation
- [x] Create `pkg/ui/progress.go` with step/summary helpers
- [x] Run `go build` to verify no compile errors
- [ ] **BLOCKER**: Add 13 missing i18n keys (service_status, col_service, col_status, col_health, col_ports, access_info, col_url, config_summary, col_setting, col_value, created_files, enabled, disabled)
- [ ] Fix hardcoded English strings ("Running", "Stopped")
- [ ] Run `golangci-lint` to check code quality (tool not available)

## Success Criteria

- [x] `PrintStatusTable` renders boxed pterm table with colored status
- [x] `PrintAccessInfo` renders clean URL table
- [x] `ShowStepHeader` shows "Step X/Y: Title" format
- [x] `PrintInitSummary` shows config table + files list
- [ ] **INCOMPLETE**: All functions use i18n keys for strings (13 keys missing)

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| pterm API changes | Low | Pin version in go.mod |
| Terminal width issues | Medium | pterm auto-wraps, test narrow terminals |

## Security Considerations

- No security concerns - UI display only

## Next Steps

- Phase 02: Add i18n message keys

# Phase 02: Command Updates

## Context
- **Parent Plan:** [plan.md](./plan.md)
- **Dependencies:** [Phase 01](./phase-01-core-ui-components.md) must be complete
- **Brainstorm:** [brainstorm report](../reports/brainstorm-260110-1620-cli-professional-output-v2.md)

## Overview

| Field | Value |
|-------|-------|
| Date | 2026-01-10 |
| Priority | P2 |
| Effort | 1.5h |
| Implementation Status | completed |
| Completion Date | 2026-01-11 |
| Review Status | pending |

**Description:** Update all kk commands to use new UI components - add command banners, use pterm spinners, enhance tables.

## Key Insights

1. Each command needs consistent banner at start
2. `kk status` simplest - just add banner
3. `kk init` already has steps, needs banner + box all tables
4. `kk start/restart` need pterm spinners for health monitoring
5. `kk update` needs updates table instead of plain list

## Requirements

- R1: Boxed tables for all displays
- R2: Verbose mode with step-by-step + summary
- R3: Professional animations

## Related Code Files

| File | Action | Description |
|------|--------|-------------|
| `cmd/status.go` | MODIFY | Add command banner |
| `cmd/init.go` | MODIFY | Add banner, box InitSummary table |
| `cmd/start.go` | MODIFY | Add banner, use pterm spinner |
| `cmd/restart.go` | MODIFY | Add banner, use pterm spinner |
| `cmd/update.go` | MODIFY | Add banner, use PrintUpdatesTable |
| `pkg/ui/progress.go` | MODIFY | Update ShowServiceProgress with pterm |
| `pkg/ui/table.go` | MODIFY | Box PrintInitSummary |

## Implementation Steps

### 1. Update cmd/status.go

```go
func runStatus(cmd *cobra.Command, args []string) error {
    // ADD: Command banner
    ui.ShowCommandBanner("kk status", ui.Msg("status_desc"))

    // ... existing code ...
}
```

### 2. Update cmd/init.go

```go
func runInit(cmd *cobra.Command, args []string) error {
    // ADD: Command banner at start
    ui.ShowCommandBanner("kk init", ui.Msg("init_desc"))

    // ... existing code ...

    // MODIFY: Final completion - use ShowCompletionBanner
    ui.ShowCompletionBanner(true,
        ui.IconComplete+" "+ui.Msg("init_complete"),
        ui.Msg("next_steps_box"))
}
```

### 3. Update cmd/start.go

```go
func runStart(cmd *cobra.Command, args []string) error {
    // ADD: Command banner
    ui.ShowCommandBanner("kk start", ui.Msg("start_desc"))

    // ... existing code ...

    // MODIFY: Step 2 - Use pterm spinner for docker-compose up
    spinner := ui.StartPtermSpinner(ui.Msg("starting_services"))
    if err := executor.Up(timeoutCtx); err != nil {
        spinner.Fail(ui.Msg("start_failed"))
        return fmt.Errorf("%s: %w", ui.Msg("start_failed"), err)
    }
    spinner.Success(ui.Msg("services_started"))

    // ... health monitoring ...
}
```

### 4. Update cmd/restart.go

Similar to start.go - add banner, use pterm spinner.

### 5. Update cmd/update.go

```go
func runUpdate(cmd *cobra.Command, args []string) error {
    // ADD: Command banner
    ui.ShowCommandBanner("kk update", ui.Msg("update_desc"))

    // ... Step 1: Pull with spinner (already exists) ...

    // MODIFY: Step 2 - Use PrintUpdatesTable instead of plain list
    if len(updates) > 0 {
        uiUpdates := make([]ui.ImageUpdate, len(updates))
        for i, u := range updates {
            uiUpdates[i] = ui.ImageUpdate{
                Image:     u.Image,
                OldDigest: u.OldDigest,
                NewDigest: u.NewDigest,
            }
        }
        ui.PrintUpdatesTable(uiUpdates)
    }

    // ... rest of flow ...
}
```

### 6. Update pkg/ui/progress.go - ShowServiceProgress

Replace plain text with pterm:

```go
func ShowServiceProgress(serviceName, status string) {
    switch status {
    case "starting":
        pterm.Info.Printfln("%s %s", serviceName, Msg("starting"))
    case "healthy", "running":
        pterm.Success.Printfln("%s %s", serviceName, Msg("ready"))
    case "unhealthy":
        pterm.Error.Printfln("%s %s", serviceName, Msg("unhealthy"))
    default:
        pterm.Warning.Printfln("%s: %s", serviceName, status)
    }
}
```

### 7. Update pkg/ui/table.go - PrintInitSummary

Add boxing to config table:

```go
func PrintInitSummary(enableSeaweedFS, enableCaddy bool, domain string, createdFiles []string) {
    // Configuration Summary - WITH BOX
    pterm.DefaultSection.Println(Msg("config_summary"))

    configData := pterm.TableData{
        {Msg("col_setting"), Msg("col_value")},
        {"SeaweedFS", boolToStatus(enableSeaweedFS)},
        {"Caddy", boolToStatus(enableCaddy)},
    }
    if enableCaddy && domain != "" {
        configData = append(configData, []string{Msg("domain"), domain})
    }

    pterm.DefaultTable.
        WithHasHeader(true).
        WithBoxed(true).  // ADD BOXING
        WithData(configData).
        Render()

    // Created Files - WITH BOX
    fmt.Println()
    pterm.DefaultSection.Println(Msg("created_files"))

    fileData := pterm.TableData{{Msg("col_file")}}
    for _, f := range createdFiles {
        fileData = append(fileData, []string{pterm.Green("✓ " + f)})
    }

    pterm.DefaultTable.
        WithHasHeader(true).
        WithBoxed(true).  // ADD BOXING
        WithData(fileData).
        Render()
}
```

## Todo List

- [ ] Update `cmd/status.go` - add ShowCommandBanner
- [ ] Update `cmd/init.go` - add banner, use ShowCompletionBanner
- [ ] Update `cmd/start.go` - add banner, use pterm spinner
- [ ] Update `cmd/restart.go` - add banner, use pterm spinner
- [ ] Update `cmd/update.go` - add banner, use PrintUpdatesTable
- [ ] Update `pkg/ui/progress.go` - pterm in ShowServiceProgress
- [ ] Update `pkg/ui/table.go` - box PrintInitSummary
- [ ] Add i18n keys: `status_desc`, `init_desc`, `start_desc`, `update_desc`, `starting`, `ready`, `col_file`, `services_started`
- [ ] Run `go build ./...` to verify
- [ ] Test each command manually

## Success Criteria

1. `kk status` shows banner at top
2. `kk init` shows banner, boxed tables, completion banner
3. `kk start` shows banner, pterm spinner, status table
4. `kk restart` shows banner, pterm spinner
5. `kk update` shows banner, boxed updates table
6. All builds without errors

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Spinner conflicts with output | Medium | Medium | Test interactive scenarios |
| Banner too wide for terminal | Low | Low | pterm handles wrapping |

## Security Considerations

- No security impact - UI-only changes

## Next Steps

After completion → proceed to [Phase 03](./phase-03-i18n-polish.md)

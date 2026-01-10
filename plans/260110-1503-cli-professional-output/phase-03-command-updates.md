# Phase 03: Command Updates

## Context Links
- Parent: [plan.md](./plan.md)
- Depends on: [Phase 01](./phase-01-ui-components.md), [Phase 02](./phase-02-i18n-messages.md)

## Overview
- **Priority**: High
- **Status**: Pending
- **Description**: Update command files to use new pterm UI components

## Key Insights
- `cmd/init.go` needs step wizard (5 steps)
- `cmd/status.go` already calls `ui.PrintStatusTable` - just needs updated function
- `cmd/start.go` uses preflight results + status table
- `pkg/validator/preflight.go` has `PrintPreflightResults` function

## Requirements

### Functional
1. `kk init` - Add step headers (Step 1/5 ... Step 5/5)
2. `kk init` - Show summary table at completion
3. `kk status` - Use updated table functions (already wired)
4. `kk start` - Use updated table functions (already wired)
5. `kk start` - Update preflight results display

## Related Code Files

| File | Action | Description |
|------|--------|-------------|
| `cmd/init.go` | Modify | Add step headers, summary table |
| `cmd/status.go` | Verify | Already uses PrintStatusTable |
| `cmd/start.go` | Verify | Already uses PrintStatusTable |
| `pkg/validator/preflight.go` | Modify | Update PrintPreflightResults with pterm |

## Implementation Steps

### 1. Update `cmd/init.go`

Add step headers throughout the init flow:

```go
func runInit(cmd *cobra.Command, args []string) error {
    // Step 1: Check Docker
    ui.ShowStepHeader(1, 5, ui.Msg("step_docker_check"))
    ui.ShowInfo(ui.IconDocker + " " + ui.MsgCheckingDocker())
    // ... existing docker checks ...

    // Step 2: Language selection
    ui.ShowStepHeader(2, 5, ui.Msg("step_language"))
    // ... existing language form ...

    // Step 3: Configuration options
    ui.ShowStepHeader(3, 5, ui.Msg("step_options"))
    // ... existing SeaweedFS/Caddy options ...

    // Step 4: Generate files
    ui.ShowStepHeader(4, 5, ui.Msg("step_generate"))
    spinner, _ := pterm.DefaultSpinner.Start(ui.IconWrite + " " + ui.Msg("generating_files"))
    // ... existing file generation ...

    // Step 5: Complete - show summary
    ui.ShowStepHeader(5, 5, ui.Msg("step_complete"))

    // Collect created files
    createdFiles := []string{"docker-compose.yml", ".env", "kkphp.conf"}
    if enableCaddy {
        createdFiles = append(createdFiles, "Caddyfile")
    }
    if enableSeaweedFS {
        createdFiles = append(createdFiles, "kkfiler.toml")
    }

    // Show summary table
    ui.PrintInitSummary(enableSeaweedFS, enableCaddy, domain, createdFiles)

    // Show next steps box (keep existing)
    fmt.Println()
    pterm.DefaultBox.
        WithTitle(ui.IconComplete + " " + ui.Msg("init_complete")).
        WithTitleTopCenter().
        WithBoxStyle(pterm.NewStyle(pterm.FgGreen)).
        Println(ui.Msg("next_steps_box"))

    return nil
}
```

### 2. Update `pkg/validator/preflight.go`

Replace text output with pterm table:

```go
// PrintPreflightResults displays preflight check results as table
func PrintPreflightResults(results []PreflightResult) {
    tableData := pterm.TableData{
        {ui.Msg("check"), ui.Msg("result")},
    }

    for _, r := range results {
        status := pterm.Green("✓ Pass")
        if !r.Passed {
            if r.Error != "" {
                status = pterm.Red("✗ " + r.Error)
            } else {
                status = pterm.Red("✗ Failed")
            }
        }
        tableData = append(tableData, []string{r.Name, status})
    }

    pterm.DefaultTable.
        WithHasHeader(true).
        WithBoxed(true).
        WithData(tableData).
        Render()
}
```

### 3. Verify `cmd/status.go`

No changes needed - already calls `ui.PrintStatusTable` and `ui.PrintAccessInfo`:

```go
// Line 49-50 - already correct
ui.PrintStatusTable(statuses)
ui.PrintAccessInfo(statuses)
```

### 4. Verify `cmd/start.go`

No changes needed for table calls - already uses:
- `validator.PrintPreflightResults(results)` - will use updated function
- `ui.PrintStatusTable(statuses)` - will use updated function
- `ui.PrintAccessInfo(statuses)` - will use updated function

## Todo List

- [ ] Update `cmd/init.go` with step headers
- [ ] Update `cmd/init.go` to call PrintInitSummary
- [ ] Update `pkg/validator/preflight.go` PrintPreflightResults
- [ ] Verify `cmd/status.go` compiles with new table.go
- [ ] Verify `cmd/start.go` compiles with new table.go
- [ ] Run full build test

## Success Criteria

- [ ] `kk init` shows Step 1/5 through Step 5/5
- [ ] `kk init` displays summary table before next steps box
- [ ] `kk status` renders pterm boxed table
- [ ] `kk start` preflight shows pterm table
- [ ] All commands compile and run correctly

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| Breaking existing functionality | High | Test each command before/after |
| Import cycle | Medium | Keep ui package dependencies clean |

## Security Considerations

- No security concerns - output formatting only

## Next Steps

- Phase 04: End-to-end testing

# Brainstorm: CLI Professional Output Enhancement v2

**Date:** 2026-01-10
**Status:** Draft

## Problem Statement

Current kkcli command outputs lack visual consistency and professional appearance:
- Mixed output styles (plain text vs pterm)
- `kk init` too compact for CLI beginners
- Tables not used consistently across commands
- Error messages not beginner-friendly

## Requirements

| ID | Requirement |
|----|-------------|
| R1 | Boxed tables for all status/info displays |
| R2 | Verbose mode with step-by-step + summary |
| R3 | Professional animations (spinners, progress bars) |
| R4 | Standard color scheme (green=success, red=error, etc.) |
| R5 | Boxed errors with fix suggestions |
| R6 | Default English, Vietnamese với dấu if selected |

## Current State Analysis

### Existing pterm Components (pkg/ui/)
- `PrintStatusTable()` - Boxed table ✅
- `PrintPreflightResults()` - Boxed table ✅
- `PrintInitSummary()` - Non-boxed table ❌
- `ShowStepHeader()` - pterm Section ✅
- `ShowServiceProgress()` - Plain text `[OK]`, `[>]` ❌
- `SimpleSpinner` - Custom implementation ❌ (should use pterm)

### Commands to Enhance

| Command | Current State | Issues |
|---------|---------------|--------|
| `kk init` | Steps + summary table | Non-boxed config table, no command header |
| `kk start` | Steps + status table | Plain service progress, no command header |
| `kk status` | Boxed status table | No command header |
| `kk restart` | Steps + status table | Same as start |
| `kk update` | Steps + plain list | Updates shown as plain text, not table |

## Proposed Solutions

### Solution A: Incremental Enhancement (Recommended)

Enhance existing pterm usage, add missing components, standardize output.

**Pros:**
- Low risk, builds on existing code
- Maintains backward compatibility
- Focused changes per command

**Cons:**
- May need multiple phases
- Some code duplication during transition

### Solution B: Complete UI Refactor

Create new unified `ui/output.go` with all rendering functions.

**Pros:**
- Clean architecture
- DRY principle
- Easier testing

**Cons:**
- Higher risk
- More time investment
- Potential regressions

## Recommended Approach: Solution A

### New UI Components to Add

```go
// pkg/ui/banner.go
func ShowCommandBanner(cmd, description string)  // Header cho mỗi command
func ShowCompletionBanner(success bool, msg string) // Footer với status

// pkg/ui/errors.go
func ShowBoxedError(title, message, suggestion string) // Error box với suggestions

// pkg/ui/table.go (enhance)
func PrintUpdatesTable(updates []updater.ImageUpdate) // Table cho updates

// pkg/ui/progress.go (replace SimpleSpinner)
func StartSpinner(msg string) *pterm.SpinnerPrinter
func ShowProgressBar(total int) *pterm.ProgressbarPrinter
func ShowServiceStatus(services []ServiceProgress) // Live updating table
```

### Output Structure Per Command

#### `kk init` Enhanced Flow
```
╭──────────────────────────────────────────╮
│  kk init - Docker Stack Initialization   │
╰──────────────────────────────────────────╯

▶ Step 1/5: Docker Check
  ✓ Docker is installed
  ✓ Docker daemon is running
  ✓ Docker Compose v2.24.0

▶ Step 2/5: Language Selection
  [Interactive prompt]

▶ Step 3/5: Configuration Options
  [Interactive prompts]

▶ Step 4/5: Generate Files
  ⠋ Generating configuration files...
  ✓ Configuration files generated

▶ Step 5/5: Complete

┌─────────────────────────────────────────┐
│         Configuration Summary           │
├──────────────┬──────────────────────────┤
│ Setting      │ Value                    │
├──────────────┼──────────────────────────┤
│ SeaweedFS    │ ✓ Enabled                │
│ Caddy        │ ✓ Enabled                │
│ Domain       │ localhost                │
└──────────────┴──────────────────────────┘

┌─────────────────────────────────────────┐
│            Created Files                │
├─────────────────────────────────────────┤
│ ✓ docker-compose.yml                    │
│ ✓ .env                                  │
│ ✓ kkphp.conf                            │
│ ✓ Caddyfile                             │
└─────────────────────────────────────────┘

╭─────────────────────────────────────────╮
│  ✅ Initialization complete!            │
│                                         │
│  Next steps:                            │
│    1. Review and edit .env if needed    │
│    2. Run: kk start                     │
╰─────────────────────────────────────────╯
```

#### `kk status` Enhanced Flow
```
╭──────────────────────────────────────────╮
│  kk status - Service Status              │
╰──────────────────────────────────────────╯

┌─────────────────────────────────────────────────────────────────┐
│                        Service Status                           │
├───────────┬────────────────┬──────────┬────────────────────────┤
│ Service   │ Status         │ Health   │ Ports                  │
├───────────┼────────────────┼──────────┼────────────────────────┤
│ kkengine  │ ● Running      │ healthy  │ 8019->8019             │
│ db        │ ● Running      │ healthy  │ 3307->3306             │
│ redis     │ ● Running      │ -        │ 6379->6379             │
│ caddy     │ ● Running      │ -        │ 80->80, 443->443       │
└───────────┴────────────────┴──────────┴────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                      Access Information                         │
├───────────┬─────────────────────────────────────────────────────┤
│ Service   │ URL                                                 │
├───────────┼─────────────────────────────────────────────────────┤
│ kkengine  │ http://localhost:8019                               │
│ db        │ localhost:3307                                      │
│ caddy     │ http://localhost (HTTPS: https://localhost)         │
└───────────┴─────────────────────────────────────────────────────┘

✅ All 4 services running
```

#### `kk update` Enhanced Flow
```
╭──────────────────────────────────────────╮
│  kk update - Pull & Recreate             │
╰──────────────────────────────────────────╯

▶ Step 1/4: Pull Images
  ⣾ Pulling images... [kkengine:latest]

▶ Step 2/4: Check Updates

┌───────────────────────────────────────────────────────────┐
│                    Updates Available                       │
├─────────────────────┬─────────────────┬───────────────────┤
│ Image               │ Current         │ New               │
├─────────────────────┼─────────────────┼───────────────────┤
│ kkengine:latest     │ sha256:abc123   │ sha256:def456     │
│ mariadb:11          │ sha256:111222   │ sha256:333444     │
└─────────────────────┴─────────────────┴───────────────────┘

? Restart services with new images? [Y/n]

▶ Step 3/4: Recreate Containers
  ⣾ Recreating kkengine...
  ✓ kkengine recreated
  ⣾ Recreating db...
  ✓ db recreated

▶ Step 4/4: Status
  ✅ Update complete!
  [Status table]
```

#### Error Display
```
╭─────────────────────────────────────────╮
│  ❌ Error: Docker Not Running           │
├─────────────────────────────────────────┤
│  Docker daemon is not responding.       │
│                                         │
│  To fix:                                │
│    sudo systemctl start docker          │
│    # or                                 │
│    sudo service docker start            │
│                                         │
│  Then run: kk start                     │
╰─────────────────────────────────────────╯
```

### Implementation Phases

#### Phase 1: Core UI Components
- [ ] Add `ShowCommandBanner()`
- [ ] Add `ShowBoxedError()` with suggestions
- [ ] Replace `SimpleSpinner` with pterm spinner
- [ ] Add `PrintUpdatesTable()` for update command

#### Phase 2: Command Updates
- [ ] `kk status` - Add command banner
- [ ] `kk init` - Box all tables, add completion banner
- [ ] `kk start` - Replace plain progress with pterm
- [ ] `kk restart` - Same as start
- [ ] `kk update` - Add updates table, enhance flow

#### Phase 3: I18n & Polish
- [ ] Update lang_en.go, lang_vi.go for new messages
- [ ] Ensure Vietnamese có dấu
- [ ] Add accessibility fallbacks (--no-color, --plain)

### Files to Modify

| File | Changes |
|------|---------|
| `pkg/ui/banner.go` | NEW - Command headers/footers |
| `pkg/ui/errors.go` | NEW - Boxed errors with suggestions |
| `pkg/ui/table.go` | Add `PrintUpdatesTable()`, box existing tables |
| `pkg/ui/progress.go` | Replace `SimpleSpinner` with pterm |
| `cmd/init.go` | Use new UI functions |
| `cmd/start.go` | Use new UI functions |
| `cmd/status.go` | Add command banner |
| `cmd/restart.go` | Use new UI functions |
| `cmd/update.go` | Use updates table |
| `pkg/ui/lang_en.go` | New message keys |
| `pkg/ui/lang_vi.go` | New message keys (với dấu) |

### pterm Features to Use

| Component | pterm Function |
|-----------|----------------|
| Command Banner | `pterm.DefaultBox.WithTitle()` |
| Tables | `pterm.DefaultTable.WithBoxed().WithHasHeader()` |
| Spinner | `pterm.DefaultSpinner.Start()` |
| Progress | `pterm.DefaultProgressbar.WithTotal()` |
| Success | `pterm.Success.Println()` |
| Error Box | `pterm.DefaultBox.WithBoxStyle(Red)` |
| Section | `pterm.DefaultSection.Println()` |
| Bullet List | `pterm.DefaultBulletList.WithItems()` |

### Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| pterm version incompatibility | High | Pin version in go.mod |
| Terminal without color support | Medium | Add `--no-color` flag |
| Unicode issues on Windows | Low | Test on Windows |
| Breaking existing output parsing | Medium | Maintain structure, only style |

### Success Metrics

1. All commands show consistent header/footer
2. All status/info uses boxed tables
3. All progress uses pterm spinners
4. Errors show suggestions for common issues
5. Output readable for CLI beginners

## Decision

**Recommended:** Solution A (Incremental Enhancement)

- Start with Phase 1 (Core UI Components)
- Test on all commands
- Proceed to Phase 2 & 3

## Next Steps

1. Create implementation plan with `/ck:plan:fast`
2. Implement Phase 1 core components
3. Update commands one by one
4. Update i18n files
5. Test with `kk init && kk start && kk status`

---

## Unresolved Questions

1. Có cần `--plain` flag cho CI/CD environments không?
2. Có muốn thêm `--verbose` / `--quiet` flags cho các commands không?

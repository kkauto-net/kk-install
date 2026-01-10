# Brainstorm: CLI Status/Error UI Improvements

**Date:** 2026-01-11
**Status:** Complete
**Focus:** Boxed tables, error messages, color scheme

---

## 1. Current State Analysis

### Existing UI Components (pkg/ui/)

| Component | File | Status |
|-----------|------|--------|
| `ShowCommandBanner` | banner.go | Boxed header with cyan title |
| `ShowCompletionBanner` | banner.go | Boxed footer (green/red) |
| `PrintStatusTable` | table.go | Boxed table with status columns |
| `PrintUpdatesTable` | table.go | Boxed table for image updates |
| `PrintAccessInfo` | table.go | **NOT boxed** - plain table |
| `ShowBoxedError` | errors.go | Red boxed error with suggestions |
| `ShowError/Warning/Info/Success` | messages.go | Plain pterm prefixed lines |
| `StartPtermSpinner` | progress.go | Animated spinner |
| `ShowStepHeader` | progress.go | Section-based step indicator |

### Color Scheme (Current)

| State | Color | Icon |
|-------|-------|------|
| Running | Green | `●` |
| Stopped | Red | `○` |
| Healthy | Green | (text only) |
| Unhealthy | Red | (text only) |
| No healthcheck | Gray | `-` |
| Unknown health | Yellow | (text only) |
| Enabled | Green | `✓` |
| Disabled | Gray | `○` |

---

## 2. Gap Analysis

### 2.1 Boxed Tables

| Issue | Severity | Notes |
|-------|----------|-------|
| `PrintAccessInfo` not boxed | Medium | Inconsistent với Status table |
| No terminal width handling | Low | pterm auto-handles, but long ports truncated |
| No empty state design | Low | Just "no services" text |

### 2.2 Error Messages

| Issue | Severity | Notes |
|-------|----------|-------|
| `ShowError()` is plain text | High | Lost in output, no suggestions |
| No error grouping | Medium | Multiple errors show separately |
| No error codes/categories | Low | Hard to search/document |
| `ShowBoxedError` exists but unused in commands | High | Only defined, never called |

### 2.3 Color Scheme

| Issue | Severity | Notes |
|-------|----------|-------|
| No "starting" state color | Medium | Currently defaults to yellow |
| Missing "warning" state for services | Low | Only health has yellow |
| Icons inconsistent: `●/○` vs `✓/○` | Low | Different patterns for status vs config |
| No colorblind mode | Medium | Red/green problematic for 8% males |

---

## 3. Proposed Solutions

### 3.1 Boxed Tables

**Option A: Box All Tables (Recommended)**
- Box `PrintAccessInfo` to match `PrintStatusTable`
- Consistent visual language
- Implementation: 1 line change

**Option B: Combine Status + Access into Single Table**
- Add URL column to status table
- Reduces visual clutter
- Con: Wide table, may truncate

**Option C: Keep Current (Not Recommended)**
- Access info intentionally lightweight
- Con: Visual inconsistency

**Recommendation:** Option A - simplest, maintains consistency

### 3.2 Error Messages

**Option A: Replace ShowError with ShowBoxedError (Recommended)**
```
┌─ Error: Docker not running ────────────────────┐
│ Docker daemon is not responding                │
│                                                │
│ To fix:                                        │
│   Start Docker Desktop or run: sudo systemctl  │
│   start docker                                 │
│                                                │
│ Then run: kk start                             │
└────────────────────────────────────────────────┘
```
- Already implemented in errors.go
- Need to integrate into commands

**Option B: Add Error Grouping**
```go
type ErrorGroup struct {
    Category string          // "preflight", "docker", "network"
    Errors   []ErrorSuggestion
}

func ShowErrorGroup(group ErrorGroup) {
    // Single box with multiple errors
}
```
- For preflight checks with multiple failures
- Show all issues at once

**Option C: Error Codes**
```
[ERR-001] Docker not running
[ERR-002] .env file missing
```
- Searchable, documentable
- Overkill for simple CLI

**Recommendation:**
- Phase 1: Option A - use existing ShowBoxedError
- Phase 2: Option B - error grouping for preflight

### 3.3 Color Scheme

**Option A: Enhanced Status Set (Recommended)**

| State | Color | Icon | Text |
|-------|-------|------|------|
| Running | Green | `●` | running |
| Starting | Blue | `◐` | starting |
| Stopped | Red | `○` | stopped |
| Healthy | Green | `✓` | healthy |
| Unhealthy | Red | `✗` | unhealthy |
| Starting (health) | Yellow | `◐` | starting |
| No healthcheck | Gray | `-` | - |
| Warning | Yellow | `⚠` | warning |

**Option B: Add Colorblind Mode**
```go
var ColorblindMode = false // Set via --colorblind flag or env

func StatusIcon(running bool) string {
    if ColorblindMode {
        if running { return "[RUN]" }
        return "[STOP]"
    }
    // Normal icons
}
```
- Use shapes/text instead of colors
- Fallback: `[OK]` `[FAIL]` `[WARN]`

**Option C: Semantic Colors Only**
- Keep current but add text labels always
- Color as enhancement, not sole indicator

**Recommendation:**
- Phase 1: Option A - better icons/states
- Phase 2: Option B - colorblind support via env var `NO_COLOR` or `--no-color`

---

## 4. Recommendations Summary

### Priority 1 (Quick Wins)
1. **Box AccessInfo table** - 1 line change in table.go
2. **Use ShowBoxedError in commands** - replace ShowError calls

### Priority 2 (Short Term)
3. **Add "starting" state** - blue icon `◐` for services coming up
4. **Error grouping for preflight** - single box for multiple failures

### Priority 3 (Future)
5. **Colorblind mode** - `NO_COLOR` env var support
6. **Error codes** - if documentation grows

---

## 5. Implementation Considerations

### Risks
- **Terminal compatibility:** Unicode icons may not render in all terminals
  - Mitigation: Fallback to ASCII (`[OK]`, `[X]`)
- **Box width:** Long service names/ports may break box alignment
  - Mitigation: pterm handles this, but test edge cases

### Testing
- Test in: iTerm2, Terminal.app, Windows Terminal, basic Linux tty
- Test with `NO_COLOR=1` environment variable
- Test narrow terminal (80 cols)

### Breaking Changes
- None expected - all additive

---

## 6. Next Steps

1. Decide on priority items to implement
2. Create implementation plan if proceeding
3. Consider adding accessibility tests

---

## Unresolved Questions

1. Should `PrintAccessInfo` show stopped services with "(stopped)" label or hide them entirely?
2. For error grouping, max errors before "and N more..." truncation?
3. Should colorblind mode be auto-detected or manual only?
4. Icon set: Unicode only or ASCII fallback in code?

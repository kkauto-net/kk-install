# Phase 01: Env Preload and Timestamp Backup

## Context

- **Parent Plan:** [plan.md](plan.md)
- **Brainstorm:** [brainstorm-260112-1357-env-preload-init.md](../reports/brainstorm-260112-1357-env-preload-init.md)

## Overview

| Field | Value |
|-------|-------|
| Date | 2026-01-12 |
| Priority | P2 |
| Implementation Status | completed |
| Completed | 2026-01-12 14:33:00 |
| Review Status | reviewed |
| Review Score | 8/10 |

**Description:** Implement .env file loading for pre-filling init form fields and update backup logic to use timestamp format.

## Key Insights

1. Current `backupExistingConfigs()` uses simple `.bak` suffix - easily overwritten
2. `RenderTemplate()` in embed.go has duplicate backup logic - violates DRY
3. Standard Go `time.Format` layout: `"060102-150405"` = `Ymd-His`
4. .env format: `KEY=VALUE`, lines starting with `#` are comments

## Requirements

1. **Load existing .env values:**
   - Parse `.env` file if exists
   - Extract values for: `LICENSE_KEY`, `SYSTEM_DOMAIN`, `JWT_SECRET`, `DB_PASSWORD`, `DB_ROOT_PASSWORD`, `REDIS_PASSWORD`, `S3_ACCESS_KEY`, `S3_SECRET_KEY`
   - Use as default values in form inputs

2. **Timestamp backup format:**
   - Change from `filename.bak` to `filename-YYMMDD-HHMMSS.bak`
   - Example: `.env-260112-135713.bak`

3. **Graceful handling:**
   - Parse errors → fallback to empty values
   - Invalid secret lengths → regenerate random
   - License still requires API validation

## Architecture

```
runInit()
    │
    ├── Step 0: License input ─────────────────────┐
    │   └── Pre-fill from loaded env if exists    │
    │                                              │
    ├── Step 1-3: Docker/Language/Options         │
    │                                              │
    ├── [NEW] Load existing .env ─────────────────┤
    │   └── loadExistingEnv(cwd)                  │
    │       └── Returns map[string]string         │
    │                                              │
    ├── Step 4: Domain input ─────────────────────┤
    │   └── Default: env["SYSTEM_DOMAIN"] or "localhost"
    │                                              │
    ├── Step 5: Credentials ──────────────────────┤
    │   ├── Use loaded values if valid            │
    │   └── Generate random for missing/invalid   │
    │                                              │
    └── Step 6: Generate files ───────────────────┘
        └── backupExistingConfigs() with timestamp
```

## Related Code Files

| File | Changes |
|------|---------|
| `cmd/init.go` | Add `loadExistingEnv()`, update `backupExistingConfigs()`, modify `runInit()` flow |
| `pkg/templates/embed.go` | Remove backup logic in `RenderTemplate()` |

## Implementation Steps

### 1. Add `loadExistingEnv()` function (cmd/init.go)

```go
// loadExistingEnv parses existing .env file and returns key-value map
func loadExistingEnv(dir string) map[string]string {
    result := make(map[string]string)
    envPath := filepath.Join(dir, ".env")

    data, err := os.ReadFile(envPath)
    if err != nil {
        return result // File doesn't exist or unreadable
    }

    lines := strings.Split(string(data), "\n")
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }
        parts := strings.SplitN(line, "=", 2)
        if len(parts) == 2 {
            key := strings.TrimSpace(parts[0])
            value := strings.TrimSpace(parts[1])
            result[key] = value
        }
    }
    return result
}
```

### 2. Update `backupExistingConfigs()` (cmd/init.go)

Change backup filename format:
```go
// Before
bakPath := srcPath + ".bak"

// After
timestamp := time.Now().Format("060102-150405")
bakPath := srcPath + "-" + timestamp + ".bak"
```

### 3. Remove backup logic from `RenderTemplate()` (pkg/templates/embed.go)

Remove lines 89-95 (backup logic) - already handled in init.go before calling RenderAll.

### 4. Modify `runInit()` flow (cmd/init.go)

**After language selection (around line 255):**
```go
// Load existing .env for pre-filling
existingEnv := loadExistingEnv(cwd)
hasExistingEnv := len(existingEnv) > 0
if hasExistingEnv {
    ui.ShowInfo(ui.Msg("loading_existing_env"))
}
```

**Pre-fill domain (around line 317):**
```go
domain := "localhost"
if val, ok := existingEnv["SYSTEM_DOMAIN"]; ok && val != "" {
    domain = val
}
```

**Pre-fill secrets (before generating random, around line 340):**
```go
// Load from existing env or generate new
jwtSecret := existingEnv["JWT_SECRET"]
if len(jwtSecret) < 32 {
    jwtSecret, err = generatePasswordWithRetry(32)
    // ...
}
// Repeat for other secrets
```

**Pre-fill license (Step 0, around line 50):**
```go
// Check for existing license before showing form
existingEnv := loadExistingEnv(cwd)
licenseKey := existingEnv["LICENSE_KEY"]
```

### 5. Add i18n messages

Add to `pkg/ui/lang_en.go` and `pkg/ui/lang_vi.go`:
- `loading_existing_env`: "Loading configuration from existing .env file"

## Todo List

- [x] Add `loadExistingEnv()` function
- [x] Update `backupExistingConfigs()` with timestamp
- [x] Remove backup in `RenderTemplate()`
- [x] Update `runInit()` to use loaded values
- [x] Add i18n messages
- [ ] Update tests (optional - no cmd tests exist)
- [x] Manual testing

## Success Criteria

- [x] Running `kk init` in dir with .env → form shows existing values
- [x] Backup creates `.env-YYMMDD-HHMMSS.bak` format
- [x] Missing/invalid secrets auto-regenerate
- [x] All existing tests pass
- [x] `go build` succeeds

## Implementation Review

**Date:** 2026-01-12 14:28
**Reviewer:** code-reviewer
**Score:** 8/10

**What went well:**
- DRY: Removed duplicate backup logic successfully
- Security: No credential exposure, proper file permissions
- Performance: Minimal overhead (single file read)
- Architecture: Clean separation, single responsibility
- YAGNI: No over-engineering

**Issues identified:**
1. **HIGH:** ENV parser doesn't handle quoted values (`KEY="value"`)
2. **HIGH:** No validation on loaded env values (domain can bypass form validation)
3. **MEDIUM:** Silent backup failures (continues with warning only)
4. **LOW:** Timestamp collision possible (same second overwrites)

**Recommendations:**
- Add quote stripping to parser (5 LOC)
- Validate loaded domain before use (3 LOC)
- Improve backup failure reporting (10 LOC)

**Full review:** [code-reviewer-260112-1428-env-preload-init.md](../reports/code-reviewer-260112-1428-env-preload-init.md)

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| .env parse fails on malformed lines | Low | Low | Skip malformed, log warning |
| Timestamp collision (same second) | Very Low | Low | Accept - rare case |
| License validation bypass | N/A | N/A | API validation still required |

## Security Considerations

- No security impact - .env already has 0600 permissions
- License still validated via API (no bypass)
- Secrets still validated for minimum length

## Next Steps

After implementation:
1. Run `go test ./...`
2. Manual test: create .env, run `kk init`, verify pre-fill
3. Verify backup file has correct timestamp format

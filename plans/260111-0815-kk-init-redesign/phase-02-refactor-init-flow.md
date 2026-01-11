---
title: "Phase 02: Refactor Init Flow"
description: "Refactor cmd/init.go to implement new 6-step flow with domain step separated and grouped credential form."
status: completed
priority: high
effort: 2 hours
branch: feature/init-refactor-phase-2
tags: [init, refactor, ui, backend]
created: 2026-01-10T10:00:00Z
completion timestamp: 2026-01-11 09:45
---

# Phase 02: Refactor Init Flow

**Effort:** 2 hours

## Objective

Refactor `cmd/init.go` để implement new 6-step flow với domain step tách riêng và grouped credential form.

## Current Flow (5 steps)

```
Step 1: Docker Check
Step 2: Language Selection
Step 3: Configuration Options (SeaweedFS, Caddy, Domain)
Step 4: Generate Files
Step 5: Complete
```

## Target Flow (6 steps)

```
Step 1: Docker Check
Step 2: Language Selection
Step 3: Service Selection (SeaweedFS, Caddy only)
Step 4: Domain Configuration
Step 5: Environment Configuration (confirm random -> grouped form)
Step 6: Generate Files + Complete
```

---

## Tasks

### 2.1 Update Step Headers

**Change step count from 5 to 6:**

```go
// Line 40
ui.ShowStepHeader(1, 6, ui.Msg("step_docker_check"))

// Line 147
ui.ShowStepHeader(2, 6, ui.Msg("step_language"))

// Line 206
ui.ShowStepHeader(3, 6, ui.Msg("step_options"))
```

### 2.2 Separate Domain from Step 3

**Current (line 233-249):** Domain asked only if Caddy enabled

**Target:** Domain asked always as separate Step 4

```go
// Step 3: Service Selection (only SeaweedFS, Caddy)
ui.ShowStepHeader(3, 6, ui.Msg("step_options"))
// ... existing SeaweedFS + Caddy form ...

// Step 4: Domain Configuration (NEW)
ui.ShowStepHeader(4, 6, ui.Msg("step_domain"))
domain := "localhost"
domainForm := huh.NewForm(
    huh.NewGroup(
        huh.NewInput().
            Title(ui.IconLink + " " + ui.Msg("enter_domain")).
            Value(&domain).
            Placeholder("localhost"),
    ),
)
if err := domainForm.Run(); err != nil {
    return err
}
if domain == "" {
    domain = "localhost"
}
```

### 2.3 Add Step 5: Environment Configuration

**New step after domain:**

```go
// Step 5: Environment Configuration
ui.ShowStepHeader(5, 6, ui.Msg("step_credentials"))

// Pre-generate all secrets
jwtSecret, _ := ui.GeneratePassword(32)
dbPass, _ := ui.GeneratePassword(24)
dbRootPass, _ := ui.GeneratePassword(24)
redisPass, _ := ui.GeneratePassword(24)
s3AccessKey, _ := generateS3AccessKey(20) // alphanumeric uppercase
s3SecretKey, _ := ui.GeneratePassword(40)

// Ask: Use random?
var useRandom bool = true
confirmForm := huh.NewForm(
    huh.NewGroup(
        huh.NewConfirm().
            Title(ui.Msg("ask_use_random")).
            Affirmative(ui.Msg("yes")).
            Negative(ui.Msg("no_edit")).
            Value(&useRandom),
    ),
)
if err := confirmForm.Run(); err != nil {
    return err
}

// If No -> Show grouped edit form
if !useRandom {
    // Build form groups
    groups := []*huh.Group{}

    // Group 1: System
    groups = append(groups, huh.NewGroup(
        huh.NewInput().
            Title("JWT_SECRET").
            Value(&jwtSecret),
    ).Title(ui.Msg("group_system")))

    // Group 2: Database Secrets
    groups = append(groups, huh.NewGroup(
        huh.NewInput().
            Title("DB_PASSWORD").
            Value(&dbPass),
        huh.NewInput().
            Title("DB_ROOT_PASSWORD").
            Value(&dbRootPass),
        huh.NewInput().
            Title("REDIS_PASSWORD").
            Value(&redisPass),
    ).Title(ui.Msg("group_db_secrets")))

    // Group 3: S3 Secrets (only if SeaweedFS enabled)
    if enableSeaweedFS {
        groups = append(groups, huh.NewGroup(
            huh.NewInput().
                Title("S3_ACCESS_KEY").
                Value(&s3AccessKey),
            huh.NewInput().
                Title("S3_SECRET_KEY").
                Value(&s3SecretKey),
        ).Title(ui.Msg("group_s3_secrets")))
    }

    editForm := huh.NewForm(groups...)
    if err := editForm.Run(); err != nil {
        return err
    }
}
```

### 2.4 Add Helper Function for S3 Access Key

```go
// generateS3AccessKey generates alphanumeric uppercase key
func generateS3AccessKey(length int) (string, error) {
    const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    result := make([]byte, length)
    for i := range result {
        idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
        if err != nil {
            return "", err
        }
        result[i] = chars[idx.Int64()]
    }
    return string(result), nil
}
```

**Note:** Add import `crypto/rand` and `math/big`

### 2.5 Update tmplCfg Construction

**Current (line 269-276):**
```go
tmplCfg := templates.Config{
    EnableSeaweedFS: enableSeaweedFS,
    EnableCaddy:     enableCaddy,
    DBPassword:      dbPass,
    DBRootPassword:  dbRootPass,
    RedisPassword:   redisPass,
    Domain:          domain,
}
```

**Target:**
```go
tmplCfg := templates.Config{
    EnableSeaweedFS: enableSeaweedFS,
    EnableCaddy:     enableCaddy,
    Domain:          domain,
    JWTSecret:       jwtSecret,
    DBPassword:      dbPass,
    DBRootPassword:  dbRootPass,
    RedisPassword:   redisPass,
    S3AccessKey:     s3AccessKey,
    S3SecretKey:     s3SecretKey,
}
```

### 2.6 Update Step 6 Header

```go
// Step 6: Generate + Complete
ui.ShowStepHeader(6, 6, ui.Msg("step_generate"))
```

---

## Validation

- Form nhóm hiển thị đúng với title
- S3 fields chỉ hiện khi EnableSeaweedFS = true
- Pre-filled values editable
- Empty validation cho required fields

## Output

- Files changed: cmd/init.go, pkg/ui/lang_en.go, pkg/ui/lang_vi.go, kk_integration_test.go
- New features: 6-step flow, --force flag, grouped credential form, domain/secret validation
- Test results: All packages passed
- Code review: 9/10

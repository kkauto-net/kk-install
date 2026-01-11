# Phase 02: Integrate License into Init Flow

## Context

- **Parent Plan:** [plan.md](plan.md)
- **Dependencies:** [Phase 01](phase-01-license-module.md)
- **Docs:** [system-architecture.md](../../docs/system-architecture.md)

## Overview

| Field | Value |
|-------|-------|
| Date | 2026-01-11 |
| Priority | P1 |
| Implementation Status | done |
| Review Status | needs_revision |
| Effort | 45m |

Integrate license verification as Step 0 in `kk init` command. Update templates to use actual license values.

## Key Insights

- Current flow: 6 steps (Docker → Language → Services → Domain → Credentials → Generate)
- New flow: 7 steps (License → Docker → Language → Services → Domain → Credentials → Generate)
- Need to update all `ui.ShowStepHeader` calls
- Store license data in `templates.Config` for rendering

## Requirements

1. Update `pkg/templates/embed.go`:
   - Add `LicenseKey` and `ServerPublicKey` fields to `Config` struct

2. Update `pkg/templates/env.tmpl`:
   - Replace hardcoded `LICENSEKEY` and `PUBLICKEY` with template vars

3. Update `cmd/init.go`:
   - Add Step 0 for license verification
   - Renumber all existing steps (1→2, 2→3, etc.)
   - Validate format before API call
   - Block on failure (no skip in force mode)

## Architecture

### Updated Init Flow

```
Step 0: License Verification [NEW]
  ├─ Prompt: Enter license key
  ├─ Validate format (regex)
  ├─ Call API
  └─ Store response in memory

Step 1: Docker Check (was Step 1)
Step 2: Language Selection (was Step 2)
Step 3: Service Selection (was Step 3)
Step 4: Domain Configuration (was Step 4)
Step 5: Credentials (was Step 5)
Step 6: Generate Files (was Step 6)
  └─ Include license data in .env
```

## Related Code Files

- [cmd/init.go:42-424](../../cmd/init.go#L42-L424) - runInit function
- [pkg/templates/embed.go:15-32](../../pkg/templates/embed.go#L15-L32) - Config struct
- [pkg/templates/env.tmpl:1-75](../../pkg/templates/env.tmpl) - Env template

## Implementation Steps

### Step 1: Update templates.Config

```go
// pkg/templates/embed.go

type Config struct {
    // Services
    EnableSeaweedFS bool
    EnableCaddy     bool

    // System
    Domain    string
    JWTSecret string

    // License [NEW]
    LicenseKey      string
    ServerPublicKey string

    // Database
    DBPassword     string
    DBRootPassword string
    RedisPassword  string

    // S3 (only used when EnableSeaweedFS)
    S3AccessKey string
    S3SecretKey string
}
```

### Step 2: Update env.tmpl

```
#--------------------------------------------------------------------
# LICENSE KKAuto
# NOTE: License is required for selfhost
# NO CHANGE THIS
#--------------------------------------------------------------------
# KKengine Configuration
KK_ENVIRONMENT=selfhost
LICENSE_KEY={{.LicenseKey}}
SERVER_PUBLIC_KEY_ENCRYPTED={{.ServerPublicKey}}
```

### Step 3: Update cmd/init.go

Add import:
```go
import (
    // ... existing imports
    "github.com/kkauto-net/kk-install/pkg/license"
)
```

Add Step 0 at beginning of `runInit`:
```go
func runInit(cmd *cobra.Command, args []string) error {
    // Command banner
    ui.ShowCommandBanner("kk init", ui.Msg("init_desc"))

    // Step 0: License Verification [NEW]
    ui.ShowStepHeader(0, 7, ui.Msg("step_license"))

    var licenseKey string
    licenseForm := huh.NewForm(
        huh.NewGroup(
            huh.NewInput().
                Title(ui.IconKey + " " + ui.Msg("enter_license")).
                Value(&licenseKey).
                Placeholder("LICENSE-XXXXXXXXXXXXXXXX").
                Validate(func(s string) error {
                    if s == "" {
                        return errors.New(ui.Msg("license_required"))
                    }
                    if !license.ValidateFormat(s) {
                        return errors.New(ui.Msg("license_invalid_format"))
                    }
                    return nil
                }),
        ),
    )
    if err := licenseForm.Run(); err != nil {
        return err
    }

    // Validate license against API
    spinner, _ := pterm.DefaultSpinner.Start(ui.IconKey + " " + ui.Msg("validating_license"))
    client := license.NewClient()
    licenseResp, err := client.Validate(licenseKey)
    if err != nil {
        spinner.Fail(ui.Msg("license_validation_failed"))
        ui.ShowBoxedError(ui.ErrorSuggestion{
            Title:      ui.Msg("license_validation_failed"),
            Message:    err.Error(),
            Suggestion: ui.Msg("license_check_key"),
        })
        return err
    }
    spinner.Success(ui.IconCheck + " " + ui.Msg("license_validated"))

    // Store license data for later use
    licenseData := struct {
        Key       string
        PublicKey string
    }{
        Key:       licenseKey,
        PublicKey: licenseResp.PublicKey,
    }

    // Step 1: Check Docker (was Step 1, update header)
    ui.ShowStepHeader(1, 7, ui.Msg("step_docker_check"))
    // ... rest unchanged

    // ... update all other step headers:
    // Step 2, 3, 4, 5 → update numbers in ShowStepHeader
    // Step 6 → now Step 7

    // ... in Step 6 (Generate Files), update tmplCfg:
    tmplCfg := templates.Config{
        // ... existing fields
        LicenseKey:      licenseData.Key,
        ServerPublicKey: licenseData.PublicKey,
    }
```

Update all `ui.ShowStepHeader` calls:
- `(1, 6, ...)` → `(1, 7, ...)`
- `(2, 6, ...)` → `(2, 7, ...)`
- etc.

## Todo List

- [x] Update `pkg/templates/embed.go` - add LicenseKey, ServerPublicKey
- [x] Update `pkg/templates/env.tmpl` - use template vars
- [x] Update `cmd/init.go` - add Step 0
- [x] Update `cmd/init.go` - renumber all steps
- [x] Update `cmd/init.go` - pass license to tmplCfg
- [x] Run `go build ./...`
- [ ] **BLOCKER:** Add i18n strings to lang_en.go and lang_vi.go
- [ ] **REQUIRED:** Handle force mode (read from env or block)
- [ ] **RECOMMENDED:** Sanitize error messages to prevent license key leakage
- [ ] Test manually: `go run . init`

## Success Criteria

- [ ] License prompt appears first
- [ ] Invalid format shows error before API call
- [ ] Invalid license blocks init
- [ ] Valid license continues to Docker check
- [ ] .env contains correct LICENSE_KEY and SERVER_PUBLIC_KEY_ENCRYPTED
- [ ] Force mode still requires license

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| Breaking existing flow | Init fails | Thorough testing |
| Step numbering mismatch | Confusing UX | Review all ShowStepHeader calls |

## Security Considerations

- License key only stored in .env (0600 permissions)
- Public key from server stored encrypted
- No license key logging

## Review Notes

**Review Date:** 2026-01-11 17:20
**Review Report:** [code-reviewer-260111-1720-phase02-init-integration.md](../reports/code-reviewer-260111-1720-phase02-init-integration.md)
**Score:** 5/10

**Blocking Issues:**
1. Missing i18n strings (step_license, enter_license, etc.) - prevents runtime execution

**Required Fixes:**
2. Force mode handling - currently not implemented
3. License key sanitization in error messages - security concern

**Status:** Needs revision before proceeding to Phase 03

## Next Steps

After fixing blocking issues and re-review approval, proceed to [Phase 03](phase-03-tests-i18n.md)

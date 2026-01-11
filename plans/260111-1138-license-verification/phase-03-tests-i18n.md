# Phase 03: Tests and i18n Messages

## Context

- **Parent Plan:** [plan.md](plan.md)
- **Dependencies:** [Phase 01](phase-01-license-module.md), [Phase 02](phase-02-init-integration.md)
- **Docs:** [code-standards.md](../../docs/code-standards.md)

## Overview

| Field | Value |
|-------|-------|
| Date | 2026-01-11 |
| Priority | P2 |
| Implementation Status | done |
| Review Status | pending |
| Effort | 30m |

Add i18n messages for license verification and update golden files.

## Requirements

1. Add i18n messages in `pkg/ui/lang_en.go` and `pkg/ui/lang_vi.go`
2. Update golden files in `pkg/templates/testdata/golden/`
3. Add icon constant for license/key

## Related Code Files

- [pkg/ui/lang_en.go](../../pkg/ui/lang_en.go)
- [pkg/ui/lang_vi.go](../../pkg/ui/lang_vi.go)
- [pkg/ui/i18n.go](../../pkg/ui/i18n.go)
- [pkg/templates/testdata/golden/env.golden](../../pkg/templates/testdata/golden/env.golden)

## Implementation Steps

### Step 1: Add icon constant

```go
// pkg/ui/i18n.go or appropriate file

const IconKey = "🔑"
```

### Step 2: Add English messages

```go
// pkg/ui/lang_en.go - add to messagesEN map

"step_license":             "License Verification",
"enter_license":            "Enter your license key",
"license_required":         "License key is required",
"license_invalid_format":   "Invalid license format. Expected: LICENSE-XXXXXXXXXXXXXXXX",
"validating_license":       "Validating license...",
"license_validated":        "License validated successfully",
"license_validation_failed": "License validation failed",
"license_check_key":        "Please check your license key and try again",
```

### Step 3: Add Vietnamese messages

```go
// pkg/ui/lang_vi.go - add to messagesVI map

"step_license":             "Xác thực License",
"enter_license":            "Nhập license key của bạn",
"license_required":         "License key là bắt buộc",
"license_invalid_format":   "Định dạng license không hợp lệ. Mong đợi: LICENSE-XXXXXXXXXXXXXXXX",
"validating_license":       "Đang xác thực license...",
"license_validated":        "Xác thực license thành công",
"license_validation_failed": "Xác thực license thất bại",
"license_check_key":        "Vui lòng kiểm tra license key và thử lại",
```

### Step 4: Update env.golden

```
#--------------------------------------------------------------------
# LICENSE KKAuto
# NOTE: License is required for selfhost
# NO CHANGE THIS
#--------------------------------------------------------------------
# KKengine Configuration
KK_ENVIRONMENT=selfhost
LICENSE_KEY=LICENSE-TESTKEY12345678
SERVER_PUBLIC_KEY_ENCRYPTED=test_public_key_encrypted
```

### Step 5: Update generate_golden.go

```go
// pkg/templates/testdata/generate_golden.go
// Update the test config to include license fields

cfg := templates.Config{
    // ... existing fields
    LicenseKey:      "LICENSE-TESTKEY12345678",
    ServerPublicKey: "test_public_key_encrypted",
}
```

## Todo List

- [ ] Add `IconKey` constant to `pkg/ui/i18n.go`
- [ ] Add English messages to `pkg/ui/lang_en.go`
- [ ] Add Vietnamese messages to `pkg/ui/lang_vi.go`
- [ ] Update `pkg/templates/testdata/golden/env.golden`
- [ ] Update `pkg/templates/testdata/generate_golden.go`
- [ ] Run `go generate ./pkg/templates/testdata/...` (if needed)
- [ ] Run `go test ./pkg/...`
- [ ] Run `golangci-lint run ./...`

## Success Criteria

- [ ] All i18n messages defined in both languages
- [ ] Golden file matches expected output with license fields
- [ ] All tests pass
- [ ] No linter warnings

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| Missing translation | Fallback to key | Review both lang files |
| Golden mismatch | Test fails | Regenerate golden files |

## Security Considerations

- No sensitive data in golden files (use test values)

## Next Steps

After completion:
1. Run full test suite: `make test`
2. Manual integration test: `go run . init`
3. Update documentation if needed

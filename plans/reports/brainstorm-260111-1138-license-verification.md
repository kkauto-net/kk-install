# Brainstorm: License Verification for kk init

**Date:** 2026-01-11
**Status:** Complete

## Problem Statement

Thêm license verification vào `kk init` command để:
- Validate license trước khi cho phép khởi tạo stack
- Lấy `public_key` từ server và lưu vào .env
- Đảm bảo chỉ licensed users mới sử dụng được

## Requirements

### API Spec
- **Endpoint:** `POST https://kkauto.net/api/license/config`
- **Body:** `{"license": "LICENSE-64ABBE22C2134D1D"}`
- **Success Response:**
```json
{
  "status": "success",
  "public_key": "<encrypted_key>",
  "message": "License configuration retrieved successfully"
}
```

### .env Output
```
LICENSE_KEY=LICENSE-64ABBE22C2134D1D
SERVER_PUBLIC_KEY_ENCRYPTED=<public_key_from_response>
```

## Decisions Made

| Aspect | Decision | Rationale |
|--------|----------|-----------|
| Step position | Step 0 (trước Docker check) | Fail fast - không waste time check docker nếu license invalid |
| Error behavior | Block hoàn toàn | License là mandatory, không có ngoại lệ |
| Force mode | Vẫn cần license | Tránh bypass, mọi init đều cần license |
| UX | Simple prompt | KISS - chỉ cần nhập license key |
| Storage | Project-level (.env) | License bound to project, không global |
| Validation | Format + Server | Check regex trước, tiết kiệm API calls |

## Proposed Solution

### New Flow (7 steps)

```
┌──────────────────────────────────────┐
│ Step 0: License Verification  [NEW]  │
│ ├─ Prompt: Enter license key         │
│ ├─ Validate format: LICENSE-XXXXX    │
│ ├─ POST https://kkauto.net/api/...   │
│ └─ Store in memory for Step 6        │
├──────────────────────────────────────┤
│ Step 1: Docker Check                 │
│ Step 2: Language Selection           │
│ Step 3: Service Selection            │
│ Step 4: Domain Configuration         │
│ Step 5: Credentials                  │
│ Step 6: Generate Files (include      │
│         license data in .env)        │
└──────────────────────────────────────┘
```

### Architecture

```
pkg/
├── license/
│   ├── license.go        # LicenseClient, Validate(), ValidateFormat()
│   └── license_test.go   # Unit tests
├── templates/
│   └── embed.go          # Add LicenseKey, ServerPublicKey to Config
```

### Implementation Details

#### 1. License Module (`pkg/license/license.go`)

```go
type LicenseClient struct {
    BaseURL    string
    HTTPClient *http.Client
}

type LicenseResponse struct {
    Status    string `json:"status"`
    PublicKey string `json:"public_key"`
    Message   string `json:"message"`
}

func (c *LicenseClient) Validate(licenseKey string) (*LicenseResponse, error)
func ValidateFormat(key string) bool // regex: LICENSE-[A-F0-9]{16}
```

#### 2. Config Update (`pkg/templates/embed.go`)

```go
type Config struct {
    // ... existing fields

    // License
    LicenseKey      string
    ServerPublicKey string
}
```

#### 3. Template Update (`pkg/templates/env.tmpl`)

```
LICENSE_KEY={{.LicenseKey}}
SERVER_PUBLIC_KEY_ENCRYPTED={{.ServerPublicKey}}
```

#### 4. Init Flow Update (`cmd/init.go`)

```go
func runInit(cmd *cobra.Command, args []string) error {
    // Step 0: License Verification [NEW]
    ui.ShowStepHeader(0, 7, ui.Msg("step_license"))

    var licenseKey string
    licenseForm := huh.NewForm(
        huh.NewGroup(
            huh.NewInput().
                Title(ui.Msg("enter_license")).
                Value(&licenseKey).
                Validate(validateLicenseFormat),
        ),
    )
    if err := licenseForm.Run(); err != nil {
        return err
    }

    // Call API
    client := license.NewClient()
    resp, err := client.Validate(licenseKey)
    if err != nil {
        ui.ShowBoxedError(...)
        return err
    }

    // Store for later use in template
    licenseConfig := LicenseConfig{
        Key:       licenseKey,
        PublicKey: resp.PublicKey,
    }

    // Continue with existing steps (renumber 1-7)
    ...
}
```

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| API downtime | Block all new installs | Clear error message + support contact |
| Network issues | User frustration | Timeout 30s + clear error |
| Invalid key format | Poor UX | Client-side validation trước |

## Success Metrics

- [ ] License validation trước Docker check
- [ ] .env có đúng LICENSE_KEY + SERVER_PUBLIC_KEY_ENCRYPTED
- [ ] Invalid license → clear error message
- [ ] Force mode vẫn require license

## Next Steps

1. Implement `pkg/license/` module
2. Update `templates.Config` struct
3. Update `env.tmpl` với actual template vars
4. Update `cmd/init.go` thêm Step 0
5. Add i18n messages cho license-related strings
6. Write tests

## Unresolved Questions

1. License format chính xác là gì? (assumed: `LICENSE-[A-F0-9]{16}`)
2. API timeout nên bao lâu? (suggested: 30s)
3. Error codes từ API khi invalid license?

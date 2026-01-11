# Phase 01: Create License Module

## Context

- **Parent Plan:** [plan.md](plan.md)
- **Dependencies:** None
- **Docs:** [code-standards.md](../../docs/code-standards.md)

## Overview

| Field | Value |
|-------|-------|
| Date | 2026-01-11 |
| Priority | P1 |
| Implementation Status | done |
| Review Status | pending |
| Effort | 45m |

Create `pkg/license/` module with HTTP client for license validation API.

## Key Insights

- Follow existing package patterns in `pkg/` directory
- Use `net/http` standard library for HTTP client
- Validate format client-side before API call
- License format: `LICENSE-[A-F0-9]{16}`

## Requirements

1. Create `pkg/license/license.go` with:
   - `LicenseClient` struct with configurable base URL
   - `LicenseResponse` struct matching API response
   - `Validate(licenseKey string) (*LicenseResponse, error)` method
   - `ValidateFormat(key string) bool` function

2. Create `pkg/license/license_test.go` with:
   - Unit tests for `ValidateFormat`
   - Integration test with mocked HTTP server for `Validate`

## Architecture

```go
package license

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "regexp"
    "time"
)

const (
    DefaultBaseURL = "https://kkauto.net"
    DefaultTimeout = 30 * time.Second
)

// LicenseResponse represents API response
type LicenseResponse struct {
    Status    string `json:"status"`
    PublicKey string `json:"public_key"`
    Message   string `json:"message"`
}

// LicenseClient handles license validation
type LicenseClient struct {
    BaseURL    string
    HTTPClient *http.Client
}

// NewClient creates default license client
func NewClient() *LicenseClient

// Validate calls API to validate license key
func (c *LicenseClient) Validate(licenseKey string) (*LicenseResponse, error)

// ValidateFormat checks license format before API call
func ValidateFormat(key string) bool
```

## Related Code Files

- `pkg/validator/docker.go` - Reference for validation patterns
- `pkg/updater/updater.go` - Reference for HTTP client usage

## Implementation Steps

### Step 1: Create license.go

```go
// pkg/license/license.go

package license

import (
    "bytes"
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
    "regexp"
    "time"
)

const (
    DefaultBaseURL = "https://kkauto.net"
    DefaultTimeout = 30 * time.Second
)

var licenseFormatRegex = regexp.MustCompile(`^LICENSE-[A-F0-9]{16}$`)

type LicenseResponse struct {
    Status    string `json:"status"`
    PublicKey string `json:"public_key"`
    Message   string `json:"message"`
}

type LicenseClient struct {
    BaseURL    string
    HTTPClient *http.Client
}

func NewClient() *LicenseClient {
    return &LicenseClient{
        BaseURL: DefaultBaseURL,
        HTTPClient: &http.Client{
            Timeout: DefaultTimeout,
        },
    }
}

func ValidateFormat(key string) bool {
    return licenseFormatRegex.MatchString(key)
}

func (c *LicenseClient) Validate(licenseKey string) (*LicenseResponse, error) {
    if !ValidateFormat(licenseKey) {
        return nil, errors.New("invalid license format")
    }

    url := c.BaseURL + "/api/license/config"
    body := map[string]string{"license": licenseKey}
    jsonBody, err := json.Marshal(body)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal request: %w", err)
    }

    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to call license API: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("license API returned status %d", resp.StatusCode)
    }

    var result LicenseResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }

    if result.Status != "success" {
        return nil, errors.New(result.Message)
    }

    return &result, nil
}
```

### Step 2: Create license_test.go

```go
// pkg/license/license_test.go

package license

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestValidateFormat(t *testing.T) {
    tests := []struct {
        name     string
        key      string
        expected bool
    }{
        {"valid format", "LICENSE-64ABBE22C2134D1D", true},
        {"lowercase invalid", "license-64abbe22c2134d1d", false},
        {"missing prefix", "64ABBE22C2134D1D", false},
        {"wrong prefix", "LIC-64ABBE22C2134D1D", false},
        {"too short", "LICENSE-64ABBE22C213", false},
        {"too long", "LICENSE-64ABBE22C2134D1D1", false},
        {"empty", "", false},
        {"with spaces", "LICENSE-64ABBE22 C2134D1D", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := ValidateFormat(tt.key)
            assert.Equal(t, tt.expected, result)
        })
    }
}

func TestValidate_Success(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        assert.Equal(t, "POST", r.Method)
        assert.Equal(t, "/api/license/config", r.URL.Path)

        resp := LicenseResponse{
            Status:    "success",
            PublicKey: "test_public_key",
            Message:   "License configuration retrieved successfully",
        }
        json.NewEncoder(w).Encode(resp)
    }))
    defer server.Close()

    client := &LicenseClient{
        BaseURL:    server.URL,
        HTTPClient: http.DefaultClient,
    }

    result, err := client.Validate("LICENSE-64ABBE22C2134D1D")
    require.NoError(t, err)
    assert.Equal(t, "success", result.Status)
    assert.Equal(t, "test_public_key", result.PublicKey)
}

func TestValidate_InvalidLicense(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        resp := LicenseResponse{
            Status:  "error",
            Message: "Invalid license key",
        }
        json.NewEncoder(w).Encode(resp)
    }))
    defer server.Close()

    client := &LicenseClient{
        BaseURL:    server.URL,
        HTTPClient: http.DefaultClient,
    }

    _, err := client.Validate("LICENSE-64ABBE22C2134D1D")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "Invalid license key")
}
```

## Todo List

- [ ] Create `pkg/license/license.go`
- [ ] Create `pkg/license/license_test.go`
- [ ] Run `go test ./pkg/license/...`
- [ ] Run `golangci-lint run ./pkg/license/...`

## Success Criteria

- [ ] `ValidateFormat` correctly validates license format
- [ ] `Validate` calls API and parses response
- [ ] All tests pass
- [ ] No linter warnings

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| API timeout | User frustration | 30s timeout, clear error message |
| Format changes | Validation fails | Regex in const, easy to update |

## Security Considerations

- HTTP client uses HTTPS only
- No sensitive data logged
- Timeout prevents hanging

## Next Steps

After completion, proceed to [Phase 02](phase-02-init-integration.md)

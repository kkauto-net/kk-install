package license

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"
)

const (
	DefaultBaseURL     = "https://kkauto.net"
	DefaultTimeout     = 30 * time.Second
	maxResponseBodySize = 1 << 20 // 1MB limit for response body
)

var licenseFormatRegex = regexp.MustCompile(`^LICENSE-[A-F0-9]{16}$`)

// LicenseResponse represents API response from license server
type LicenseResponse struct {
	Status    string `json:"status"`
	PublicKey string `json:"public_key"`
	Message   string `json:"message"`
}

// LicenseClient handles license validation against remote API
type LicenseClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates default license client with production URL
func NewClient() *LicenseClient {
	return &LicenseClient{
		BaseURL: DefaultBaseURL,
		HTTPClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

// ValidateFormat checks license key format before API call
// Expected format: LICENSE-[A-F0-9]{16}
func ValidateFormat(key string) bool {
	return licenseFormatRegex.MatchString(key)
}

// Validate calls remote API to validate license key
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
	limitedBody := io.LimitReader(resp.Body, maxResponseBodySize)
	if err := json.NewDecoder(limitedBody).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Status != "success" {
		return nil, errors.New(result.Message)
	}

	return &result, nil
}

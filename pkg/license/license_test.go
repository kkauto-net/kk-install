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
		{"valid second key", "LICENSE-ABCDEF0123456789", true},
		{"mixed case hex invalid", "LICENSE-64aBbE22C2134D1D", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateFormat(tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewClient(t *testing.T) {
	client := NewClient()
	assert.Equal(t, DefaultBaseURL, client.BaseURL)
	assert.NotNil(t, client.HTTPClient)
	assert.Equal(t, DefaultTimeout, client.HTTPClient.Timeout)
}

func TestValidate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/license/config", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var reqBody map[string]string
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		require.NoError(t, err)
		assert.Equal(t, "LICENSE-64ABBE22C2134D1D", reqBody["license"])

		resp := LicenseResponse{
			Status:    "success",
			PublicKey: "test_public_key_encrypted",
			Message:   "License configuration retrieved successfully",
		}
		w.Header().Set("Content-Type", "application/json")
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
	assert.Equal(t, "test_public_key_encrypted", result.PublicKey)
	assert.Equal(t, "License configuration retrieved successfully", result.Message)
}

func TestValidate_InvalidFormat(t *testing.T) {
	client := NewClient()
	_, err := client.Validate("invalid-key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid license format")
}

func TestValidate_InvalidLicense(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := LicenseResponse{
			Status:  "error",
			Message: "Invalid license key",
		}
		w.Header().Set("Content-Type", "application/json")
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

func TestValidate_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := &LicenseClient{
		BaseURL:    server.URL,
		HTTPClient: http.DefaultClient,
	}

	_, err := client.Validate("LICENSE-64ABBE22C2134D1D")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "license API returned status 500")
}

func TestValidate_NetworkError(t *testing.T) {
	client := &LicenseClient{
		BaseURL:    "http://localhost:99999",
		HTTPClient: http.DefaultClient,
	}

	_, err := client.Validate("LICENSE-64ABBE22C2134D1D")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to call license API")
}

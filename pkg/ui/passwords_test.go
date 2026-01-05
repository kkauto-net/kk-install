package ui

import (
	"regexp"
	"testing"
)

func TestGeneratePassword(t *testing.T) {
	testCases := []struct {
		name      string
		length    int
		wantError bool
	}{
		{"Valid length 16", 16, false},
		{"Valid length 32", 32, false},
		{"Length 0", 0, false}, // Should return empty string, no error
		{"Length 1", 1, false},
		// crypto/rand.Read might return error for very large lengths, but it's not expected for typical password lengths.
		// base64.RawURLEncoding.EncodeToString will panic for negative length, but it's handled by make([]byte, length) which panics earlier.
	}

	urlSafeRegex := regexp.MustCompile("^[a-zA-Z0-9_-]*$")

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			password, err := GeneratePassword(tc.length)

			if (err != nil) != tc.wantError {
				t.Fatalf("GeneratePassword() error = %v, wantError %v", err, tc.wantError)
			}

			if !tc.wantError {
				if len(password) != tc.length {
					t.Errorf("GeneratePassword() generated password length = %v, want %v", len(password), tc.length)
				}
				if !urlSafeRegex.MatchString(password) {
					t.Errorf("GeneratePassword() generated password contains non-URL-safe characters: %v", password)
				}
			}
		})
	}
}

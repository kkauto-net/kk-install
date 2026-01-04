package ui

import (
	"crypto/rand"
	"encoding/base64"
)

// GeneratePassword creates cryptographically secure random password
func GeneratePassword(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	// Use URL-safe base64, no special chars that might break shell
	return base64.RawURLEncoding.EncodeToString(bytes)[:length], nil
}

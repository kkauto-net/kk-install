package ui

import (
	"errors"
	"strings"
	"testing"
)

func TestSanitizeError(t *testing.T) {
	longMsg := strings.Repeat("x", 300)
	err := errors.New(longMsg)
	result := SanitizeError(err)
	if len(result) > maxSanitizedErrorLen {
		t.Errorf("expected sanitized length <= %d, got %d", maxSanitizedErrorLen, len(result))
	}
	if !strings.HasSuffix(result, "...") {
		t.Errorf("expected truncated suffix, got %q", result)
	}
}

func TestShowWarningfUsesMsg(t *testing.T) {
	original := GetLanguage()
	defer SetLanguage(original)

	SetLanguage(LangEN)
	msg := MsgF("warn_weak_password", "DB_PASSWORD")
	if !strings.Contains(msg, "DB_PASSWORD") {
		t.Errorf("expected formatted warning message, got %q", msg)
	}
}

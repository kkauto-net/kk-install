package validator

import (
	"errors"
	"strings"
	"testing"
)

func TestTranslateError(t *testing.T) {
	t.Run("UserError translation", func(t *testing.T) {
		err := &UserError{
			Key:        "test_error",
			Message:    "Test error message",
			Suggestion: "Try this fix",
		}

		result := TranslateError(err)
		if !strings.Contains(result, "Test error message") {
			t.Errorf("Expected message in result, got %q", result)
		}
		if !strings.Contains(result, "Try this fix") {
			t.Errorf("Expected suggestion in result, got %q", result)
		}
	})

	t.Run("Generic error translation", func(t *testing.T) {
		err := errors.New("generic error")
		result := TranslateError(err)
		if !strings.Contains(result, "Loi:") {
			t.Errorf("Expected 'Loi:' prefix, got %q", result)
		}
		if !strings.Contains(result, "generic error") {
			t.Errorf("Expected error message in result, got %q", result)
		}
	})
}

func TestUserError(t *testing.T) {
	t.Run("Error with suggestion", func(t *testing.T) {
		err := &UserError{
			Key:        "test",
			Message:    "Error occurred",
			Suggestion: "Fix it",
		}

		expected := "Error occurred - Fix it"
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})

	t.Run("Error without suggestion", func(t *testing.T) {
		err := &UserError{
			Key:     "test",
			Message: "Error occurred",
		}

		expected := "Error occurred"
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err.Error())
		}
	})
}

func TestErrorMessages(t *testing.T) {
	expectedKeys := []string{
		ErrDockerNotInstalled,
		ErrDockerNotRunning,
		ErrPortConflict,
		ErrEnvMissing,
		ErrEnvMissingVars,
		ErrComposeMissing,
		ErrComposeSyntax,
		ErrDiskLow,
	}

	for _, key := range expectedKeys {
		if msg, ok := ErrorMessages[key]; !ok {
			t.Errorf("Error message not defined for key %q", key)
		} else {
			if msg.Message == "" {
				t.Errorf("Empty message for key %q", key)
			}
			if msg.Suggestion == "" {
				t.Errorf("Empty suggestion for key %q", key)
			}
		}
	}
}

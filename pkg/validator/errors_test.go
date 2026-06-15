package validator

import (
	"errors"
	"strings"
	"testing"

	"github.com/kkauto-net/kk-install/pkg/ui"
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

func TestErrorSuggestionKeys(t *testing.T) {
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
		suggestionKey, ok := errorSuggestionKeys[key]
		if !ok {
			t.Errorf("Suggestion key not defined for %q", key)
			continue
		}
		if ui.Msg(suggestionKey) == suggestionKey {
			t.Errorf("Missing i18n suggestion for key %q", suggestionKey)
		}
	}
}

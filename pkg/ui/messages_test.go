package ui

import (
	"fmt"
	"testing"
)

func TestMessageFunctions(t *testing.T) {
	// Save original state
	original := GetLanguage()
	defer SetLanguage(original)

	// Test with default language (English)
	SetLanguage(LangEN)
	tests := []struct {
		name     string
		function func() string
		expected string
	}{
		{"MsgCheckingDocker", MsgCheckingDocker, "Checking Docker..."},
		{"MsgDockerOK", MsgDockerOK, "Docker is ready"},
		{"MsgInitComplete", MsgInitComplete, "Initialization complete!"},
		{"MsgDockerNotInstalled", MsgDockerNotInstalled, "Docker is not installed"},
		{"MsgDockerNotRunning", MsgDockerNotRunning, "Docker daemon is not running"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.function(); got != tt.expected {
				t.Errorf("%s() = %q, want %q", tt.name, got, tt.expected)
			}
		})
	}
}

func TestMsgCreated(t *testing.T) {
	// Save original state
	original := GetLanguage()
	defer SetLanguage(original)

	// Test with default language (English)
	SetLanguage(LangEN)
	fileName := "docker-compose.yml"
	expected := fmt.Sprintf("Created: %s", fileName)
	if got := MsgCreated(fileName); got != expected {
		t.Errorf("MsgCreated(%q) = %q, want %q", fileName, got, expected)
	}

	emptyFileName := ""
	expectedEmpty := fmt.Sprintf("Created: %s", emptyFileName)
	if got := MsgCreated(emptyFileName); got != expectedEmpty {
		t.Errorf("MsgCreated(%q) = %q, want %q", emptyFileName, got, expectedEmpty)
	}
}

func TestMsgNextSteps(t *testing.T) {
	// Save original state
	original := GetLanguage()
	defer SetLanguage(original)

	// Test with default language (English)
	SetLanguage(LangEN)
	expected := `
Next steps:
  1. Review and edit .env if needed
  2. Run: kk start
`
	if got := MsgNextSteps(); got != expected {
		t.Errorf("MsgNextSteps() = %q, want %q", got, expected)
	}
}

// Test for ShowSuccess, ShowError, ShowInfo, ShowWarning are omitted
// because they interact with stdout/stderr via pterm and are difficult to test
// without mocking pterm or redirecting output, which is out of scope for
// basic unit tests of string messages.

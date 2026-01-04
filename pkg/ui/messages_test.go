package ui

import (
	"fmt"
	"testing"
)

func TestMessageFunctions(t *testing.T) {
	tests := []struct {
		name     string
		function func() string
		expected string
	}{
		{"MsgCheckingDocker", MsgCheckingDocker, "Dang kiem tra Docker..."},
		{"MsgDockerOK", MsgDockerOK, "Docker da san sang"},
		{"MsgInitComplete", MsgInitComplete, "Khoi tao hoan tat!"},
		{"MsgDockerNotInstalled", MsgDockerNotInstalled, "Docker chua cai dat"},
		{"MsgDockerNotRunning", MsgDockerNotRunning, "Docker daemon khong chay"},
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
	fileName := "docker-compose.yml"
	expected := fmt.Sprintf("Da tao: %s", fileName)
	if got := MsgCreated(fileName); got != expected {
		t.Errorf("MsgCreated(%q) = %q, want %q", fileName, got, expected)
	}

	emptyFileName := ""
	expectedEmpty := fmt.Sprintf("Da tao: %s", emptyFileName)
	if got := MsgCreated(emptyFileName); got != expectedEmpty {
		t.Errorf("MsgCreated(%q) = %q, want %q", emptyFileName, got, expectedEmpty)
	}
}

func TestMsgNextSteps(t *testing.T) {
	expected := `
Buoc tiep theo:
  1. Kiem tra va chinh sua .env neu can
  2. Chay: kk start
`
	if got := MsgNextSteps(); got != expected {
		t.Errorf("MsgNextSteps() = %q, want %q", got, expected)
	}
}

// Test for ShowSuccess, ShowError, ShowInfo, ShowWarning are omitted
// because they interact with stdout/stderr via pterm and are difficult to test
// without mocking pterm or redirecting output, which is out of scope for
// basic unit tests of string messages.

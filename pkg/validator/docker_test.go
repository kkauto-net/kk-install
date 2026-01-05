package validator

import (
	"context"
	"os"
	"os/exec"
	"testing"
)

// Define mock functions that return pre-defined values
func mockLookPath(file string) (string, error) {
	if file == "docker" {
		return "/usr/bin/docker", nil
	}
	return "", os.ErrNotExist
}

func mockLookPathNotFound(file string) (string, error) {
	return "", os.ErrNotExist
}

func mockCommandContextSuccess(ctx context.Context, name string, arg ...string) *exec.Cmd {
	cmd := exec.Command("true") // 'true' is a Unix command that always exits with zero status
	return cmd
}

func mockCommandContextFailure(ctx context.Context, name string, arg ...string) *exec.Cmd {
	cmd := exec.Command("false") // 'false' is a Unix command that always exits with non-zero status
	return cmd
}

func TestDockerValidator_CheckDockerInstalled(t *testing.T) {
	// Test case 1: Docker is installed
	vInstalled := &DockerValidator{LookPath: mockLookPath, CommandContext: mockCommandContextSuccess}
	err := vInstalled.CheckDockerInstalled()
	if err != nil {
		t.Errorf("CheckDockerInstalled() failed when Docker is simulated as installed: %v", err)
	}

	// Test case 2: Docker is NOT installed
	vNotInstalled := &DockerValidator{LookPath: mockLookPathNotFound, CommandContext: mockCommandContextFailure}
	err = vNotInstalled.CheckDockerInstalled()
	if err == nil {
		t.Errorf("CheckDockerInstalled() did not return an error when Docker is simulated as not installed")
	}
	userErr, ok := err.(*UserError)
	if !ok {
		t.Errorf("CheckDockerInstalled() returned error of unexpected type: %T, want *UserError", err)
	}
	if userErr.Key != "docker_not_installed" {
		t.Errorf("UserError Key mismatch. Got: %q, Want: %q", userErr.Key, "docker_not_installed")
	}
}

func TestDockerValidator_CheckDockerDaemon(t *testing.T) {
	// Test case 1: Docker daemon is running
	vDaemonRunning := &DockerValidator{LookPath: mockLookPath, CommandContext: mockCommandContextSuccess}
	err := vDaemonRunning.CheckDockerDaemon()
	if err != nil {
		t.Errorf("CheckDockerDaemon() failed when Docker daemon is simulated as running: %v", err)
	}

	// Test case 2: Docker daemon is NOT running
	vDaemonNotRunning := &DockerValidator{LookPath: mockLookPath, CommandContext: mockCommandContextFailure}
	err = vDaemonNotRunning.CheckDockerDaemon()
	if err == nil {
		t.Errorf("CheckDockerDaemon() did not return an error when Docker daemon is simulated as not running")
	}
	userErr, ok := err.(*UserError)
	if !ok {
		t.Errorf("CheckDockerDaemon() returned error of unexpected type: %T, want *UserError", err)
	}
	if userErr.Key != "docker_not_running" {
		t.Errorf("UserError Key mismatch. Got: %q, Want: %q", userErr.Key, "docker_not_running")
	}
}

func TestUserError_Error(t *testing.T) {
	// Test case 1: Error with suggestion
	err1 := &UserError{
		Key:        "test_key",
		Message:    "Test message",
		Suggestion: "Test suggestion",
	}
	expected1 := "Test message - Test suggestion"
	if err1.Error() != expected1 {
		t.Errorf("UserError.Error() mismatch. Got: %q, Want: %q", err1.Error(), expected1)
	}

	// Test case 2: Error without suggestion
	err2 := &UserError{
		Key:     "test_key_no_suggestion",
		Message: "Another test message",
	}
	expected2 := "Another test message"
	if err2.Error() != expected2 {
		t.Errorf("UserError.Error() mismatch. Got: %q, Want: %q", err2.Error(), expected2)
	}
}

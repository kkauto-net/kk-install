package validator

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// Define mock functions that return pre-defined values
func mockLookPath(file string) (string, error) {
	if file == "docker" {
		return "/usr/bin/docker", nil
	}
	if file == "curl" || file == "sudo" {
		return "/usr/bin/" + file, nil
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

func TestEnsureDockerReadyAutoFixInstallsWhenMissing(t *testing.T) {
	installCalls := 0
	v := &DockerValidator{
		LookPath: func(file string) (string, error) {
			switch file {
			case "docker":
				if installCalls > 0 {
					return "/usr/bin/docker", nil
				}
				return "", os.ErrNotExist
			case "curl", "sudo":
				return "/usr/bin/" + file, nil
			default:
				return "", os.ErrNotExist
			}
		},
		CommandContext: func(ctx context.Context, name string, arg ...string) *exec.Cmd {
			joined := strings.Join(append([]string{name}, arg...), " ")
			if strings.Contains(joined, "get.docker.com") {
				installCalls++
				return exec.Command("true")
			}
			if strings.Contains(joined, "usermod") || strings.Contains(joined, "sg docker") {
				return exec.Command("true")
			}
			return exec.Command("true")
		},
	}

	err := v.EnsureDockerReady(EnsureDockerOptions{AutoFix: true, MaxRetries: 0})
	if err != nil {
		t.Fatalf("EnsureDockerReady() error = %v", err)
	}
	if installCalls != 1 {
		t.Fatalf("installCalls = %d, want 1", installCalls)
	}
}

func TestEnsureDockerReadyWithoutAutoFixReturnsMissingDocker(t *testing.T) {
	v := &DockerValidator{
		LookPath:       mockLookPathNotFound,
		CommandContext: mockCommandContextFailure,
	}

	err := v.EnsureDockerReady(EnsureDockerOptions{})
	if err == nil {
		t.Fatal("EnsureDockerReady() expected error")
	}
	if UserErrorKey(err) != "docker_not_installed" {
		t.Fatalf("UserErrorKey() = %q, want docker_not_installed", UserErrorKey(err))
	}
}

func TestEnsureDockerReadyConfirmInstallDeclined(t *testing.T) {
	v := &DockerValidator{
		LookPath:       mockLookPathNotFound,
		CommandContext: mockCommandContextFailure,
	}

	err := v.EnsureDockerReady(EnsureDockerOptions{
		ConfirmInstall: func() (bool, error) { return false, nil },
	})
	if err == nil {
		t.Fatal("EnsureDockerReady() expected error")
	}
	if UserErrorKey(err) != "docker_not_installed" {
		t.Fatalf("UserErrorKey() = %q, want docker_not_installed", UserErrorKey(err))
	}
}

func TestEnsureDockerReadyAutoFixStartsDaemon(t *testing.T) {
	startCalls := 0
	v := &DockerValidator{
		LookPath: mockLookPath,
		CommandContext: func(ctx context.Context, name string, arg ...string) *exec.Cmd {
			joined := strings.Join(append([]string{name}, arg...), " ")
			if strings.Contains(joined, "docker info") {
				if startCalls > 0 {
					return exec.Command("true")
				}
				return exec.Command("sh", "-c", "echo cannot connect >&2; exit 1")
			}
			if strings.Contains(joined, "systemctl start docker") || strings.Contains(joined, "service docker start") {
				startCalls++
				return exec.Command("true")
			}
			if strings.Contains(joined, "compose") && strings.Contains(joined, "version") {
				return exec.Command("sh", "-c", "printf '2.30.0'")
			}
			return exec.Command("true")
		},
	}

	err := v.EnsureDockerReady(EnsureDockerOptions{AutoFix: true, MaxRetries: 0})
	if err != nil {
		t.Fatalf("EnsureDockerReady() error = %v", err)
	}
	if startCalls < 1 {
		t.Fatalf("startCalls = %d, want >= 1", startCalls)
	}
}

func TestEnsureDockerReadyReturnsPermissionPendingWhenDaemonRunning(t *testing.T) {
	startCalls := 0
	v := &DockerValidator{
		LookPath: mockLookPath,
		CommandContext: func(ctx context.Context, name string, arg ...string) *exec.Cmd {
			joined := strings.Join(append([]string{name}, arg...), " ")
			if name == "docker" && len(arg) > 0 && arg[0] == "info" {
				return exec.Command("sh", "-c", "echo permission denied >&2; exit 1")
			}
			if strings.Contains(joined, "sudo") && strings.Contains(joined, "docker info") {
				return exec.Command("true")
			}
			if strings.Contains(joined, "usermod") {
				return exec.Command("true")
			}
			if strings.Contains(joined, "sg docker") {
				return exec.Command("false")
			}
			if strings.Contains(joined, "systemctl start docker") || strings.Contains(joined, "service docker start") {
				startCalls++
			}
			if strings.Contains(joined, "compose") && strings.Contains(joined, "version") {
				return exec.Command("sh", "-c", "printf '2.30.0'")
			}
			return exec.Command("true")
		},
	}

	err := v.EnsureDockerReady(EnsureDockerOptions{AutoFix: true, MaxRetries: 0})
	if UserErrorKey(err) != "docker_permission_not_effective" {
		t.Fatalf("UserErrorKey() = %q, want docker_permission_not_effective", UserErrorKey(err))
	}
	if startCalls != 0 {
		t.Fatalf("startCalls = %d, want 0 when daemon is already running", startCalls)
	}
}

func TestUserErrorKey(t *testing.T) {
	if got := UserErrorKey(nil); got != "" {
		t.Fatalf("UserErrorKey(nil) = %q, want empty", got)
	}
	err := &UserError{Key: "docker_not_running"}
	if got := UserErrorKey(err); got != "docker_not_running" {
		t.Fatalf("UserErrorKey() = %q, want docker_not_running", got)
	}
}

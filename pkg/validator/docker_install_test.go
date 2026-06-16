package validator

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kkauto-net/kk-install/pkg/ui"
)

func TestClassifyDockerInstallFailure(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		runErr  error
		wantKey string
	}{
		{name: "sudo password", output: "sudo: a password is required", wantKey: "docker_install_err_sudo_password"},
		{name: "sudo timed out", output: "sudo: timed out", wantKey: "docker_install_err_sudo_password"},
		{name: "network", output: "curl: (6) Could not resolve host get.docker.com", wantKey: "docker_install_err_network"},
		{name: "apt lock", output: "E: Could not get lock /var/lib/dpkg/lock", wantKey: "docker_install_err_pkg_lock"},
		{name: "timeout context", output: "", runErr: context.DeadlineExceeded, wantKey: "docker_install_err_timeout"},
		{name: "timeout curl", output: "curl: (28) Operation timed out", wantKey: "docker_install_err_timeout"},
		{name: "generic", output: "unknown install failure", wantKey: "docker_install_err_generic"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := classifyDockerInstallFailure(tt.output, tt.runErr)
			if err.Key != tt.wantKey {
				t.Fatalf("classifyDockerInstallFailure() key = %q, want %q", err.Key, tt.wantKey)
			}
		})
	}
}

func TestInstallDockerMissingPrerequisites(t *testing.T) {
	v := &DockerValidator{
		LookPath: func(file string) (string, error) {
			if file == "docker" {
				return "/usr/bin/docker", nil
			}
			return "", os.ErrNotExist
		},
		CommandContext: mockCommandContextSuccess,
	}

	err := v.InstallDocker()
	if err == nil {
		t.Fatal("InstallDocker() expected prerequisite error")
	}
	if UserErrorKey(err) != "docker_install_err_curl_missing" {
		t.Fatalf("UserErrorKey() = %q, want docker_install_err_curl_missing", UserErrorKey(err))
	}
}

func TestUserErrorMessageLocalized(t *testing.T) {
	original := ui.GetLanguage()
	defer ui.SetLanguage(original)

	ui.SetLanguage(ui.LangVI)
	err := &UserError{Key: "docker_install_err_network"}
	if !strings.Contains(UserErrorMessage(err), "Không thể tải") {
		t.Fatalf("UserErrorMessage() = %q, want Vietnamese network message", UserErrorMessage(err))
	}
}

func TestFormatUserErrorForBoxIncludesDetail(t *testing.T) {
	err := &UserError{
		Key:    "docker_install_err_network",
		Detail: "line1\nline2\nline3\nline4",
	}
	box := FormatUserErrorForBox(err)
	if !strings.Contains(box, "line4") {
		t.Fatalf("FormatUserErrorForBox() = %q, want trimmed detail", box)
	}
}

func TestIsDockerInstallError(t *testing.T) {
	if !IsDockerInstallError(&UserError{Key: "docker_install_err_timeout"}) {
		t.Fatal("expected docker install sub-key to match")
	}
	if IsDockerInstallError(&UserError{Key: "docker_not_running"}) {
		t.Fatal("docker_not_running should not match install error")
	}
}

func TestDockerInstallTimeoutFromEnv(t *testing.T) {
	t.Setenv("KK_DOCKER_INSTALL_TIMEOUT", "10m")
	if got := dockerInstallTimeout(); got != 10*time.Minute {
		t.Fatalf("dockerInstallTimeout() = %v, want 10m", got)
	}
}

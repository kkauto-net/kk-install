package validator

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestDockerGroupReexecCommandPrefersSg(t *testing.T) {
	v := &DockerValidator{
		LookPath: func(file string) (string, error) {
			if file == "sg" || file == "newgrp" {
				return "/usr/bin/" + file, nil
			}
			return "", os.ErrNotExist
		},
	}

	got := v.DockerGroupReexecCommand("kk init")
	if !strings.HasPrefix(got, `sg docker -c`) {
		t.Fatalf("DockerGroupReexecCommand() = %q, want sg docker prefix", got)
	}
}

func TestDockerGroupReexecCommandFallsBackToNewgrp(t *testing.T) {
	v := &DockerValidator{
		LookPath: func(file string) (string, error) {
			if file == "newgrp" {
				return "/usr/bin/newgrp", nil
			}
			return "", os.ErrNotExist
		},
	}

	got := v.DockerGroupReexecCommand("kk init")
	if got != `newgrp docker -c "kk init"` {
		t.Fatalf("DockerGroupReexecCommand() = %q, want newgrp fallback", got)
	}
}

func TestIsUserInDockerGroup(t *testing.T) {
	v := &DockerValidator{
		LookPath: mockLookPath,
		CommandContext: func(ctx context.Context, name string, arg ...string) *exec.Cmd {
			if name == "getent" {
				return exec.Command("sh", "-c", "printf 'docker:x:999:tieutinh'")
			}
			return exec.Command("true")
		},
	}

	t.Setenv("USER", "tieutinh")
	if !v.isUserInDockerGroup() {
		t.Fatal("isUserInDockerGroup() = false, want true")
	}
}

func TestDockerGroupReexecCommandReloginWhenNoRunner(t *testing.T) {
	v := &DockerValidator{
		LookPath: mockLookPathNotFound,
	}

	got := v.DockerGroupReexecCommand("kk init")
	if got == `newgrp docker -c "kk init"` || got == "" {
		t.Fatalf("DockerGroupReexecCommand() = %q, want relogin command", got)
	}
}

func TestHasDockerGroupRunner(t *testing.T) {
	withSG := &DockerValidator{
		LookPath: func(file string) (string, error) {
			if file == "sg" {
				return "/usr/bin/sg", nil
			}
			return "", os.ErrNotExist
		},
	}
	if !withSG.HasDockerGroupRunner() {
		t.Fatal("HasDockerGroupRunner() = false, want true")
	}
	if (&DockerValidator{LookPath: mockLookPathNotFound}).HasDockerGroupRunner() {
		t.Fatal("HasDockerGroupRunner() = true, want false")
	}
}

func TestTryActivateDockerSudoFallback(t *testing.T) {
	t.Setenv("KK_DOCKER_SUDO", "")
	called := false
	v := &DockerValidator{
		LookPath: mockLookPathNotFound,
		CommandContext: func(ctx context.Context, name string, arg ...string) *exec.Cmd {
			joined := strings.Join(append([]string{name}, arg...), " ")
			if strings.Contains(joined, "getent group docker") {
				return exec.Command("sh", "-c", "printf 'docker:x:999:tieutinh'")
			}
			if strings.Contains(joined, "sudo -n docker info") || strings.Contains(joined, "sudo docker info") {
				called = true
				return exec.Command("true")
			}
			return exec.Command("false")
		},
	}
	t.Setenv("USER", "tieutinh")

	if !v.TryActivateDockerSudoFallback() {
		t.Fatal("TryActivateDockerSudoFallback() = false, want true")
	}
	if !called {
		t.Fatal("expected sudo docker info probe")
	}
	if os.Getenv("KK_DOCKER_SUDO") != "1" {
		t.Fatal("expected KK_DOCKER_SUDO=1 to be set")
	}
}

func TestRunCommandWithDockerGroupMissingRunner(t *testing.T) {
	v := &DockerValidator{
		LookPath: mockLookPathNotFound,
	}
	if err := v.RunCommandWithDockerGroup("kk init", os.Environ()); err == nil {
		t.Fatal("RunCommandWithDockerGroup() expected error when sg/newgrp missing")
	}
}

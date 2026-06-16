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

func TestRunCommandWithDockerGroupMissingRunner(t *testing.T) {
	v := &DockerValidator{
		LookPath: mockLookPathNotFound,
	}
	if err := v.RunCommandWithDockerGroup("kk init", os.Environ()); err == nil {
		t.Fatal("RunCommandWithDockerGroup() expected error when sg/newgrp missing")
	}
}

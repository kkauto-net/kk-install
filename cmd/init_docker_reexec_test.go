package cmd

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"testing"

	"github.com/kkauto-net/kk-install/pkg/validator"
)

func TestBuildInitReexecCommand(t *testing.T) {
	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	os.Args = []string{"/usr/bin/kk", "init", "--domain", "example.com"}

	command, err := buildInitReexecCommand()
	if err != nil {
		t.Fatalf("buildInitReexecCommand() error = %v", err)
	}
	if command == "" {
		t.Fatal("buildInitReexecCommand() returned empty command")
	}
}

func TestShouldAttemptDockerGroupReexecSkipsWhenAlreadyReexeced(t *testing.T) {
	t.Setenv(dockerGroupReexecEnv, "1")
	err := &validator.UserError{Key: "docker_permission_not_effective"}
	if shouldAttemptDockerGroupReexec(err) {
		t.Fatal("expected false when KK_DOCKER_GROUP_REEXEC is set")
	}
}

func TestShouldAttemptDockerGroupReexecSkipsUnrelatedErrors(t *testing.T) {
	t.Setenv(dockerGroupReexecEnv, "")
	err := &validator.UserError{Key: "docker_not_installed"}
	if shouldAttemptDockerGroupReexec(err) {
		t.Fatal("expected false for unrelated docker errors")
	}
}

func TestConsumeReexecLicenseEnv(t *testing.T) {
	t.Setenv(dockerGroupReexecEnv, "1")
	t.Setenv(initValidatedLicenseEnv, "LICENSE-ABCDEF0123456789")
	t.Setenv(initValidatedLicensePubEnv, "PUBKEY")

	key, pub, ok := consumeReexecLicenseEnv()
	if !ok {
		t.Fatal("consumeReexecLicenseEnv() expected ok")
	}
	if key != "LICENSE-ABCDEF0123456789" || pub != "PUBKEY" {
		t.Fatalf("consumeReexecLicenseEnv() = (%q, %q)", key, pub)
	}
	if os.Getenv(initValidatedLicenseEnv) != "" {
		t.Fatal("expected license env to be unset after consume")
	}
}

func TestTryReexecInitWithDockerGroupReturnsOriginalErrorWhenSGUnavailable(t *testing.T) {
	t.Setenv(dockerGroupReexecEnv, "")

	DockerValidatorInstance = &validator.DockerValidator{
		CommandContext: func(_ context.Context, _ string, _ ...string) *exec.Cmd {
			return exec.Command("false")
		},
	}

	dockerErr := &validator.UserError{Key: "docker_permission_not_effective"}
	got := tryReexecInitWithDockerGroup(dockerErr, "LICENSE-ABCDEF0123456789", "PUB")
	if got == nil {
		t.Fatal("tryReexecInitWithDockerGroup() expected error")
	}
	if !errors.Is(got, dockerErr) {
		if validator.UserErrorKey(got) != "docker_permission_not_effective" {
			t.Fatalf("tryReexecInitWithDockerGroup() = %v", got)
		}
	}
}

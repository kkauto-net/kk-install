package validator

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/kkauto-net/kk-install/pkg/ui"
)

var errDockerGroupRunnerMissing = errors.New("neither sg nor newgrp is available")

func (v *DockerValidator) dockerGroupProbeCommand() string {
	if _, err := v.LookPath("sg"); err == nil {
		return `sg docker -c "docker info"`
	}
	if _, err := v.LookPath("newgrp"); err == nil {
		return `newgrp docker -c "docker info"`
	}
	return ""
}

// CanAccessDockerViaGroupSubcommand reports whether docker works via sg or newgrp.
func (v *DockerValidator) CanAccessDockerViaGroupSubcommand() bool {
	probe := v.dockerGroupProbeCommand()
	if probe == "" {
		return false
	}

	verifyCtx, verifyCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer verifyCancel()
	verifyCmd := v.CommandContext(verifyCtx, "sh", "-c", probe)
	return verifyCmd.Run() == nil
}

// HasDockerGroupRunner reports whether sg or newgrp is available.
func (v *DockerValidator) HasDockerGroupRunner() bool {
	if _, err := v.LookPath("sg"); err == nil {
		return true
	}
	if _, err := v.LookPath("newgrp"); err == nil {
		return true
	}
	return false
}

func dockerUsesSudo() bool {
	return os.Getenv("KK_DOCKER_SUDO") == "1"
}

func (v *DockerValidator) canAccessDockerWithSudo() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := v.CommandContext(ctx, "sudo", "docker", "info")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

func (v *DockerValidator) canAccessDockerWithSudoNoPassword() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := v.CommandContext(ctx, "sudo", "-n", "docker", "info")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

// TryActivateDockerSudoFallback enables KK_DOCKER_SUDO when group runners are unavailable.
func (v *DockerValidator) TryActivateDockerSudoFallback() bool {
	if v.HasDockerGroupRunner() {
		return false
	}
	if dockerUsesSudo() {
		return v.canAccessDockerWithSudo()
	}
	if !v.isUserInDockerGroup() && !v.isDockerDaemonRunningPrivileged() {
		return false
	}
	if v.canAccessDockerWithSudoNoPassword() {
		return v.activateDockerSudoEnv()
	}
	if !isInteractiveTTY() {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := ensureSudoAccess(v, ctx); err != nil {
		return false
	}
	if !v.canAccessDockerWithSudo() {
		return false
	}
	return v.activateDockerSudoEnv()
}

func (v *DockerValidator) activateDockerSudoEnv() bool {
	if err := os.Setenv("KK_DOCKER_SUDO", "1"); err != nil {
		return false
	}
	ui.ShowNote(ui.Msg("docker_sudo_fallback_note"))
	return true
}

func (v *DockerValidator) handlePermissionPending() error {
	if v.TryActivateDockerSudoFallback() {
		if err := v.CheckDockerDaemon(); err == nil {
			return nil
		}
	}
	return &UserError{Key: "docker_permission_not_effective"}
}

// DockerGroupReexecCommand returns a shell command to rerun kk init with docker group access.
func (v *DockerValidator) DockerGroupReexecCommand(initCommand string) string {
	if initCommand == "" {
		initCommand = "kk init"
	}
	if _, err := v.LookPath("sg"); err == nil {
		return fmt.Sprintf(`sg docker -c %q`, initCommand)
	}
	if _, err := v.LookPath("newgrp"); err == nil {
		return fmt.Sprintf(`newgrp docker -c %q`, initCommand)
	}
	return ui.Msg("docker_session_relogin_command")
}

// RunCommandWithDockerGroup runs a shell command with docker group privileges.
func (v *DockerValidator) RunCommandWithDockerGroup(command string, env []string) error {
	if _, err := v.LookPath("sg"); err == nil {
		cmd := exec.Command("sg", "docker", "-c", command)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = env
		return cmd.Run()
	}
	if _, err := v.LookPath("newgrp"); err == nil {
		cmd := exec.Command("newgrp", "docker", "-c", command)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = env
		return cmd.Run()
	}
	return errDockerGroupRunnerMissing
}

func (v *DockerValidator) isUserInDockerGroup() bool {
	user := strings.TrimSpace(os.Getenv("USER"))
	if user == "" {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := v.CommandContext(ctx, "getent", "group", "docker")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	parts := strings.SplitN(strings.TrimSpace(string(output)), ":", 4)
	if len(parts) < 4 {
		return false
	}

	for member := range strings.SplitSeq(parts[3], ",") {
		if strings.TrimSpace(member) == user {
			return true
		}
	}
	return false
}

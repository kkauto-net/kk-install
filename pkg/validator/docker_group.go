package validator

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
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
	return "newgrp docker"
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

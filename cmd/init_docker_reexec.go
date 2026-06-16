package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/kkauto-net/kk-install/pkg/ui"
	"github.com/kkauto-net/kk-install/pkg/validator"
)

const (
	dockerGroupReexecEnv       = "KK_DOCKER_GROUP_REEXEC"
	initValidatedLicenseEnv    = "KK_INIT_VALIDATED_LICENSE"
	initValidatedLicensePubEnv = "KK_INIT_LICENSE_PUBLIC_KEY"
)

var runSgDockerGroupWithEnv = func(command string, env []string) error {
	cmd := exec.Command("sg", "docker", "-c", command)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = env
	return cmd.Run()
}

func shouldAttemptDockerGroupReexec(err error) bool {
	if err == nil || os.Getenv(dockerGroupReexecEnv) == "1" {
		return false
	}

	key := validator.UserErrorKey(err)
	if key != "docker_permission_not_effective" && key != "docker_permission_denied" {
		return false
	}

	return DockerValidatorInstance.CanAccessDockerViaSG()
}

func buildInitReexecCommand() (string, error) {
	if len(os.Args) == 0 {
		return "", errors.New("missing process args")
	}

	quoted := make([]string, len(os.Args))
	for i, arg := range os.Args {
		quoted[i] = strconv.Quote(arg)
	}
	return strings.Join(quoted, " "), nil
}

func tryReexecInitWithDockerGroup(dockerErr error, licenseKey, licensePublicKey string) error {
	if !shouldAttemptDockerGroupReexec(dockerErr) {
		return dockerErr
	}

	command, err := buildInitReexecCommand()
	if err != nil {
		return dockerErr
	}

	ui.ShowInfo(ui.Msg("docker_group_reexec_info"))

	env := os.Environ()
	env = append(env, dockerGroupReexecEnv+"=1")
	if strings.TrimSpace(licenseKey) != "" {
		env = append(env, initValidatedLicenseEnv+"="+licenseKey)
	}
	if strings.TrimSpace(licensePublicKey) != "" {
		env = append(env, initValidatedLicensePubEnv+"="+licensePublicKey)
	}

	sgErr := runSgDockerGroupWithEnv(command, env)

	if sgErr == nil {
		os.Exit(0)
	}

	var exitErr *exec.ExitError
	if errors.As(sgErr, &exitErr) {
		os.Exit(exitErr.ExitCode())
	}

	return fmt.Errorf("%w: %v", dockerErr, sgErr)
}

func consumeReexecLicenseEnv() (licenseKey, publicKey string, ok bool) {
	if os.Getenv(dockerGroupReexecEnv) != "1" {
		return "", "", false
	}

	licenseKey = strings.TrimSpace(os.Getenv(initValidatedLicenseEnv))
	publicKey = strings.TrimSpace(os.Getenv(initValidatedLicensePubEnv))
	for _, key := range []string{initValidatedLicenseEnv, initValidatedLicensePubEnv} {
		if err := os.Unsetenv(key); err != nil {
			continue
		}
	}

	if licenseKey == "" {
		return "", "", false
	}
	return licenseKey, publicKey, true
}

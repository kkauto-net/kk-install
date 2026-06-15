package validator

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/kkauto-net/kk-install/pkg/ui"
)

// Define function types for mocking
type LookPathFunc func(file string) (string, error)
type CommandContextFunc func(ctx context.Context, name string, arg ...string) *exec.Cmd

// Validator struct holds the functions to be used, allowing them to be mocked
type DockerValidator struct {
	LookPath       LookPathFunc
	CommandContext CommandContextFunc
}

// EnsureDockerOptions controls Docker preflight and optional auto-remediation.
type EnsureDockerOptions struct {
	AutoFix        bool
	MaxRetries     int
	ConfirmInstall func() (bool, error)
	ConfirmStart   func() (bool, error)
	Install        func() error
	Start          func() error
}

// NewDockerValidator creates a new Validator with default (real) implementations
func NewDockerValidator() *DockerValidator {
	return &DockerValidator{
		LookPath:       exec.LookPath,
		CommandContext: exec.CommandContext,
	}
}

// CheckDockerInstalled verifies docker binary exists
func (v *DockerValidator) CheckDockerInstalled() error {
	_, err := v.LookPath("docker")
	if err != nil {
		return &UserError{Key: "docker_not_installed"}
	}
	return nil
}

// CheckDockerDaemon verifies docker daemon is running
func (v *DockerValidator) CheckDockerDaemon() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := v.CommandContext(ctx, "docker", "info")
	output, err := cmd.CombinedOutput()
	if err != nil {
		outputStr := strings.ToLower(string(output))
		// Check if it's a permission error (user not in docker group)
		if strings.Contains(outputStr, "permission denied") ||
			strings.Contains(outputStr, "got permission denied") ||
			strings.Contains(outputStr, "connect: permission denied") {
			return &UserError{Key: "docker_permission_denied"}
		}
		// Check if daemon is not running
		if strings.Contains(outputStr, "cannot connect") ||
			strings.Contains(outputStr, "is the docker daemon running") {
			return &UserError{Key: "docker_not_running"}
		}
		// Generic error
		return &UserError{Key: "docker_not_running"}
	}
	return nil
}

// CheckComposeVersion verifies Docker Compose v2.0+ is installed
func (v *DockerValidator) CheckComposeVersion() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try docker compose (v2) first
	cmd := v.CommandContext(ctx, "docker", "compose", "version", "--short")
	output, err := cmd.Output()
	if err != nil {
		// Fallback: try docker-compose (v1)
		cmd = v.CommandContext(ctx, "docker-compose", "version", "--short")
		output, err = cmd.Output()
		if err != nil {
			return &UserError{Key: "compose_not_found"}
		}
	}

	version := strings.TrimSpace(string(output))

	// Parse version (e.g., "v2.5.0" or "2.5.0")
	version = strings.TrimPrefix(version, "v")

	// Extract major version
	versionRegex := regexp.MustCompile(`^(\d+)\.(\d+)\.(\d+)`)
	matches := versionRegex.FindStringSubmatch(version)
	if len(matches) < 2 {
		// Cannot parse version, warn but don't block
		ui.ShowWarningf(ui.Msg("warn_compose_version_read"), version)
		return nil
	}

	major, err := strconv.Atoi(matches[1])
	if err != nil || major < 2 {
		return &UserError{
			Key:  "compose_version_old",
			Args: []any{version},
		}
	}

	return nil
}

// UserError represents user-friendly error
type UserError struct {
	Key        string
	Message    string
	Suggestion string
	Args       []any
}

func (e *UserError) Error() string {
	msg := e.Message
	if msg == "" && e.Key != "" {
		if len(e.Args) > 0 {
			msg = ui.MsgF(e.Key, e.Args...)
		} else {
			msg = ui.Msg(e.Key)
		}
	}
	if e.Suggestion != "" {
		return msg + " - " + e.Suggestion
	}
	suggestionKey := e.Key + "_suggestion"
	if e.Key != "" && ui.Msg(suggestionKey) != suggestionKey {
		return msg + " - " + ui.Msg(suggestionKey)
	}
	return msg
}

// UserErrorKey returns the UserError key when err is a UserError.
func UserErrorKey(err error) string {
	if err == nil {
		return ""
	}
	if userErr, ok := err.(*UserError); ok {
		return userErr.Key
	}
	return ""
}

// EnsureDockerReady validates Docker installation, daemon, and Compose.
// When AutoFix is enabled, missing Docker or a stopped daemon are remediated automatically.
func (v *DockerValidator) EnsureDockerReady(opts EnsureDockerOptions) error {
	maxRetries := opts.maxRetries()

	if err := v.CheckDockerInstalled(); err != nil {
		approved, approveErr := opts.approveInstall()
		if approveErr != nil {
			return approveErr
		}
		if !approved {
			return err
		}
		if installErr := v.installDockerWithRetry(maxRetries, opts.Install); installErr != nil {
			return installErr
		}
	}

	if err := v.ensureDaemonReady(opts, maxRetries); err != nil {
		return err
	}

	return v.CheckComposeVersion()
}

func (opts EnsureDockerOptions) maxRetries() int {
	if opts.MaxRetries <= 0 {
		return 1
	}
	return opts.MaxRetries
}

func (opts EnsureDockerOptions) approveInstall() (bool, error) {
	if opts.AutoFix {
		return true, nil
	}
	if opts.ConfirmInstall == nil {
		return false, nil
	}
	return opts.ConfirmInstall()
}

func (opts EnsureDockerOptions) approveStart() (bool, error) {
	if opts.AutoFix {
		return true, nil
	}
	if opts.ConfirmStart == nil {
		return false, nil
	}
	return opts.ConfirmStart()
}

func (v *DockerValidator) ensureDaemonReady(opts EnsureDockerOptions, maxRetries int) error {
	err := v.CheckDockerDaemon()
	if err == nil {
		return nil
	}

	key := UserErrorKey(err)
	if key == "docker_permission_denied" && (opts.AutoFix || opts.ConfirmStart != nil) {
		if fixErr := v.FixDockerPermissions(); fixErr != nil {
			ui.ShowWarningf(ui.Msg("warn_docker_permissions_fix_failed"), fixErr)
		} else if recheckErr := v.CheckDockerDaemon(); recheckErr == nil {
			return nil
		}
	}

	approved, approveErr := opts.approveStart()
	if approveErr != nil {
		return approveErr
	}
	if !approved {
		return err
	}

	if startErr := v.startDockerDaemonWithRetry(maxRetries, opts.Start); startErr != nil {
		return startErr
	}

	return v.waitForDockerDaemon(30 * time.Second)
}

func (v *DockerValidator) installDockerWithRetry(maxRetries int, install func() error) error {
	if install == nil {
		install = v.InstallDocker
	}
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		lastErr = install()
		if lastErr == nil {
			return nil
		}
	}
	return lastErr
}

func (v *DockerValidator) startDockerDaemonWithRetry(maxRetries int, start func() error) error {
	if start == nil {
		start = v.StartDockerDaemon
	}
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		lastErr = start()
		if lastErr == nil {
			return nil
		}
	}
	return lastErr
}

func (v *DockerValidator) waitForDockerDaemon(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	delay := 2 * time.Second

	for time.Now().Before(deadline) {
		if err := v.CheckDockerDaemon(); err == nil {
			return nil
		}
		time.Sleep(delay)
		if delay < 8*time.Second {
			delay += time.Second
		}
	}

	return &UserError{
		Key:        "docker_not_running",
		Message:    "Docker daemon chua san sang sau khi khoi dong",
		Suggestion: "Thu: sudo systemctl start docker && docker info",
	}
}

func (v *DockerValidator) checkInstallPrerequisites() error {
	if _, err := v.LookPath("curl"); err != nil {
		return &UserError{
			Key:        "docker_install_failed",
			Message:    "Thieu curl de cai Docker",
			Suggestion: "Cai curl roi chay lai hoac cai Docker thu cong: https://docs.docker.com/get-docker/",
		}
	}
	if _, err := v.LookPath("sudo"); err != nil {
		return &UserError{
			Key:        "docker_install_failed",
			Message:    "Thieu sudo de cai Docker",
			Suggestion: "Chay voi quyen root hoac cai Docker thu cong: https://docs.docker.com/get-docker/",
		}
	}
	return nil
}

func classifyDockerInstallFailure(output string) *UserError {
	lower := strings.ToLower(output)
	switch {
	case strings.Contains(lower, "sudo:") && strings.Contains(lower, "password"):
		return &UserError{
			Key:        "docker_install_failed",
			Message:    "Can quyen sudo de cai Docker",
			Suggestion: "Chay lai voi user co sudo hoac cai thu cong: https://docs.docker.com/get-docker/",
		}
	case strings.Contains(lower, "could not resolve host"), strings.Contains(lower, "failed to connect"), strings.Contains(lower, "timed out"):
		return &UserError{
			Key:        "docker_install_failed",
			Message:    "Khong the tai script cai Docker",
			Suggestion: "Kiem tra mang hoac cai thu cong: https://docs.docker.com/get-docker/",
		}
	case strings.Contains(lower, "could not get lock"), strings.Contains(lower, "dpkg lock"), strings.Contains(lower, "another process"):
		return &UserError{
			Key:        "docker_install_failed",
			Message:    "Package manager dang bi khoa",
			Suggestion: "Doi package manager xong roi chay lai hoac cai thu cong: https://docs.docker.com/get-docker/",
		}
	default:
		return &UserError{
			Key:        "docker_install_failed",
			Message:    "Docker installation failed",
			Suggestion: "Try manual install: https://docs.docker.com/get-docker/",
		}
	}
}

// InstallDocker attempts to install Docker using the official convenience script.
func (v *DockerValidator) InstallDocker() error {
	if err := v.checkInstallPrerequisites(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Print newline before sudo prompt for better formatting
	fmt.Println()

	cmd := v.CommandContext(ctx, "sh", "-c", "curl -fsSL https://get.docker.com | sudo sh")
	var stderr bytes.Buffer
	cmd.Stdout = nil
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		output := stderr.String()
		if output == "" {
			output = err.Error()
		}
		return classifyDockerInstallFailure(output)
	}

	if err := v.FixDockerPermissions(); err != nil {
		ui.ShowWarningf(ui.Msg("warn_docker_group_add_failed"), err)
	}

	return nil
}

// FixDockerPermissions adds the current user to the docker group and verifies access.
func (v *DockerValidator) FixDockerPermissions() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	userCmd := v.CommandContext(ctx, "sh", "-c", "sudo usermod -aG docker $USER")
	if err := userCmd.Run(); err != nil {
		return err
	}

	verifyCtx, verifyCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer verifyCancel()
	verifyCmd := v.CommandContext(verifyCtx, "sh", "-c", "sg docker -c \"docker info\"")
	if err := verifyCmd.Run(); err != nil {
		return &UserError{
			Key:        "docker_permission_denied",
			Message:    "Da them user vao docker group nhung quyen chua co hieu luc",
			Suggestion: "Chay: newgrp docker hoac dang nhap lai, roi thu lai",
		}
	}

	return nil
}

// StartDockerDaemon attempts to start the Docker daemon
func (v *DockerValidator) StartDockerDaemon() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Print newline before sudo prompt for better formatting
	fmt.Println()

	// Try systemctl first (most common on Linux)
	cmd := v.CommandContext(ctx, "sudo", "systemctl", "start", "docker")
	if err := cmd.Run(); err != nil {
		// Fallback: try service command
		cmd = v.CommandContext(ctx, "sudo", "service", "docker", "start")
		if err := cmd.Run(); err != nil {
			return &UserError{
				Key:        "docker_start_failed",
				Message:    "Failed to start Docker daemon",
				Suggestion: "Try: sudo systemctl start docker",
			}
		}
	}

	return nil
}

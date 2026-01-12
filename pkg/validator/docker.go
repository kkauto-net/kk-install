package validator

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Define function types for mocking
type LookPathFunc func(file string) (string, error)
type CommandContextFunc func(ctx context.Context, name string, arg ...string) *exec.Cmd

// Validator struct holds the functions to be used, allowing them to be mocked
type DockerValidator struct {
	LookPath       LookPathFunc
	CommandContext CommandContextFunc
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
		return &UserError{
			Key:        "docker_not_installed",
			Message:    "Docker chua cai dat",
			Suggestion: "Cai tai: https://docs.docker.com/get-docker/",
		}
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
			return &UserError{
				Key:        "docker_permission_denied",
				Message:    "Khong co quyen truy cap Docker",
				Suggestion: "Them user vao docker group: sudo usermod -aG docker $USER && newgrp docker",
			}
		}
		// Check if daemon is not running
		if strings.Contains(outputStr, "cannot connect") ||
			strings.Contains(outputStr, "is the docker daemon running") {
			return &UserError{
				Key:        "docker_not_running",
				Message:    "Docker daemon khong chay",
				Suggestion: "Chay: sudo systemctl start docker",
			}
		}
		// Generic error
		return &UserError{
			Key:        "docker_not_running",
			Message:    "Docker daemon khong chay hoac khong co quyen",
			Suggestion: "Thu: sudo systemctl start docker HOAC sudo usermod -aG docker $USER && newgrp docker",
		}
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
			return &UserError{
				Key:        "compose_not_found",
				Message:    "Docker Compose khong tim thay",
				Suggestion: "Cai dat Docker Compose: https://docs.docker.com/compose/install/",
			}
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
		fmt.Printf("  [!] Canh bao: Khong doc duoc phien ban Docker Compose (%s)\n", version)
		return nil
	}

	major, err := strconv.Atoi(matches[1])
	if err != nil || major < 2 {
		return &UserError{
			Key:        "compose_version_old",
			Message:    fmt.Sprintf("Docker Compose phien ban cu (%s), can >= v2.0", version),
			Suggestion: "Cap nhat Docker Compose: https://docs.docker.com/compose/install/",
		}
	}

	return nil
}

// UserError represents user-friendly error
type UserError struct {
	Key        string
	Message    string
	Suggestion string
}

func (e *UserError) Error() string {
	if e.Suggestion != "" {
		return e.Message + " - " + e.Suggestion
	}
	return e.Message
}

// InstallDocker attempts to install Docker using the official convenience script
// Returns nil on success, error on failure
func (v *DockerValidator) InstallDocker() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Print newline before sudo prompt for better formatting
	fmt.Println()

	// Use official Docker install script for Linux
	// curl -fsSL https://get.docker.com | sh
	cmd := v.CommandContext(ctx, "sh", "-c", "curl -fsSL https://get.docker.com | sudo sh")
	cmd.Stdout = nil // Will be captured
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return &UserError{
			Key:        "docker_install_failed",
			Message:    "Docker installation failed",
			Suggestion: "Try manual install: https://docs.docker.com/get-docker/",
		}
	}

	// Add current user to docker group
	userCmd := v.CommandContext(ctx, "sh", "-c", "sudo usermod -aG docker $USER")
	_ = userCmd.Run() // Best effort, don't fail if this fails

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

	// Wait a bit for daemon to be ready
	time.Sleep(2 * time.Second)

	return nil
}

package validator

import (
	"context"
	"os/exec"
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
	if err := cmd.Run(); err != nil {
		return &UserError{
			Key:        "docker_not_running",
			Message:    "Docker daemon khong chay",
			Suggestion: "Chay: sudo systemctl start docker",
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

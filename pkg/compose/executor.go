package compose

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Variables for dependency injection in tests
var (
	execCommand  = exec.CommandContext
	execLookPath = exec.LookPath
)

// Executor wraps docker-compose commands
type Executor struct {
	WorkDir     string
	ComposeFile string
}

func NewExecutor(workDir string) *Executor {
	return &Executor{
		WorkDir:     workDir,
		ComposeFile: filepath.Join(workDir, "docker-compose.yml"),
	}
}

// Up runs docker-compose up -d
func (e *Executor) Up(ctx context.Context) error {
	return e.runWithStderrCapture(ctx, "up", "-d")
}

// Down runs docker-compose down
func (e *Executor) Down(ctx context.Context) error {
	return e.run(ctx, "down")
}

// DownWithVolumes runs docker-compose down -v (removes volumes too)
func (e *Executor) DownWithVolumes(ctx context.Context) error {
	return e.run(ctx, "down", "-v")
}

// Restart runs docker-compose restart
func (e *Executor) Restart(ctx context.Context) error {
	return e.run(ctx, "restart")
}

// Pull runs docker-compose pull
func (e *Executor) Pull(ctx context.Context) (string, error) {
	return e.runWithOutput(ctx, "pull")
}

// Ps runs docker-compose ps
func (e *Executor) Ps(ctx context.Context) (string, error) {
	return e.runWithOutput(ctx, "ps", "--format", "json")
}

// ForceRecreate runs docker-compose up -d --force-recreate
func (e *Executor) ForceRecreate(ctx context.Context) error {
	return e.run(ctx, "up", "-d", "--force-recreate")
}

func (e *Executor) run(ctx context.Context, args ...string) error {
	cmd := e.buildCmd(ctx, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// runWithStderrCapture runs command with stdout to console but captures stderr for error details
func (e *Executor) runWithStderrCapture(ctx context.Context, args ...string) error {
	cmd := e.buildCmd(ctx, args...)
	var stderr bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = io.MultiWriter(os.Stderr, &stderr)

	err := cmd.Run()
	if err != nil {
		// Include stderr in error message for better error detection
		stderrStr := stderr.String()
		if stderrStr != "" {
			return fmt.Errorf("%s", stderrStr)
		}
		return err
	}
	return nil
}

func (e *Executor) runWithOutput(ctx context.Context, args ...string) (string, error) {
	cmd := e.buildCmd(ctx, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, stderr.String())
	}
	return stdout.String(), nil
}

func (e *Executor) buildCmd(ctx context.Context, args ...string) *exec.Cmd {
	// Try docker compose (v2) first, fallback to docker-compose (v1)
	cmdName := "docker"
	cmdArgs := append([]string{"compose", "-f", e.ComposeFile}, args...)

	// Check if docker compose v2 is available
	if _, err := execLookPath("docker"); err == nil {
		testCmd := exec.Command("docker", "compose", "version")
		if testCmd.Run() != nil {
			// Fallback to docker-compose v1
			cmdName = "docker-compose"
			cmdArgs = append([]string{"-f", e.ComposeFile}, args...)
		}
	}

	cmd := execCommand(ctx, cmdName, cmdArgs...)
	cmd.Dir = e.WorkDir
	return cmd
}

// DefaultTimeout for compose operations
const DefaultTimeout = 5 * time.Minute

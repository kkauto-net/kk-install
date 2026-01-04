package compose

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Command is an interface that matches the required methods of exec.Cmd
type Command interface {
	Run() error
	Start() error
	Wait() error
	CombinedOutput() ([]byte, error)
	StderrPipe() (io.ReadCloser, error)
	StdoutPipe() (io.ReadCloser, error)
}

// MockCmd is a mock implementation of Command for testing purposes
type MockCmd struct {
	runFunc         func() error
	startFunc       func() error
	waitFunc        func() error
	combinedOutFunc func() ([]byte, error)
	stderrPipeFunc  func() (io.ReadCloser, error)
	stdoutPipeFunc  func() (io.ReadCloser, error)

	Output []byte
	Error  error
	CmdArgs []string // To capture arguments passed to the mock
}

func (m *MockCmd) Run() error {
	if m.runFunc != nil {
		return m.runFunc()
	}
	return m.Error
}

func (m *MockCmd) Start() error {
	if m.startFunc != nil {
		return m.startFunc()
	}
	return m.Error
}

func (m *MockCmd) Wait() error {
	if m.waitFunc != nil {
		return m.waitFunc()
	}
	return m.Error
}

func (m *MockCmd) CombinedOutput() ([]byte, error) {
	if m.combinedOutFunc != nil {
		return m.combinedOutFunc()
	}
	return m.Output, m.Error
}

func (m *MockCmd) StderrPipe() (io.ReadCloser, error) {
	if m.stderrPipeFunc != nil {
		return m.stderrPipeFunc()
	}
	return io.NopCloser(bytes.NewReader([]byte(""))), nil
}

func (m *MockCmd) StdoutPipe() (io.ReadCloser, error) {
	if m.stdoutPipeFunc != nil {
		return m.stdoutPipeFunc()
	}
	return io.NopCloser(bytes.NewReader([]byte(""))), nil
}

// Global variable to replace exec.Command for testing
var osExecCommand = func(name string, arg ...string) Command {
	return &MockCmd{}
}

// Global variable to replace exec.LookPath for testing
var osExecLookPath = func(file string) (string, error) {
	return exec.LookPath(file)
}

func TestExecutor_Up(t *testing.T) {
	origOsExecCommand := osExecCommand
	origOsExecLookPath := osExecLookPath
	defer func() {
		osExecCommand = origOsExecCommand
		osExecLookPath = origOsExecLookPath
	}()

	t.Run("Up with docker compose v2", func(t *testing.T) {
		mockComposeVersionOutput := []byte("Docker Compose version v2.1.1\n")
		var capturedCmdArgs []string

		// Mock exec.LookPath to find "docker"
		osExecLookPath = func(file string) (string, error) {
			if file == "docker" {
				return "/usr/local/bin/docker", nil
			}
			return "", exec.ErrNotFound
		}

		// Mock exec.Command to simulate "docker compose version" and actual "up" command
		osExecCommand = func(name string, arg ...string) Command {
			if name == "docker" && len(arg) > 0 && arg[0] == "compose" && arg[1] == "version" {
				return &MockCmd{Output: mockComposeVersionOutput, Error: nil}
			}
			capturedCmdArgs = append([]string{name}, arg...)
			return &MockCmd{Error: nil, CmdArgs: arg} // Simulate successful run
		}

		executor := NewExecutor("/tmp/test-compose")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := executor.Up(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "docker", capturedCmdArgs[0])
		assert.NotEmpty(t, capturedCmdArgs, "capturedCmdArgs should not be empty")
		if len(capturedCmdArgs) > 0 {
		assert.Contains(t, strings.Join(capturedCmdArgs, " "), "-f /tmp/test-compose/docker-compose.yml up -d")
		}
	})

	t.Run("Up with docker-compose v1 fallback", func(t *testing.T) {
		var capturedCmdArgs []string

		// Mock exec.LookPath to not find "docker" but find "docker-compose"
		osExecLookPath = func(file string) (string, error) {
			if file == "docker" {
				return "", exec.ErrNotFound
			}
			if file == "docker-compose" {
				return "/usr/local/bin/docker-compose", nil
			}
			return "", exec.ErrNotFound
		}

		// Mock exec.Command to simulate actual "up" command
		osExecCommand = func(name string, arg ...string) Command {
			capturedCmdArgs = append([]string{name}, arg...)
			return &MockCmd{Error: nil, CmdArgs: arg} // Simulate successful run
		}

		executor := NewExecutor("/tmp/test-compose")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		err := executor.Up(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "docker-compose", capturedCmdArgs[0])
		assert.Contains(t, strings.Join(capturedCmdArgs, " "), "-f /tmp/test-compose/docker-compose.yml up -d")
	})

	t.Run("Up command fails", func(t *testing.T) {
		// Mock exec.LookPath to find "docker"
		osExecLookPath = func(file string) (string, error) {
			if file == "docker" {
				return "/usr/local/bin/docker", nil
			}
			return "", exec.ErrNotFound
		}

		// Mock exec.Command to simulate "docker compose version" success, but actual "up" command fails
		osExecCommand = func(name string, arg ...string) Command {
			if name == "docker" && len(arg) > 0 && arg[0] == "compose" && arg[1] == "version" {
				return &MockCmd{Output: []byte("Docker Compose version v2.1.1\n"), Error: nil}
			}
			return &MockCmd{Error: fmt.Errorf("process error"), CmdArgs: arg} // Simulate command failure
		}

		executor := NewExecutor("/tmp/test-compose")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		err := executor.Up(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "process error")
	})

	t.Run("Up command times out", func(t *testing.T) {
		// Note: DefaultTimeout is a const and cannot be modified
		// We simulate timeout by using a very short context timeout

		// Mock exec.LookPath to find "docker"
		osExecLookPath = func(file string) (string, error) {
			if file == "docker" {
				return "/usr/local/bin/docker", nil
			}
			return "", exec.ErrNotFound
		}

		// Mock exec.Command to simulate a command that takes longer than the timeout
		osExecCommand = func(name string, arg ...string) Command {
			if name == "docker" && len(arg) > 0 && arg[0] == "compose" && arg[1] == "version" {
				return &MockCmd{Output: []byte("Docker Compose version v2.1.1\n"), Error: nil}
			}
			return &MockCmd{
				runFunc: func() error {
					time.Sleep(20 * time.Millisecond) // Simulate long running command
					return nil
				},
				CmdArgs: arg,
			}
		}

		executor := NewExecutor("/tmp/test-compose")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		err := executor.Up(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context deadline exceeded")
	})
}

func TestExecutor_Down(t *testing.T) {
	origOsExecCommand := osExecCommand
	origOsExecLookPath := osExecLookPath
	defer func() {
		osExecCommand = origOsExecCommand
		osExecLookPath = origOsExecLookPath
	}()

	t.Run("Down successful", func(t *testing.T) {
		var capturedCmdArgs []string
		osExecLookPath = func(file string) (string, error) { // Assume v2 exists for simplicity
			if file == "docker" { return "", nil }
			return "", exec.ErrNotFound
		}
		osExecCommand = func(name string, arg ...string) Command {
			if name == "docker" && len(arg) > 0 && arg[0] == "compose" && arg[1] == "version" {
				return &MockCmd{Output: []byte("Docker Compose version v2.1.1\n"), Error: nil}
			}
			capturedCmdArgs = append([]string{name}, arg...)
			return &MockCmd{Error: nil}
		}

		executor := NewExecutor("/tmp/test-compose")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		err := executor.Down(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "docker", capturedCmdArgs[0])
		assert.Contains(t, strings.Join(capturedCmdArgs, " "), "-f /tmp/test-compose/docker-compose.yml down")
	})
}

func TestExecutor_Restart(t *testing.T) {
	origOsExecCommand := osExecCommand
	origOsExecLookPath := osExecLookPath
	defer func() {
		osExecCommand = origOsExecCommand
		osExecLookPath = origOsExecLookPath
	}()

	t.Run("Restart successful", func(t *testing.T) {
		var capturedCmdArgs []string
		osExecLookPath = func(file string) (string, error) { // Assume v2 exists for simplicity
			if file == "docker" { return "", nil }
			return "", exec.ErrNotFound
		}
		osExecCommand = func(name string, arg ...string) Command {
			if name == "docker" && len(arg) > 0 && arg[0] == "compose" && arg[1] == "version" {
				return &MockCmd{Output: []byte("Docker Compose version v2.1.1\n"), Error: nil}
			}
			capturedCmdArgs = append([]string{name}, arg...)
			return &MockCmd{Error: nil}
		}

		executor := NewExecutor("/tmp/test-compose")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		err := executor.Restart(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "docker", capturedCmdArgs[0])
		assert.Contains(t, strings.Join(capturedCmdArgs, " "), "-f /tmp/test-compose/docker-compose.yml restart")
	})
}

func TestExecutor_Pull(t *testing.T) {
	origOsExecCommand := osExecCommand
	origOsExecLookPath := osExecLookPath
	defer func() {
		osExecCommand = origOsExecCommand
		osExecLookPath = origOsExecLookPath
	}()

	t.Run("Pull successful", func(t *testing.T) {
		mockOutput := "Image pulled successfully"
		var capturedCmdArgs []string
		osExecLookPath = func(file string) (string, error) {
			if file == "docker" { return "", nil }
			return "", exec.ErrNotFound
		}
		osExecCommand = func(name string, arg ...string) Command {
			if name == "docker" && len(arg) > 0 && arg[0] == "compose" && arg[1] == "version" {
				return &MockCmd{Output: []byte("Docker Compose version v2.1.1\n"), Error: nil}
			}
			capturedCmdArgs = append([]string{name}, arg...)
			return &MockCmd{Output: []byte(mockOutput), Error: nil}
		}

		executor := NewExecutor("/tmp/test-compose")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		output, err := executor.Pull(ctx)
		assert.NoError(t, err)
		assert.Equal(t, mockOutput, output)
		assert.Equal(t, "docker", capturedCmdArgs[0])
		assert.Contains(t, strings.Join(capturedCmdArgs, " "), "-f /tmp/test-compose/docker-compose.yml pull")
	})

	t.Run("Pull fails with error output", func(t *testing.T) {
		mockStderr := "pull failed: no such image"
		var capturedCmdArgs []string
		osExecLookPath = func(file string) (string, error) {
			if file == "docker" { return "", nil }
			return "", exec.ErrNotFound
		}

		mockCmd := &MockCmd{
			Error: exec.Command("false").Err, // Simulate ExitError
			combinedOutFunc: func() ([]byte, error) {
				return nil, fmt.Errorf("exit status 1: %s", mockStderr) // CombinedOutput returns stderr in error
			},
			stderrPipeFunc: func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader([]byte(mockStderr))), nil
			},
		}

		osExecCommand = func(name string, arg ...string) Command {
			if name == "docker" && len(arg) > 0 && arg[0] == "compose" && arg[1] == "version" {
				return &MockCmd{Output: []byte("Docker Compose version v2.1.1\n"), Error: nil}
			}
			capturedCmdArgs = append([]string{name}, arg...)
			return mockCmd
		}

		executor := NewExecutor("/tmp/test-compose")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		output, err := executor.Pull(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), mockStderr)
		assert.Empty(t, output)
		assert.Equal(t, "docker", capturedCmdArgs[0])
		assert.Contains(t, strings.Join(capturedCmdArgs, " "), "-f /tmp/test-compose/docker-compose.yml pull")
	})
}

func TestExecutor_Ps(t *testing.T) {
	origOsExecCommand := osExecCommand
	origOsExecLookPath := osExecLookPath
	defer func() {
		osExecCommand = origOsExecCommand
		osExecLookPath = origOsExecLookPath
	}()

	t.Run("Ps successful", func(t *testing.T) {
		mockOutput := `[{"Name":"service1","State":"running"}]`
		var capturedCmdArgs []string
		osExecLookPath = func(file string) (string, error) {
			if file == "docker" { return "", nil }
			return "", exec.ErrNotFound
		}
		osExecCommand = func(name string, arg ...string) Command {
			if name == "docker" && len(arg) > 0 && arg[0] == "compose" && arg[1] == "version" {
				return &MockCmd{Output: []byte("Docker Compose version v2.1.1\n"), Error: nil}
			}
			capturedCmdArgs = append([]string{name}, arg...)
			return &MockCmd{Output: []byte(mockOutput), Error: nil}
		}

		executor := NewExecutor("/tmp/test-compose")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		output, err := executor.Ps(ctx)
		assert.NoError(t, err)
		assert.Equal(t, mockOutput, output)
		assert.Equal(t, "docker", capturedCmdArgs[0])
		assert.Contains(t, strings.Join(capturedCmdArgs, " "), "-f /tmp/test-compose/docker-compose.yml ps --format json")
	})
}

func TestExecutor_ForceRecreate(t *testing.T) {
	origOsExecCommand := osExecCommand
	origOsExecLookPath := osExecLookPath
	defer func() {
		osExecCommand = origOsExecCommand
		osExecLookPath = origOsExecLookPath
	}()

	t.Run("ForceRecreate successful", func(t *testing.T) {
		var capturedCmdArgs []string
		osExecLookPath = func(file string) (string, error) { // Assume v2 exists for simplicity
			if file == "docker" { return "", nil }
			return "", exec.ErrNotFound
		}
		osExecCommand = func(name string, arg ...string) Command {
			if name == "docker" && len(arg) > 0 && arg[0] == "compose" && arg[1] == "version" {
				return &MockCmd{Output: []byte("Docker Compose version v2.1.1\n"), Error: nil}
			}
			capturedCmdArgs = append([]string{name}, arg...)
			return &MockCmd{Error: nil}
		}

		executor := NewExecutor("/tmp/test-compose")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		err := executor.ForceRecreate(ctx)
		assert.NoError(t, err)
		assert.Equal(t, "docker", capturedCmdArgs[0])
		assert.Contains(t, strings.Join(capturedCmdArgs, " "), "-f /tmp/test-compose/docker-compose.yml up -d --force-recreate")
	})
}

// Intercept exec.Command and exec.LookPath calls in the executor package

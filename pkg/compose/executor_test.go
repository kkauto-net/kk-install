package compose

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// All tests in this file require Docker to be running
// Skip in CI environment where Docker may not be available

func TestExecutor_Up(t *testing.T) {
	t.Skip("Skipping Docker integration test - requires Docker daemon")
}

func TestExecutor_Down(t *testing.T) {
	t.Skip("Skipping Docker integration test - requires Docker daemon")
}

func TestExecutor_Restart(t *testing.T) {
	t.Skip("Skipping Docker integration test - requires Docker daemon")
}

func TestExecutor_Pull(t *testing.T) {
	t.Skip("Skipping Docker integration test - requires Docker daemon")
}

func TestExecutor_Ps(t *testing.T) {
	t.Skip("Skipping Docker integration test - requires Docker daemon")
}

func TestExecutor_ForceRecreate(t *testing.T) {
	t.Skip("Skipping Docker integration test - requires Docker daemon")
}

func TestExecutorCommandConstruction(t *testing.T) {
	tests := []struct {
		name     string
		psOutput string
		run      func(context.Context, *Executor) error
		want     []string
	}{
		{name: "up", run: func(ctx context.Context, e *Executor) error { return e.Up(ctx) }, want: []string{"docker compose -f COMPOSE up -d"}},
		{name: "down", run: func(ctx context.Context, e *Executor) error { return e.Down(ctx) }, want: []string{"docker compose -f COMPOSE down"}},
		{name: "down with volumes", run: func(ctx context.Context, e *Executor) error { return e.DownWithVolumes(ctx) }, want: []string{"docker compose -f COMPOSE down -v"}},
		{name: "restart when running", psOutput: "abc123\n", run: func(ctx context.Context, e *Executor) error { return e.Restart(ctx) }, want: []string{"docker compose -f COMPOSE ps -q", "docker compose -f COMPOSE restart"}},
		{name: "restart when stopped", psOutput: "", run: func(ctx context.Context, e *Executor) error { return e.Restart(ctx) }, want: []string{"docker compose -f COMPOSE ps -q", "docker compose -f COMPOSE up -d"}},
		{name: "force recreate", run: func(ctx context.Context, e *Executor) error { return e.ForceRecreate(ctx) }, want: []string{"docker compose -f COMPOSE up -d --force-recreate"}},
		{name: "pull", run: func(ctx context.Context, e *Executor) error { _, err := e.Pull(ctx); return err }, want: []string{"docker compose -f COMPOSE pull"}},
		{name: "ps", run: func(ctx context.Context, e *Executor) error { _, err := e.Ps(ctx); return err }, want: []string{"docker compose -f COMPOSE ps --format json"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := withFakeComposeCommands(t, false, 0, tt.psOutput, "ok\n")
			dir := t.TempDir()
			executor := NewExecutor(dir)

			if err := tt.run(context.Background(), executor); err != nil {
				t.Fatalf("executor command error = %v", err)
			}

			got := normalizeComposeCalls(*calls, executor.ComposeFile)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("commands = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestExecutorFallbackToDockerComposeV1(t *testing.T) {
	calls := withFakeComposeCommands(t, true, 0, "", "ok\n")
	dir := t.TempDir()
	executor := NewExecutor(dir)

	if err := executor.Down(context.Background()); err != nil {
		t.Fatalf("Down() error = %v", err)
	}

	got := normalizeComposeCalls(*calls, executor.ComposeFile)
	want := []string{"docker-compose -f COMPOSE down"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("commands = %#v, want %#v", got, want)
	}
}

func TestExecutorUsesSudoWhenEnvSet(t *testing.T) {
	t.Setenv("KK_DOCKER_SUDO", "1")
	calls := withFakeComposeCommands(t, false, 0, "", "ok\n")
	dir := t.TempDir()
	executor := NewExecutor(dir)

	if err := executor.Down(context.Background()); err != nil {
		t.Fatalf("Down() error = %v", err)
	}

	got := normalizeComposeCalls(*calls, executor.ComposeFile)
	want := []string{"sudo docker compose -f COMPOSE down"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("commands = %#v, want %#v", got, want)
	}
}

func TestExecutorPropagatesCommandErrors(t *testing.T) {
	withFakeComposeCommands(t, false, 7, "", "compose failed")
	executor := NewExecutor(t.TempDir())

	err := executor.Up(context.Background())
	if err == nil {
		t.Fatal("Up() expected error")
	}
	if !strings.Contains(err.Error(), "compose failed") {
		t.Fatalf("Up() error = %q, want stderr", err.Error())
	}
}

func withFakeComposeCommands(t *testing.T, dockerV2Unavailable bool, exitCode int, psOutput string, defaultOutput string) *[]string {
	t.Helper()
	oldExecCommand := execCommand
	oldExecLookPath := execLookPath
	calls := []string{}
	execLookPath = func(file string) (string, error) {
		if file == "docker" {
			return "/usr/bin/docker", nil
		}
		return "", errors.New("not found")
	}
	execCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		if name == "docker" && reflect.DeepEqual(args, []string{"compose", "version"}) {
			if dockerV2Unavailable {
				return fakeExecCommand(t, 1, "")
			}
			return fakeExecCommand(t, 0, "Docker Compose version v2.0.0")
		}
		calls = append(calls, strings.Join(append([]string{name}, args...), " "))
		output := defaultOutput
		if len(args) >= 3 && args[0] == "compose" && args[len(args)-2] == "ps" && args[len(args)-1] == "-q" {
			output = psOutput
		}
		return fakeExecCommand(t, exitCode, output)
	}
	t.Cleanup(func() {
		execCommand = oldExecCommand
		execLookPath = oldExecLookPath
	})
	return &calls
}

func fakeExecCommand(t *testing.T, exitCode int, output string) *exec.Cmd {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=TestFakeExecutorProcess", "--")
	cmd.Env = append(os.Environ(),
		"KK_FAKE_EXEC=1",
		"KK_FAKE_EXEC_EXIT_CODE="+strconv.Itoa(exitCode),
		"KK_FAKE_EXEC_OUTPUT="+output,
	)
	return cmd
}

func TestFakeExecutorProcess(t *testing.T) {
	if os.Getenv("KK_FAKE_EXEC") != "1" {
		return
	}
	output := os.Getenv("KK_FAKE_EXEC_OUTPUT")
	if _, err := os.Stdout.WriteString(output); err != nil {
		os.Exit(1)
	}
	if _, err := os.Stderr.WriteString(output); err != nil {
		os.Exit(1)
	}
	if os.Getenv("KK_FAKE_EXEC_EXIT_CODE") != "0" {
		os.Exit(1)
	}
	os.Exit(0)
}

func normalizeComposeCalls(calls []string, composeFile string) []string {
	normalized := make([]string, len(calls))
	for i, call := range calls {
		normalized[i] = strings.ReplaceAll(call, filepath.ToSlash(composeFile), "COMPOSE")
	}
	return normalized
}

package main

import (
	"bytes"
	"context" // Add context for mockDockerValidator
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kkauto-net/kk-install/cmd" // Import the cmd package to access DockerValidatorInstance
	"github.com/kkauto-net/kk-install/pkg/validator"
)

// ensureKkBinary builds the 'kk' binary if it doesn't exist or is outdated.
func ensureKkBinary(t *testing.T) string {
	kkPath := filepath.Join(os.TempDir(), "kk_test_binary") // Build to a temp location

	// Check if the binary already exists and is executable
	info, err := os.Stat(kkPath)
	if err == nil && !info.IsDir() && info.Mode()&0111 != 0 {
		t.Logf("Using existing kk binary at %s", kkPath)
		return kkPath
	}

	cmdExec := exec.Command("go", "build", "-o", kkPath, ".")
	cmdExec.Dir = "/home/kkdev/kkcli" // Project root
	var stderr bytes.Buffer
	cmdExec.Stderr = &stderr
	if err := cmdExec.Run(); err != nil {
		t.Fatalf("Failed to build kk binary: %v\n%s", err, stderr.String())
	}
	t.Logf("Successfully built kk binary at %s", kkPath)
	return kkPath
}

// mockDockerValidator allows simulating Docker status for integration tests
func mockDockerValidator(installed bool, daemonRunning bool) *validator.DockerValidator {
	return &validator.DockerValidator{
		LookPath: func(file string) (string, error) {
			if installed {
				return "/usr/bin/docker", nil
			}
			return "", os.ErrNotExist
		},
		CommandContext: func(ctx context.Context, name string, arg ...string) *exec.Cmd {
			if daemonRunning {
				return exec.CommandContext(ctx, "true") // Simulate success
			}
			return exec.CommandContext(ctx, "false") // Simulate failure
		},
	}
}


func TestKkVersion(t *testing.T) {
	kkPath := ensureKkBinary(t)

	// Reset to default validator after test
	originalValidator := cmd.DockerValidatorInstance
	defer func() { cmd.DockerValidatorInstance = originalValidator }()
	cmd.DockerValidatorInstance = mockDockerValidator(true, true) // Ensure Docker is seen as working

	cmdExec := exec.Command(kkPath, "--version")
	output, err := cmdExec.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to run 'kk --version': %v\nOutput: %s", err, output)
	}

	expectedVersion := "kk version 0.1.0" // Based on cmd/root.go
	if !strings.Contains(string(output), expectedVersion) {
		t.Errorf("Version output mismatch. Got:\n%s\nWant to contain: %q", output, expectedVersion)
	}
}

func TestKkInit_HappyPath(t *testing.T) {
	kkPath := ensureKkBinary(t)
	tempDir := t.TempDir() // Create a temporary directory for this test

	// Reset to default validator after test
	originalValidator := cmd.DockerValidatorInstance
	defer func() { cmd.DockerValidatorInstance = originalValidator }()
	cmd.DockerValidatorInstance = mockDockerValidator(true, true) // Ensure Docker is seen as working

	// Simulate user input for huh forms:
	// 1. No overwrite (since file doesn't exist)
	// 2. No SeaweedFS (n)
	// 3. No Caddy (n)
	input := strings.NewReader("n\nn\n")

	cmdExec := exec.Command(kkPath, "init")
	cmdExec.Dir = tempDir
	cmdExec.Stdin = input
	output, err := cmdExec.CombinedOutput()
	if err != nil {
		t.Fatalf("kk init failed: %v\nOutput:\n%s", err, output)
	}

	// Verify expected files are created
	expectedFiles := []string{"docker-compose.yml", ".env", "kkphp.conf"}
	for _, file := range expectedFiles {
		path := filepath.Join(tempDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected file %s not created", file)
		}
	}

	// Verify Caddyfile and kkfiler.toml are NOT created
	unexpectedFiles := []string{"Caddyfile", "kkfiler.toml"}
	for _, file := range unexpectedFiles {
		path := filepath.Join(tempDir, file)
		if _, err := os.Stat(path); err == nil {
			t.Errorf("Unexpected file %s created", file)
		}
	}

	// Verify .env content (passwords and domain=localhost)
	envContent, err := os.ReadFile(filepath.Join(tempDir, ".env"))
	if err != nil {
		t.Fatalf("Failed to read .env file: %v", err)
	}
	if !strings.Contains(string(envContent), "DOMAIN=localhost") ||
		!strings.Contains(string(envContent), "DB_PASSWORD=") ||
		!strings.Contains(string(envContent), "REDIS_PASSWORD=") {
		t.Errorf(".env content mismatch. Got:\n%s", string(envContent))
	}
	info, err := os.Stat(filepath.Join(tempDir, ".env"))
	if err != nil {
		t.Fatalf("Failed to stat .env file: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf(".env permissions mismatch. Got: %v, Want: 0600", info.Mode().Perm())
	}

	// Verify output messages
	if !strings.Contains(string(output), "Khoi tao hoan tat!") {
		t.Errorf("Expected 'Khoi tao hoan tat!' message not found. Output:\n%s", output)
	}
}

func TestKkInit_WithSeaweedFS(t *testing.T) {
	kkPath := ensureKkBinary(t)
	tempDir := t.TempDir()

	originalValidator := cmd.DockerValidatorInstance
	defer func() { cmd.DockerValidatorInstance = originalValidator }()
	cmd.DockerValidatorInstance = mockDockerValidator(true, true) // Ensure Docker is seen as working

	// Simulate user input:
	// 1. No overwrite
	// 2. Enable SeaweedFS (y)
	// 3. No Caddy (n)
	input := strings.NewReader("y\nn\n")

	cmdExec := exec.Command(kkPath, "init")
	cmdExec.Dir = tempDir
	cmdExec.Stdin = input
	output, err := cmdExec.CombinedOutput()
	if err != nil {
		t.Fatalf("kk init failed with SeaweedFS: %v\nOutput:\n%s", err, output)
	}

	// Verify kkfiler.toml is created
	if _, err := os.Stat(filepath.Join(tempDir, "kkfiler.toml")); os.IsNotExist(err) {
		t.Errorf("Expected kkfiler.toml not created with SeaweedFS enabled")
	}
	if !strings.Contains(string(output), "Da tao: kkfiler.toml") {
		t.Errorf("Expected 'Da tao: kkfiler.toml' message not found. Output:\n%s", output)
	}
}

func TestKkInit_WithCaddy(t *testing.T) {
	kkPath := ensureKkBinary(t)
	tempDir := t.TempDir()

	originalValidator := cmd.DockerValidatorInstance
	defer func() { cmd.DockerValidatorInstance = originalValidator }()
	cmd.DockerValidatorInstance = mockDockerValidator(true, true) // Ensure Docker is seen as working

	// Simulate user input:
	// 1. No overwrite
	// 2. No SeaweedFS (n)
	// 3. Enable Caddy (y)
	// 4. Domain: mydomain.com
	input := strings.NewReader("n\ny\nmydomain.com\n")

	cmdExec := exec.Command(kkPath, "init")
	cmdExec.Dir = tempDir
	cmdExec.Stdin = input
	output, err := cmdExec.CombinedOutput()
	if err != nil {
		t.Fatalf("kk init failed with Caddy: %v\nOutput:\n%s", err, output)
	}

	// Verify Caddyfile is created
	if _, err := os.Stat(filepath.Join(tempDir, "Caddyfile")); os.IsNotExist(err) {
		t.Errorf("Expected Caddyfile not created with Caddy enabled")
	}
	if !strings.Contains(string(output), "Da tao: Caddyfile") {
		t.Errorf("Expected 'Da tao: Caddyfile' message not found. Output:\n%s", output)
	}
	// Verify Caddyfile content contains the domain
	caddyContent, err := os.ReadFile(filepath.Join(tempDir, "Caddyfile"))
	if err != nil {
		t.Fatalf("Failed to read Caddyfile: %v", err)
	}
	if !strings.Contains(string(caddyContent), "caddy config for mydomain.com") {
		t.Errorf("Caddyfile content mismatch. Got:\n%s", string(caddyContent))
	}
}

func TestKkInit_OverwriteExistingCompose(t *testing.T) {
	kkPath := ensureKkBinary(t)
	tempDir := t.TempDir()

	originalValidator := cmd.DockerValidatorInstance
	defer func() { cmd.DockerValidatorInstance = originalValidator }()
	cmd.DockerValidatorInstance = mockDockerValidator(true, true) // Ensure Docker is seen as working

	// Create a dummy docker-compose.yml file
	dummyComposePath := filepath.Join(tempDir, "docker-compose.yml")
	err := ioutil.WriteFile(dummyComposePath, []byte("existing compose"), 0644)
	if err != nil {
		t.Fatalf("Failed to create dummy docker-compose.yml: %v", err)
	}

	// Simulate user input:
	// 1. Overwrite (y)
	// 2. No SeaweedFS (n)
	// 3. No Caddy (n)
	input := strings.NewReader("y\nn\nn\n")

	cmdExec := exec.Command(kkPath, "init")
	cmdExec.Dir = tempDir
	cmdExec.Stdin = input
	output, err := cmdExec.CombinedOutput()
	if err != nil {
		t.Fatalf("kk init failed during overwrite test: %v\nOutput:\n%s", err, output)
	}

	// Verify backup file is created and contains original content
	backupPath := dummyComposePath + ".bak"
	backupContent, err := ioutil.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("Failed to read backup file: %v", err)
	}
	if string(backupContent) != "existing compose" {
		t.Errorf("Backup file content mismatch. Got: %q, Want: %q", string(backupContent), "existing compose")
	}

	// Verify new docker-compose.yml is created and contains new content (check for password placeholders)
	newComposeContent, err := ioutil.ReadFile(dummyComposePath)
	if err != nil {
		t.Fatalf("Failed to read new docker-compose.yml: %v", err)
	}
	if !strings.Contains(string(newComposeContent), "MYSQL_PASSWORD") { // Check for password placeholders
		t.Errorf("New docker-compose.yml content mismatch (missing password placeholder). Got:\n%s", string(newComposeContent))
	}
	if strings.Contains(string(newComposeContent), "existing compose") {
		t.Errorf("New docker-compose.yml still contains old content after overwrite.")
	}
}

func TestKkInit_NoOverwriteExistingCompose(t *testing.T) {
	kkPath := ensureKkBinary(t)
	tempDir := t.TempDir()

	originalValidator := cmd.DockerValidatorInstance
	defer func() { cmd.DockerValidatorInstance = originalValidator }()
	cmd.DockerValidatorInstance = mockDockerValidator(true, true) // Ensure Docker is seen as working

	// Create a dummy docker-compose.yml file
	dummyComposePath := filepath.Join(tempDir, "docker-compose.yml")
	err := ioutil.WriteFile(dummyComposePath, []byte("original content that should remain"), 0644)
	if err != nil {
		t.Fatalf("Failed to create dummy docker-compose.yml: %v", err)
	}

	// Simulate user input:
	// 1. Do NOT overwrite (n)
	input := strings.NewReader("n\n")

	cmdExec := exec.Command(kkPath, "init")
	cmdExec.Dir = tempDir
	cmdExec.Stdin = input
	output, err := cmdExec.CombinedOutput()
	if err == nil {
		t.Fatalf("kk init did not return an error when user chose not to overwrite. Output:\n%s", output)
	}
	// The original cobra error message is "Error: Initialization cancelled".
	// The exit status 1 comes from the application itself returning error,
	// not directly from exec.Command.
	// So, we just check for the specific message in the output.
	if !strings.Contains(string(output), "Initialization cancelled") {
		t.Errorf("Expected 'Initialization cancelled' message not found. Output:\n%s", output)
	}

	// Verify the original file content remains unchanged
	finalComposeContent, err := ioutil.ReadFile(dummyComposePath)
	if err != nil {
		t.Fatalf("Failed to read docker-compose.yml after no-overwrite: %v", err)
	}
	if string(finalComposeContent) != "original content that should remain" {
		t.Errorf("docker-compose.yml content changed after no-overwrite. Got: %q", string(finalComposeContent))
	}
}
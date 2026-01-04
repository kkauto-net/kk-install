package templates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderTemplate(t *testing.T) {
	tempDir := t.TempDir() // Add this line
	testTmplName := "docker-compose.yml" // Use an existing embedded template
	outputPath := filepath.Join(tempDir, "test_output.yml")

	cfg := Config{
		DBPassword: "testdbpassword",
		DBRootPassword: "testdbrootpassword",
		RedisPassword: "testredispassword",
		Domain: "test.com",
	}

	// Test 1: Happy path - render to a new file
	err := RenderTemplate(testTmplName, cfg, outputPath)
	if err != nil {
		t.Fatalf("RenderTemplate failed: %v", err)
	}
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read rendered file: %v", err)
	}
	if !strings.Contains(string(content), "MYSQL_PASSWORD: testdbpassword") {
		t.Errorf("Rendered content mismatch for DB_PASSWORD. Got:\n%s", string(content))
	}
	if !strings.Contains(string(content), "MYSQL_ROOT_PASSWORD: testdbrootpassword") {
		t.Errorf("Rendered content mismatch for DB_ROOT_PASSWORD. Got:\n%s", string(content))
	}
	if !strings.Contains(string(content), "redis-server --requirepass testredispassword") { // Corrected assertion
		t.Errorf("Rendered content missing REDIS_PASSWORD. Got:\n%s", string(content))
	}

	// Test 2: Backup existing file
	err = os.WriteFile(outputPath, []byte("Original compose content"), 0644)
	if err != nil {
		t.Fatalf("Failed to write original file for backup test: %v", err)
	}
	err = RenderTemplate(testTmplName, cfg, outputPath)
	if err != nil {
		t.Fatalf("RenderTemplate with backup failed: %v", err)
	}
	backupPath := outputPath + ".bak"
	backupContent, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("Failed to read backup file: %v", err)
	}
	if string(backupContent) != "Original compose content" {
		t.Errorf("Backup content mismatch. Got: %q, Want: %q", string(backupContent), "Original compose content")
	}
	newContent, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read new file after backup: %v", err)
	}
	if !strings.Contains(string(newContent), "MYSQL_PASSWORD: testdbpassword") {
		t.Errorf("New file content after backup mismatch for DB_PASSWORD. Got:\n%s", string(newContent))
	}
	if !strings.Contains(string(newContent), "MYSQL_ROOT_PASSWORD: testdbrootpassword") {
		t.Errorf("New file content after backup mismatch for DB_ROOT_PASSWORD. Got:\n%s", string(newContent))
	}
	if !strings.Contains(string(newContent), "redis-server --requirepass testredispassword") {
		t.Errorf("New file content after backup missing REDIS_PASSWORD. Got:\n%s", string(newContent))
	}

	// Test 3: Template not found (should return an error)
	err = RenderTemplate("non_existent_template", cfg, filepath.Join(tempDir, "no_such_file.txt"))
	if err == nil {
		t.Errorf("RenderTemplate for non-existent template did not return an error")
	}
	if !strings.Contains(err.Error(), "no such file or directory") && !strings.Contains(err.Error(), "asset not found") && !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("Expected 'no such file or directory' or 'asset not found' error, got: %v", err)
	}
}
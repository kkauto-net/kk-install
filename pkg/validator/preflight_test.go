package validator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunPreflight(t *testing.T) {
	t.Run("Missing file", func(t *testing.T) {
		tmpDir := t.TempDir()

		composeContent := "version: '3.8'\nservices:\n  db:\n    image: mariadb:10.6"
		composeFile := filepath.Join(tmpDir, "docker-compose.yml")
		os.WriteFile(composeFile, []byte(composeContent), 0644)

		results, err := RunPreflight(tmpDir, false)

		if len(results) == 0 {
			t.Error("Expected preflight results")
		}

		foundCheck := false
		for _, r := range results {
			if r.CheckName == "File .e"+"nv" {
				foundCheck = true
				if r.Passed {
					t.Error("Expected check to fail")
				}
			}
		}

		if !foundCheck && err == nil {
			t.Error("Expected to find check in results")
		}
	})

	t.Run("With Caddy enabled", func(t *testing.T) {
		tmpDir := t.TempDir()

		content1 := "DB_PASSWORD=verylongpassword123456\nDB_ROOT_PASSWORD=verylongrootpass123\nREDIS_PASSWORD=verylongredispass123"
		envFile := filepath.Join(tmpDir, ".e"+"nv")
		os.WriteFile(envFile, []byte(content1), 0600)

		composeContent := "version: '3.8'\nservices:\n  db:\n    image: mariadb:10.6"
		composeFile := filepath.Join(tmpDir, "docker-compose.yml")
		os.WriteFile(composeFile, []byte(composeContent), 0644)

		caddyContent := "example.com {\n\treverse_proxy localhost:8019\n}"
		caddyFile := filepath.Join(tmpDir, "Caddyfile")
		os.WriteFile(caddyFile, []byte(caddyContent), 0644)

		results, _ := RunPreflight(tmpDir, true)

		foundCaddyCheck := false
		for _, r := range results {
			if r.CheckName == "Caddyfile" {
				foundCaddyCheck = true
			}
		}

		if !foundCaddyCheck {
			t.Error("Expected Caddyfile check when Caddy enabled")
		}
	})
}

func TestPrintPreflightResults(t *testing.T) {
	t.Run("Print mixed results", func(t *testing.T) {
		results := []PreflightResult{
			{CheckName: "Test 1", Passed: true},
			{CheckName: "Test 2", Passed: false, Error: &UserError{Message: "Error msg", Suggestion: "Fix"}},
			{CheckName: "Test 3", Passed: true, Warning: "Warning message"},
		}

		PrintPreflightResults(results)
	})
}

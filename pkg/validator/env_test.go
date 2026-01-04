package validator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateEnvFile(t *testing.T) {
	t.Run("Missing env file", func(t *testing.T) {
		tmpDir := t.TempDir()
		err := ValidateEnvFile(tmpDir)
		if err == nil {
			t.Error("Expected error for missing file")
		}
		if ue, ok := err.(*UserError); ok {
			if ue.Key != "env_missing" {
				t.Errorf("Expected error key 'env_missing', got %q", ue.Key)
			}
		}
	})

	t.Run("Valid file", func(t *testing.T) {
		tmpDir := t.TempDir()
		content := "DB_PASSWORD=verylongpassword123456\nDB_ROOT_PASSWORD=verylongrootpass123\nREDIS_PASSWORD=verylongredispass123"
		envFile := filepath.Join(tmpDir, ".e"+"nv")
		os.WriteFile(envFile, []byte(content), 0600)

		err := ValidateEnvFile(tmpDir)
		if err != nil {
			t.Errorf("Expected no error for valid file, got %v", err)
		}
	})

	t.Run("Missing required vars", func(t *testing.T) {
		tmpDir := t.TempDir()
		content := "DB_PASSWORD=test123456789012"
		envFile := filepath.Join(tmpDir, ".e"+"nv")
		os.WriteFile(envFile, []byte(content), 0600)

		err := ValidateEnvFile(tmpDir)
		if err == nil {
			t.Error("Expected error for missing required vars")
		}
		if ue, ok := err.(*UserError); ok {
			if ue.Key != "env_missing_vars" {
				t.Errorf("Expected error key 'env_missing_vars', got %q", ue.Key)
			}
		}
	})
}

func TestParseEnvFile(t *testing.T) {
	t.Run("Parse valid file", func(t *testing.T) {
		tmpDir := t.TempDir()
		content := "KEY1=value1\nKEY2=\"value2\"\n# Comment line\nKEY3='value3'"
		envPath := filepath.Join(tmpDir, ".e"+"nv")
		os.WriteFile(envPath, []byte(content), 0600)

		vars, err := parseEnvFile(envPath)
		if err != nil {
			t.Fatalf("parseEnvFile failed: %v", err)
		}

		if vars["KEY1"] != "value1" {
			t.Errorf("Expected KEY1=value1, got %q", vars["KEY1"])
		}
		if vars["KEY2"] != "value2" {
			t.Errorf("Expected KEY2=value2, got %q", vars["KEY2"])
		}
	})
}

func TestCheckEnvPermissions(t *testing.T) {
	t.Run("Non-existent file", func(t *testing.T) {
		tmpDir := t.TempDir()
		CheckEnvPermissions(tmpDir)
	})
}

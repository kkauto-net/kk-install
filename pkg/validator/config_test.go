package validator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateDockerCompose(t *testing.T) {
	t.Run("Missing docker-compose.yml", func(t *testing.T) {
		tmpDir := t.TempDir()
		err := ValidateDockerCompose(tmpDir)
		if err == nil {
			t.Error("Expected error for missing docker-compose.yml")
		}
		if ue, ok := err.(*UserError); ok {
			if ue.Key != "compose_missing" {
				t.Errorf("Expected error key 'compose_missing', got %q", ue.Key)
			}
		}
	})

	t.Run("Valid docker-compose.yml", func(t *testing.T) {
		tmpDir := t.TempDir()
		content := `version: '3.8'
services:
  db:
    image: mariadb:10.6
    ports:
      - "3307:3306"`
		os.WriteFile(filepath.Join(tmpDir, "docker-compose.yml"), []byte(content), 0644)

		err := ValidateDockerCompose(tmpDir)
		if err != nil {
			t.Errorf("Expected no error for valid docker-compose.yml, got %v", err)
		}
	})

	t.Run("Missing services section", func(t *testing.T) {
		tmpDir := t.TempDir()
		content := `version: '3.8'
networks:
  mynetwork:`
		os.WriteFile(filepath.Join(tmpDir, "docker-compose.yml"), []byte(content), 0644)

		err := ValidateDockerCompose(tmpDir)
		if err == nil {
			t.Error("Expected error for missing services section")
		}
		if ue, ok := err.(*UserError); ok {
			if ue.Key != "compose_no_services" {
				t.Errorf("Expected error key 'compose_no_services', got %q", ue.Key)
			}
		}
	})
}

func TestValidateCaddyfile(t *testing.T) {
	t.Run("Missing Caddyfile (optional)", func(t *testing.T) {
		tmpDir := t.TempDir()
		err := ValidateCaddyfile(tmpDir)
		if err != nil {
			t.Errorf("Expected no error for missing optional Caddyfile, got %v", err)
		}
	})

	t.Run("Valid Caddyfile", func(t *testing.T) {
		tmpDir := t.TempDir()
		content := `example.com {
	reverse_proxy localhost:8019
}`
		os.WriteFile(filepath.Join(tmpDir, "Caddyfile"), []byte(content), 0644)

		err := ValidateCaddyfile(tmpDir)
		if err != nil {
			t.Errorf("Expected no error for valid Caddyfile, got %v", err)
		}
	})

	t.Run("Empty Caddyfile", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.WriteFile(filepath.Join(tmpDir, "Caddyfile"), []byte(""), 0644)

		err := ValidateCaddyfile(tmpDir)
		if err == nil {
			t.Error("Expected error for empty Caddyfile")
		}
		if ue, ok := err.(*UserError); ok {
			if ue.Key != "caddy_empty" {
				t.Errorf("Expected error key 'caddy_empty', got %q", ue.Key)
			}
		}
	})
}

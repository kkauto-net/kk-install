package validator

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ValidateDockerCompose checks docker-compose.yml syntax
func ValidateDockerCompose(dir string) error {
	composePath := filepath.Join(dir, "docker-compose.yml")

	if _, err := os.Stat(composePath); os.IsNotExist(err) {
		return &UserError{
			Key:        "compose_missing",
			Message:    "File docker-compose.yml khong ton tai",
			Suggestion: "Chay: kk init",
		}
	}

	content, err := os.ReadFile(composePath)
	if err != nil {
		return &UserError{
			Key:        "compose_read_error",
			Message:    fmt.Sprintf("Khong doc duoc docker-compose.yml: %v", err),
			Suggestion: "Kiem tra quyen truy cap file",
		}
	}

	// Parse YAML to validate syntax
	var compose map[string]interface{}
	if err := yaml.Unmarshal(content, &compose); err != nil {
		return &UserError{
			Key:        "compose_syntax_error",
			Message:    fmt.Sprintf("Loi cu phap docker-compose.yml: %v", err),
			Suggestion: "Kiem tra cu phap YAML (indentation, colons, quotes)",
		}
	}

	// Check required sections
	if _, ok := compose["services"]; !ok {
		return &UserError{
			Key:        "compose_no_services",
			Message:    "docker-compose.yml thieu section 'services'",
			Suggestion: "Them section services vao file",
		}
	}

	return nil
}

// ValidateCaddyfile does basic Caddyfile syntax check
func ValidateCaddyfile(dir string) error {
	caddyPath := filepath.Join(dir, "Caddyfile")

	if _, err := os.Stat(caddyPath); os.IsNotExist(err) {
		// Caddyfile is optional
		return nil
	}

	content, err := os.ReadFile(caddyPath)
	if err != nil {
		return &UserError{
			Key:        "caddy_read_error",
			Message:    fmt.Sprintf("Khong doc duoc Caddyfile: %v", err),
			Suggestion: "Kiem tra quyen truy cap file",
		}
	}

	// Basic check: file should not be empty if exists
	if len(content) == 0 {
		return &UserError{
			Key:        "caddy_empty",
			Message:    "Caddyfile trong",
			Suggestion: "Them cau hinh domain vao Caddyfile",
		}
	}

	return nil
}

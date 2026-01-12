package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	configDirName  = ".kk"
	configFileName = "config.yaml"
)

// Config represents user configuration
type Config struct {
	Language   string `yaml:"language"`    // "en" or "vi"
	ProjectDir string `yaml:"project_dir"` // Path to project with docker-compose.yml
}

// ConfigDir returns the config directory path (~/.kk)
func ConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, configDirName)
}

// ConfigPath returns the full config file path
func ConfigPath() string {
	return filepath.Join(ConfigDir(), configFileName)
}

// Load reads config from disk, returns defaults if not exists
func Load() (*Config, error) {
	cfg := &Config{Language: "en"} // default to English

	data, err := os.ReadFile(ConfigPath())
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // Return default
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// Validate language, default to English if invalid
	if cfg.Language != "en" && cfg.Language != "vi" {
		cfg.Language = "en"
	}

	return cfg, nil
}

// Save writes config to disk
func (c *Config) Save() error {
	// Create dir if not exists
	if err := os.MkdirAll(ConfigDir(), 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(ConfigPath(), data, 0644)
}

// EnsureProjectDir validates and changes to the configured project directory.
// Returns the project directory path or error if not configured/invalid.
// If the project directory is invalid, it will be cleared from config.
func EnsureProjectDir() (string, error) {
	cfg, err := Load()
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.ProjectDir == "" {
		return "", errors.New("no project configured, run 'kk init' first")
	}

	projectDir := cfg.ProjectDir

	// Check directory exists
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		// Clear invalid config and save
		cfg.ProjectDir = ""
		if saveErr := cfg.Save(); saveErr != nil {
			// Log but don't fail - the main error is more important
			fmt.Fprintf(os.Stderr, "Warning: failed to clear invalid config: %v\n", saveErr)
		}
		return "", fmt.Errorf("project directory no longer exists: %s", projectDir)
	}

	// Check docker-compose.yml exists
	composePath := filepath.Join(projectDir, "docker-compose.yml")
	if _, err := os.Stat(composePath); os.IsNotExist(err) {
		cfg.ProjectDir = ""
		if saveErr := cfg.Save(); saveErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to clear invalid config: %v\n", saveErr)
		}
		return "", fmt.Errorf("docker-compose.yml not found in: %s", projectDir)
	}

	// Change to project directory
	if err := os.Chdir(projectDir); err != nil {
		return "", fmt.Errorf("failed to change to project directory %s: %w", projectDir, err)
	}

	return projectDir, nil
}

// ReadEnvValue reads a specific key from the .env file in projectDir.
// Returns empty string if file doesn't exist or key not found.
func ReadEnvValue(projectDir, key string) string {
	envPath := filepath.Join(projectDir, ".env")
	data, err := os.ReadFile(envPath)
	if err != nil {
		return ""
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if idx := strings.Index(line, "="); idx > 0 {
			k := strings.TrimSpace(line[:idx])
			if k == key {
				return strings.TrimSpace(line[idx+1:])
			}
		}
	}
	return ""
}

// IsProjectNotConfiguredError returns true if the error is about project not being configured
func IsProjectNotConfiguredError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return msg == "no project configured, run 'kk init' first" ||
		strings.Contains(msg, "project directory no longer exists") ||
		strings.Contains(msg, "docker-compose.yml not found in")
}

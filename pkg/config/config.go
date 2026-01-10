package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	configDirName  = ".kk"
	configFileName = "config.yaml"
)

// Config represents user configuration
type Config struct {
	Language string `yaml:"language"` // "en" or "vi"
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

// Package n8n provides configuration and template rendering for n8n workflow automation.
package n8n

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kkauto-net/kk-install/pkg/config"
)

const (
	// N8nSubDir is the subdirectory name for n8n within ProjectDir
	N8nSubDir = "n8n"

	// MinEncryptionKeyLength is the minimum length for the encryption key
	MinEncryptionKeyLength = 32

	// MinDBPasswordLength is the minimum length for database password
	MinDBPasswordLength = 16
)

// N8nConfig holds n8n installation configuration
type N8nConfig struct {
	Domain          string // n8n domain (e.g., n8n.example.com)
	N8nHost         string // N8N_HOST env value
	DBUser          string // PostgreSQL username
	DBPassword      string // PostgreSQL password
	EncryptionKey   string // N8N_ENCRYPTION_KEY (critical - never lose this!)
	Timezone        string // Timezone (default: Asia/Ho_Chi_Minh)
	ConnectKKEngine bool   // Whether to join kkengine_net network
}

// Validate checks that all required fields meet minimum security requirements.
func (c N8nConfig) Validate() error {
	if len(c.EncryptionKey) < MinEncryptionKeyLength {
		return fmt.Errorf("encryption key must be at least %d characters (got %d)",
			MinEncryptionKeyLength, len(c.EncryptionKey))
	}
	if len(c.DBPassword) < MinDBPasswordLength {
		return fmt.Errorf("database password must be at least %d characters (got %d)",
			MinDBPasswordLength, len(c.DBPassword))
	}
	if c.DBUser == "" {
		return errors.New("database user is required")
	}
	if c.N8nHost == "" {
		return errors.New("n8n host is required")
	}
	return nil
}

// N8nDir returns the n8n installation directory.
// Uses ProjectDir from config, fallback to ~/.kk/n8n if ProjectDir not set.
func N8nDir() string {
	cfg, err := config.Load()
	if err == nil && cfg.ProjectDir != "" {
		return filepath.Join(cfg.ProjectDir, N8nSubDir)
	}
	// Fallback to ~/.kk/n8n
	home, err := os.UserHomeDir()
	if err != nil {
		// Ultimate fallback to /tmp for edge cases
		return filepath.Join("/tmp", ".kk", N8nSubDir)
	}
	return filepath.Join(home, ".kk", N8nSubDir)
}

// DataDir returns the n8n data directory
func DataDir() string {
	return filepath.Join(N8nDir(), "data")
}

// PostgresDir returns the PostgreSQL data directory
func PostgresDir() string {
	return filepath.Join(N8nDir(), "postgres")
}

// ComposePath returns the path to docker-compose.yml
func ComposePath() string {
	return filepath.Join(N8nDir(), "docker-compose.yml")
}

// EnvPath returns the path to .env file
func EnvPath() string {
	return filepath.Join(N8nDir(), ".env")
}

// IsInstalled checks if n8n is already installed
func IsInstalled() bool {
	_, err := os.Stat(ComposePath())
	return err == nil
}

// EnsureDirectories creates required directories for n8n installation
func EnsureDirectories() error {
	dirs := []string{N8nDir(), DataDir(), PostgresDir()}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

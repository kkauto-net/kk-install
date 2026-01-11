package templates

import (
	"embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed *.tmpl
var templateFS embed.FS // Force recompile

type Config struct {
	// Services
	EnableSeaweedFS bool
	EnableCaddy     bool

	// System
	Domain    string
	JWTSecret string

	// License
	LicenseKey      string
	ServerPublicKey string

	// Database
	DBPassword     string
	DBRootPassword string
	RedisPassword  string

	// S3 (only used when EnableSeaweedFS)
	S3AccessKey string
	S3SecretKey string
}

// Secret length requirements
const (
	MinJWTSecretLength   = 32 // OWASP recommended minimum for HMAC secrets
	MinDBPasswordLength  = 16
	MinS3AccessKeyLength = 16
	MinS3SecretKeyLength = 32
)

// ValidateSecrets validates that all secrets meet minimum security requirements
func (c Config) ValidateSecrets() error {
	if len(c.JWTSecret) < MinJWTSecretLength {
		return fmt.Errorf("JWT_SECRET must be at least %d characters (got %d)", MinJWTSecretLength, len(c.JWTSecret))
	}
	if len(c.DBPassword) < MinDBPasswordLength {
		return fmt.Errorf("DB_PASSWORD must be at least %d characters (got %d)", MinDBPasswordLength, len(c.DBPassword))
	}
	if len(c.DBRootPassword) < MinDBPasswordLength {
		return fmt.Errorf("DB_ROOT_PASSWORD must be at least %d characters (got %d)", MinDBPasswordLength, len(c.DBRootPassword))
	}
	if len(c.RedisPassword) < MinDBPasswordLength {
		return fmt.Errorf("REDIS_PASSWORD must be at least %d characters (got %d)", MinDBPasswordLength, len(c.RedisPassword))
	}
	if c.EnableSeaweedFS {
		if len(c.S3AccessKey) < MinS3AccessKeyLength {
			return fmt.Errorf("S3_ACCESS_KEY must be at least %d characters (got %d)", MinS3AccessKeyLength, len(c.S3AccessKey))
		}
		if len(c.S3SecretKey) < MinS3SecretKeyLength {
			return fmt.Errorf("S3_SECRET_KEY must be at least %d characters (got %d)", MinS3SecretKeyLength, len(c.S3SecretKey))
		}
	}
	return nil
}

// RenderTemplate renders a single template file
func RenderTemplate(name string, cfg Config, outputPath string) error {
	tmplContent, err := templateFS.ReadFile(name + ".tmpl")
	if err != nil {
		return err
	}

	tmpl, err := template.New(name).Parse(string(tmplContent))
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Backup existing file if it exists
	if _, err := os.Stat(outputPath); err == nil {
		backupPath := outputPath + ".bak"
		if err := os.Rename(outputPath, backupPath); err != nil {
			return err
		}
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, cfg)
}

// RenderAll renders all templates to the target directory
func RenderAll(cfg Config, targetDir string) error {
	// Validate secrets before rendering
	if err := cfg.ValidateSecrets(); err != nil {
		return errors.New("invalid config: " + err.Error())
	}

	files := map[string]string{
		"docker-compose.yml": "docker-compose.yml",
		"env":                ".env",
		"kkphp.conf":         "kkphp.conf",
	}

	if cfg.EnableCaddy {
		files["Caddyfile"] = "Caddyfile"
	}
	if cfg.EnableSeaweedFS {
		files["kkfiler.toml"] = "kkfiler.toml"
	}

	for tmplName, outputName := range files {
		outputPath := filepath.Join(targetDir, outputName)
		if err := RenderTemplate(tmplName, cfg, outputPath); err != nil {
			return err
		}
	}

	// Set .env permissions to 0600 (owner read/write only)
	envPath := filepath.Join(targetDir, ".env")
	if err := os.Chmod(envPath, 0600); err != nil {
		return err
	}

	return nil
}

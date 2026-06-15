package validator

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/kkauto-net/kk-install/pkg/ui"
)

// RequiredEnvVars lists mandatory environment variables
var RequiredEnvVars = []string{
	"DB_PASSWORD",
	"DB_ROOT_PASSWORD",
	"REDIS_PASSWORD",
}

// OptionalEnvVars lists optional environment variables with defaults
var OptionalEnvVars = map[string]string{
	"DB_HOSTNAME": "db",
	"DB_PORT":     "3306",
	"DB_DATABASE": "kkengine",
	"DB_USERNAME": "kkengine",
	"REDIS_HOST":  "redis",
	"REDIS_PORT":  "6379",
}

// ValidateEnvFile checks .env file exists and contains required vars
func ValidateEnvFile(dir string) error {
	envPath := filepath.Join(dir, ".env")

	// Check file exists
	info, err := os.Stat(envPath)
	if os.IsNotExist(err) {
		return &UserError{
			Key: ErrEnvMissing,
		}
	}
	if err != nil {
		return &UserError{
			Key: "env_stat_error",
		}
	}

	// Check file permissions (warn if too permissive)
	mode := info.Mode()
	if mode.Perm()&0044 != 0 { // Readable by group or others
		ui.ShowWarningf(ui.Msg("warn_env_permissions"), mode.Perm())
		ui.ShowNote(ui.Msg("warn_env_permissions_fix"))
	}

	// Parse .env file
	envVars, err := parseEnvFile(envPath)
	if err != nil {
		return &UserError{
			Key: "env_parse_error",
		}
	}

	// Check required vars
	var missing []string
	for _, key := range RequiredEnvVars {
		if val, ok := envVars[key]; !ok || val == "" {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		return &UserError{
			Key:  ErrEnvMissingVars,
			Args: []any{strings.Join(missing, ", ")},
		}
	}

	// Check password strength (minimum 16 chars)
	passwordVars := []string{"DB_PASSWORD", "DB_ROOT_PASSWORD", "REDIS_PASSWORD"}
	var weakPasswords []string
	for _, key := range passwordVars {
		if val, ok := envVars[key]; ok && len(val) < 16 {
			weakPasswords = append(weakPasswords, key)
		}
	}

	if len(weakPasswords) > 0 {
		ui.ShowWarningf(ui.Msg("warn_weak_password"), strings.Join(weakPasswords, ", "))
	}

	return nil
}

func parseEnvFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer closeFile(file)

	vars := make(map[string]string)
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue // Skip malformed lines
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		value = strings.Trim(value, "\"'")

		vars[key] = value
	}

	return vars, scanner.Err()
}

// CheckEnvPermissions warns if .env is world-readable
func CheckEnvPermissions(dir string) {
	envPath := filepath.Join(dir, ".env")
	info, err := os.Stat(envPath)
	if err != nil {
		return
	}

	mode := info.Mode()
	// Check if others have read permission (Unix)
	if mode&0004 != 0 {
		ui.ShowWarning(ui.Msg("warn_env_world_readable"))
		ui.ShowNote(ui.MsgF("warn_env_world_readable_fix", envPath))
	}
}

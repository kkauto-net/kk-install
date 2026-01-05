package validator

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
			Key:        "env_missing",
			Message:    "File .env khong ton tai",
			Suggestion: "Chay: kk init",
		}
	}
	if err != nil {
		return &UserError{
			Key:        "env_stat_error",
			Message:    fmt.Sprintf("Loi doc thong tin file .env: %v", err),
			Suggestion: "Kiem tra quyen truy cap file",
		}
	}

	// Check file permissions (warn if too permissive)
	mode := info.Mode()
	if mode.Perm()&0044 != 0 { // Readable by group or others
		fmt.Printf("  [!] Canh bao: File .env co quyen truy cap qua rong (%o)\n", mode.Perm())
		fmt.Printf("      Nen thiet lap: chmod 600 .env (chi user hien tai doc/ghi)\n")
	}

	// Parse .env file
	envVars, err := parseEnvFile(envPath)
	if err != nil {
		return &UserError{
			Key:        "env_parse_error",
			Message:    fmt.Sprintf("Loi doc file .env: %v", err),
			Suggestion: "Kiem tra cu phap file .env",
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
			Key:        "env_missing_vars",
			Message:    "Thieu bien moi truong trong .env",
			Suggestion: fmt.Sprintf("Them vao .env: %s", strings.Join(missing, ", ")),
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
		// Warning only, don't block
		fmt.Printf("  [!] Canh bao: Mat khau yeu cho: %s (nen >= 16 ky tu)\n",
			strings.Join(weakPasswords, ", "))
	}

	return nil
}

func parseEnvFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

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
		fmt.Printf("  [!] Canh bao: File .env co the doc boi nguoi khac.\n")
		fmt.Printf("      Chay: chmod 600 %s\n", envPath)
	}
}

package templates

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v3"
)

func TestRenderTemplate(t *testing.T) {
	tempDir := t.TempDir()               // Add this line
	testTmplName := "docker-compose.yml" // Use an existing embedded template
	outputPath := filepath.Join(tempDir, "test_output.yml")

	cfg := Config{
		EnableSeaweedFS: true, // Enable all optional services for full test coverage
		EnableCaddy:     true,
		DBPassword:      "testdbpassword",
		DBRootPassword:  "testdbrootpassword",
		RedisPassword:   "testredispassword",
		Domain:          "test.com",
	}

	// Test 1: Happy path - render to a new file
	err := RenderTemplate(testTmplName, cfg, outputPath)
	if err != nil {
		t.Fatalf("RenderTemplate failed: %v", err)
	}
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read rendered file: %v", err)
	}
	// Verify env var substitution (not hardcoded passwords)
	if !strings.Contains(string(content), "MYSQL_PASSWORD: ${DB_PASSWORD}") {
		t.Errorf("Rendered content should use ${DB_PASSWORD} env var. Got:\n%s", string(content))
	}
	if !strings.Contains(string(content), "MYSQL_ROOT_PASSWORD: ${DB_ROOT_PASSWORD}") {
		t.Errorf("Rendered content should use ${DB_ROOT_PASSWORD} env var. Got:\n%s", string(content))
	}
	if !strings.Contains(string(content), "redis-server --requirepass ${REDIS_PASSWORD}") {
		t.Errorf("Rendered content should use ${REDIS_PASSWORD} env var. Got:\n%s", string(content))
	}

	// Test 2: Backup existing file
	err = os.WriteFile(outputPath, []byte("Original compose content"), 0644)
	if err != nil {
		t.Fatalf("Failed to write original file for backup test: %v", err)
	}
	err = RenderTemplate(testTmplName, cfg, outputPath)
	if err != nil {
		t.Fatalf("RenderTemplate with backup failed: %v", err)
	}
	backupPath := outputPath + ".bak"
	backupContent, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("Failed to read backup file: %v", err)
	}
	if string(backupContent) != "Original compose content" {
		t.Errorf("Backup content mismatch. Got: %q, Want: %q", string(backupContent), "Original compose content")
	}
	newContent, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read new file after backup: %v", err)
	}
	// Verify env var substitution after backup
	if !strings.Contains(string(newContent), "MYSQL_PASSWORD: ${DB_PASSWORD}") {
		t.Errorf("New file should use ${DB_PASSWORD} env var. Got:\n%s", string(newContent))
	}
	if !strings.Contains(string(newContent), "MYSQL_ROOT_PASSWORD: ${DB_ROOT_PASSWORD}") {
		t.Errorf("New file should use ${DB_ROOT_PASSWORD} env var. Got:\n%s", string(newContent))
	}
	if !strings.Contains(string(newContent), "redis-server --requirepass ${REDIS_PASSWORD}") {
		t.Errorf("New file should use ${REDIS_PASSWORD} env var. Got:\n%s", string(newContent))
	}

	// Test 3: Template not found (should return an error)
	err = RenderTemplate("non_existent_template", cfg, filepath.Join(tempDir, "no_such_file.txt"))
	if err == nil {
		t.Errorf("RenderTemplate for non-existent template did not return an error")
	}
	if !strings.Contains(err.Error(), "no such file or directory") && !strings.Contains(err.Error(), "asset not found") && !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("Expected 'no such file or directory' or 'asset not found' error, got: %v", err)
	}
}

// TestAllTemplatesExist verifies all required templates are embedded
func TestAllTemplatesExist(t *testing.T) {
	required := []string{
		"Caddyfile.tmpl",
		"kkfiler.toml.tmpl",
		"kkphp.conf.tmpl",
		"docker-compose.yml.tmpl",
		"env.tmpl",
	}
	for _, name := range required {
		_, err := templateFS.ReadFile(name)
		if err != nil {
			t.Errorf("template %s not found: %v", name, err)
		}
	}
}

// TestAllTemplatesParseable verifies templates can be parsed
func TestAllTemplatesParseable(t *testing.T) {
	err := fs.WalkDir(templateFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".tmpl") {
			return nil
		}

		// Try to parse template
		_, parseErr := RenderTemplateToString(strings.TrimSuffix(path, ".tmpl"), Config{
			EnableSeaweedFS: true,
			EnableCaddy:     true,
			DBPassword:      "test",
			DBRootPassword:  "test",
			RedisPassword:   "test",
			Domain:          "test.com",
		})
		if parseErr != nil {
			t.Errorf("template %s failed to parse: %v", path, parseErr)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir failed: %v", err)
	}
}

// Helper function to render template to string
func RenderTemplateToString(name string, cfg Config) (string, error) {
	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, name+"_test")
	defer os.Remove(tempFile)

	err := RenderTemplate(name, cfg, tempFile)
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(tempFile)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// TestAllConfigCombinations tests all EnableSeaweedFS/EnableCaddy combinations
func TestAllConfigCombinations(t *testing.T) {
	combinations := []struct {
		name    string
		seaweed bool
		caddy   bool
	}{
		{"none", false, false},
		{"seaweed_only", true, false},
		{"caddy_only", false, true},
		{"both", true, true},
	}

	for _, combo := range combinations {
		t.Run(combo.name, func(t *testing.T) {
			cfg := Config{
				EnableSeaweedFS: combo.seaweed,
				EnableCaddy:     combo.caddy,
				Domain:          "test.example.com",
				JWTSecret:       "test_jwt_secret_32chars_long!!!!",
				DBPassword:      "test_db_password_16!",
				DBRootPassword:  "test_root_password!",
				RedisPassword:   "test_redis_pass_16!",
				S3AccessKey:     "TESTACCESSKEY12345678",
				S3SecretKey:     "testsecretkey1234567890123456789012345678",
			}

			tempDir := t.TempDir()
			err := RenderAll(cfg, tempDir)
			if err != nil {
				t.Fatalf("RenderAll failed for %s: %v", combo.name, err)
			}

			// Verify expected files exist
			expectedFiles := []string{"docker-compose.yml", ".env", "kkphp.conf"}
			if combo.caddy {
				expectedFiles = append(expectedFiles, "Caddyfile")
			}
			if combo.seaweed {
				expectedFiles = append(expectedFiles, "kkfiler.toml")
			}

			for _, file := range expectedFiles {
				path := filepath.Join(tempDir, file)
				if _, err := os.Stat(path); os.IsNotExist(err) {
					t.Errorf("expected file %s not found", file)
				}
			}
		})
	}
}

// TestValidateTOML validates kkfiler.toml syntax
func TestValidateTOML(t *testing.T) {
	cfg := Config{
		EnableSeaweedFS: true,
		DBPassword:      "test",
		DBRootPassword:  "test",
		RedisPassword:   "test",
		Domain:          "test.com",
	}

	rendered, err := RenderTemplateToString("kkfiler.toml", cfg)
	if err != nil {
		t.Fatalf("Failed to render kkfiler.toml: %v", err)
	}

	// Parse TOML to validate syntax
	var result map[string]interface{}
	_, err = toml.Decode(rendered, &result)
	if err != nil {
		t.Errorf("kkfiler.toml has invalid TOML syntax: %v", err)
	}

	// Verify required sections exist
	if _, ok := result["mysql"]; !ok {
		t.Error("kkfiler.toml missing [mysql] section")
	}
	if _, ok := result["leveldb2"]; !ok {
		t.Error("kkfiler.toml missing [leveldb2] section")
	}
}

// TestValidateYAML validates docker-compose.yml syntax
func TestValidateYAML(t *testing.T) {
	cfg := Config{
		EnableSeaweedFS: true,
		EnableCaddy:     true,
		DBPassword:      "test",
		DBRootPassword:  "test",
		RedisPassword:   "test",
		Domain:          "test.com",
	}

	rendered, err := RenderTemplateToString("docker-compose.yml", cfg)
	if err != nil {
		t.Fatalf("Failed to render docker-compose.yml: %v", err)
	}

	// Parse YAML to validate syntax
	var result map[string]interface{}
	err = yaml.Unmarshal([]byte(rendered), &result)
	if err != nil {
		t.Errorf("docker-compose.yml has invalid YAML syntax: %v", err)
	}

	// Verify required top-level keys (docker-compose v3.8 doesn't require 'version')
	requiredKeys := []string{"services", "networks", "volumes"}
	for _, key := range requiredKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("docker-compose.yml missing required key: %s", key)
		}
	}
}

// TestCaddyfileSyntax validates Caddyfile structure
func TestCaddyfileSyntax(t *testing.T) {
	cfg := Config{
		EnableCaddy: true,
		Domain:      "example.com",
	}

	rendered, err := RenderTemplateToString("Caddyfile", cfg)
	if err != nil {
		t.Fatalf("Failed to render Caddyfile: %v", err)
	}

	// Basic syntax check: braces matching
	openBraces := strings.Count(rendered, "{")
	closeBraces := strings.Count(rendered, "}")
	if openBraces != closeBraces {
		t.Errorf("Caddyfile has mismatched braces: %d open, %d close", openBraces, closeBraces)
	}

	// Check domain is present
	if !strings.Contains(rendered, cfg.Domain) {
		t.Error("Caddyfile does not contain domain")
	}

	// Check reverse_proxy directive exists
	if !strings.Contains(rendered, "reverse_proxy") {
		t.Error("Caddyfile missing reverse_proxy directive")
	}
}

// TestGoldenFiles compares rendered output against golden files
func TestGoldenFiles(t *testing.T) {
	cfg := Config{
		EnableSeaweedFS: true,
		EnableCaddy:     true,
		Domain:          "example.com",
		JWTSecret:       "test_jwt_secret_32chars_long!!!!",
		LicenseKey:      "LICENSE-TESTKEY12345678",
		ServerPublicKey: "test_public_key_encrypted",
		DBPassword:      "test_db_pass",
		DBRootPassword:  "test_db_root_pass",
		RedisPassword:   "test_redis_pass",
		S3AccessKey:     "TESTACCESSKEY12345678",
		S3SecretKey:     "testsecretkey1234567890123456789012345678",
	}

	goldenTests := []struct {
		template   string
		goldenFile string
	}{
		{"Caddyfile", "Caddyfile.golden"},
		{"kkfiler.toml", "kkfiler.toml.golden"},
		{"kkphp.conf", "kkphp.conf.golden"},
		{"docker-compose.yml", "docker-compose.yml.golden"},
		{"env", "env.golden"},
	}

	for _, tt := range goldenTests {
		t.Run(tt.template, func(t *testing.T) {
			rendered, err := RenderTemplateToString(tt.template, cfg)
			if err != nil {
				t.Fatalf("Failed to render %s: %v", tt.template, err)
			}

			goldenPath := filepath.Join("testdata", "golden", tt.goldenFile)
			golden, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("Failed to read golden file %s: %v", goldenPath, err)
			}

			if diff := cmp.Diff(string(golden), rendered); diff != "" {
				t.Errorf("%s mismatch (-want +got):\n%s", tt.template, diff)
			}
		})
	}
}

// TestValidateSecrets tests the ValidateSecrets function
func TestValidateSecrets(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			cfg: Config{
				JWTSecret:      "this_is_a_32_character_secret!!!", // 32 chars
				DBPassword:     "password_16chars", // 16 chars
				DBRootPassword: "password_16chars",
				RedisPassword:  "password_16chars",
			},
			wantErr: false,
		},
		{
			name: "jwt secret too short",
			cfg: Config{
				JWTSecret:      "short",
				DBPassword:     "password_16chars",
				DBRootPassword: "password_16chars",
				RedisPassword:  "password_16chars",
			},
			wantErr: true,
			errMsg:  "JWT_SECRET must be at least 32 characters",
		},
		{
			name: "db password too short",
			cfg: Config{
				JWTSecret:      "this_is_a_32_character_secret!!!",
				DBPassword:     "short",
				DBRootPassword: "password_16chars",
				RedisPassword:  "password_16chars",
			},
			wantErr: true,
			errMsg:  "DB_PASSWORD must be at least 16 characters",
		},
		{
			name: "s3 validation only when seaweedfs enabled",
			cfg: Config{
				EnableSeaweedFS: true,
				JWTSecret:       "this_is_a_32_character_secret!!!",
				DBPassword:      "password_16chars",
				DBRootPassword:  "password_16chars",
				RedisPassword:   "password_16chars",
				S3AccessKey:     "short", // too short
				S3SecretKey:     "short",
			},
			wantErr: true,
			errMsg:  "S3_ACCESS_KEY must be at least 16 characters",
		},
		{
			name: "s3 not validated when seaweedfs disabled",
			cfg: Config{
				EnableSeaweedFS: false,
				JWTSecret:       "this_is_a_32_character_secret!!!",
				DBPassword:      "password_16chars",
				DBRootPassword:  "password_16chars",
				RedisPassword:   "password_16chars",
				S3AccessKey:     "short", // short but ignored
				S3SecretKey:     "short",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.ValidateSecrets()
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errMsg, err.Error())
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

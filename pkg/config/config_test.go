package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigDir(t *testing.T) {
	dir := ConfigDir()
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".kk")
	assert.Equal(t, expected, dir)
}

func TestConfigPath(t *testing.T) {
	path := ConfigPath()
	assert.Contains(t, path, ".kk")
	assert.Contains(t, path, "config.yaml")
}

func TestLoad_DefaultsWhenNoFile(t *testing.T) {
	// Use temp dir to avoid affecting real config
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	defer func() {
		t.Setenv("HOME", origHome)
	}()

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "en", cfg.Language)
}

func TestSaveAndLoad(t *testing.T) {
	// Use temp dir
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	defer func() {
		t.Setenv("HOME", origHome)
	}()

	// Save config
	cfg := &Config{Language: "vi"}
	err := cfg.Save()
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(ConfigPath())
	require.NoError(t, err)

	// Load and verify
	loaded, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "vi", loaded.Language)
}

func TestLoad_InvalidLanguageDefaultsToEnglish(t *testing.T) {
	// Use temp dir
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	defer func() {
		t.Setenv("HOME", origHome)
	}()

	// Create config with invalid language
	configDir := filepath.Join(tmpDir, ".kk")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("language: invalid"), 0644)
	require.NoError(t, err)

	// Load should default to English
	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "en", cfg.Language)
}

func TestLoad_CorruptYAML(t *testing.T) {
	// Use temp dir
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	defer func() {
		t.Setenv("HOME", origHome)
	}()

	// Create corrupt config file
	configDir := filepath.Join(tmpDir, ".kk")
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("not: valid: yaml: here"), 0644)
	require.NoError(t, err)

	// Load should return error
	_, err = Load()
	assert.Error(t, err)
}

func TestEnsureProjectDir_NoProjectConfigured(t *testing.T) {
	// Use temp dir
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	defer func() {
		t.Setenv("HOME", origHome)
	}()

	// No config file = no project configured
	_, err := EnsureProjectDir()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no project configured")
}

func TestEnsureProjectDir_DirectoryNotExists(t *testing.T) {
	// Use temp dir
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	defer func() {
		t.Setenv("HOME", origHome)
	}()

	// Save config with non-existent directory
	cfg := &Config{
		Language:   "en",
		ProjectDir: "/nonexistent/directory/path",
	}
	err := cfg.Save()
	require.NoError(t, err)

	// EnsureProjectDir should fail
	_, err = EnsureProjectDir()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "project directory no longer exists")

	// Config should be cleared
	loaded, _ := Load()
	assert.Empty(t, loaded.ProjectDir)
}

func TestEnsureProjectDir_NoDockerCompose(t *testing.T) {
	// Use temp dir
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	defer func() {
		t.Setenv("HOME", origHome)
	}()

	// Create project directory without docker-compose.yml
	projectDir := filepath.Join(tmpDir, "myproject")
	err := os.MkdirAll(projectDir, 0755)
	require.NoError(t, err)

	// Save config
	cfg := &Config{
		Language:   "en",
		ProjectDir: projectDir,
	}
	err = cfg.Save()
	require.NoError(t, err)

	// EnsureProjectDir should fail
	_, err = EnsureProjectDir()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "docker-compose.yml not found")

	// Config should be cleared
	loaded, _ := Load()
	assert.Empty(t, loaded.ProjectDir)
}

func TestEnsureProjectDir_Success(t *testing.T) {
	// Use temp dir
	origHome := os.Getenv("HOME")
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	defer func() {
		t.Setenv("HOME", origHome)
	}()

	// Create valid project directory with docker-compose.yml
	projectDir := filepath.Join(tmpDir, "myproject")
	err := os.MkdirAll(projectDir, 0755)
	require.NoError(t, err)

	composePath := filepath.Join(projectDir, "docker-compose.yml")
	err = os.WriteFile(composePath, []byte("version: '3'\n"), 0644)
	require.NoError(t, err)

	// Save config
	cfg := &Config{
		Language:   "en",
		ProjectDir: projectDir,
	}
	err = cfg.Save()
	require.NoError(t, err)

	// EnsureProjectDir should succeed
	resultDir, err := EnsureProjectDir()
	assert.NoError(t, err)
	assert.Equal(t, projectDir, resultDir)

	// Verify we changed to the correct directory
	cwd, _ := os.Getwd()
	assert.Equal(t, projectDir, cwd)
}

func TestIsProjectNotConfiguredError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"no project", errors.New("no project configured, run 'kk init' first"), true},
		{"dir not exists", errors.New("project directory no longer exists: /foo"), true},
		{"no compose", errors.New("docker-compose.yml not found in: /foo"), true},
		{"other error", errors.New("some other error"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsProjectNotConfiguredError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestReadEnvValue(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test .env file
	envContent := `# Comment line
SYSTEM_DOMAIN=example.com
DB_PASSWORD=secret123
EMPTY_VALUE=
SPACED_KEY = spaced_value
`
	err := os.WriteFile(filepath.Join(tmpDir, ".env"), []byte(envContent), 0644)
	require.NoError(t, err)

	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{"existing key", "SYSTEM_DOMAIN", "example.com"},
		{"another key", "DB_PASSWORD", "secret123"},
		{"empty value", "EMPTY_VALUE", ""},
		{"spaced key", "SPACED_KEY", "spaced_value"},
		{"non-existent key", "NOT_EXISTS", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReadEnvValue(tmpDir, tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}

	// Test non-existent directory
	t.Run("non-existent dir", func(t *testing.T) {
		result := ReadEnvValue("/nonexistent/path", "SYSTEM_DOMAIN")
		assert.Equal(t, "", result)
	})
}

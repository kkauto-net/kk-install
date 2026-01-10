package config

import (
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

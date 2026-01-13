package n8n

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestN8nDir(t *testing.T) {
	dir := N8nDir()
	if dir == "" {
		t.Error("N8nDir() returned empty string")
	}
	// Should end with /n8n
	if !strings.HasSuffix(dir, "/n8n") {
		t.Errorf("N8nDir() = %q, expected to end with /n8n", dir)
	}
}

func TestDataDir(t *testing.T) {
	dir := DataDir()
	if !strings.HasSuffix(dir, "/n8n/data") {
		t.Errorf("DataDir() = %q, expected to end with /n8n/data", dir)
	}
}

func TestPostgresDir(t *testing.T) {
	dir := PostgresDir()
	if !strings.HasSuffix(dir, "/n8n/postgres") {
		t.Errorf("PostgresDir() = %q, expected to end with /n8n/postgres", dir)
	}
}

func TestComposePath(t *testing.T) {
	path := ComposePath()
	if !strings.HasSuffix(path, "/n8n/docker-compose.yml") {
		t.Errorf("ComposePath() = %q, expected to end with /n8n/docker-compose.yml", path)
	}
}

func TestEnvPath(t *testing.T) {
	path := EnvPath()
	if !strings.HasSuffix(path, "/n8n/.env") {
		t.Errorf("EnvPath() = %q, expected to end with /n8n/.env", path)
	}
}

func TestIsInstalled(t *testing.T) {
	// Should return false for non-existent directory
	if IsInstalled() {
		// This is expected if n8n is actually installed
		t.Log("n8n appears to be installed (or test is running in installed env)")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     N8nConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			cfg: N8nConfig{
				N8nHost:       "n8n.example.com",
				DBUser:        "n8n",
				DBPassword:    "password12345678",
				EncryptionKey: "encryption-key-12345678901234567890",
			},
			wantErr: false,
		},
		{
			name: "encryption key too short",
			cfg: N8nConfig{
				N8nHost:       "n8n.example.com",
				DBUser:        "n8n",
				DBPassword:    "password12345678",
				EncryptionKey: "short",
			},
			wantErr: true,
			errMsg:  "encryption key must be at least 32 characters",
		},
		{
			name: "db password too short",
			cfg: N8nConfig{
				N8nHost:       "n8n.example.com",
				DBUser:        "n8n",
				DBPassword:    "short",
				EncryptionKey: "encryption-key-12345678901234567890",
			},
			wantErr: true,
			errMsg:  "database password must be at least 16 characters",
		},
		{
			name: "missing db user",
			cfg: N8nConfig{
				N8nHost:       "n8n.example.com",
				DBUser:        "",
				DBPassword:    "password12345678",
				EncryptionKey: "encryption-key-12345678901234567890",
			},
			wantErr: true,
			errMsg:  "database user is required",
		},
		{
			name: "missing n8n host",
			cfg: N8nConfig{
				N8nHost:       "",
				DBUser:        "n8n",
				DBPassword:    "password12345678",
				EncryptionKey: "encryption-key-12345678901234567890",
			},
			wantErr: true,
			errMsg:  "n8n host is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Validate() error = %v, want error containing %q", err, tt.errMsg)
			}
		})
	}
}

func TestEnsureDirectories(t *testing.T) {
	// Create temp directory for testing
	tmpDir := t.TempDir()

	// Override N8nDir for this test by setting env
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	err := EnsureDirectories()
	if err != nil {
		t.Fatalf("EnsureDirectories() error = %v", err)
	}

	// Verify directories exist
	dirs := []string{
		filepath.Join(tmpDir, ".kk", "n8n"),
		filepath.Join(tmpDir, ".kk", "n8n", "data"),
		filepath.Join(tmpDir, ".kk", "n8n", "postgres"),
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("EnsureDirectories() did not create %q", dir)
		}
	}
}

func TestRenderAll(t *testing.T) {
	// Create temp directory for testing
	tmpDir := t.TempDir()

	// Override HOME for N8nDir fallback
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cfg := N8nConfig{
		Domain:          "n8n.example.com",
		N8nHost:         "n8n.example.com",
		DBUser:          "n8n",
		DBPassword:      "testpassword12345678",
		EncryptionKey:   "test-encryption-key-12345678901234",
		Timezone:        "Asia/Ho_Chi_Minh",
		ConnectKKEngine: false,
	}

	err := RenderAll(cfg)
	if err != nil {
		t.Fatalf("RenderAll() error = %v", err)
	}

	// Verify files exist
	expectedFiles := []string{
		filepath.Join(tmpDir, ".kk", "n8n", "docker-compose.yml"),
		filepath.Join(tmpDir, ".kk", "n8n", ".env"),
	}

	for _, path := range expectedFiles {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("RenderAll() did not create %q", path)
		}
	}

	// Verify .env permissions (0600)
	envPath := filepath.Join(tmpDir, ".kk", "n8n", ".env")
	info, err := os.Stat(envPath)
	if err != nil {
		t.Fatalf("Failed to stat .env: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf(".env permissions = %o, want 0600", perm)
	}

	// Verify .env content
	envContent, err := os.ReadFile(envPath)
	if err != nil {
		t.Fatalf("Failed to read .env: %v", err)
	}

	expectedStrings := []string{
		"N8N_HOST=n8n.example.com",
		"N8N_DB_USER=n8n",
		"N8N_DB_PASSWORD=testpassword12345678",
		"N8N_ENCRYPTION_KEY=test-encryption-key-12345678901234",
		"TZ=Asia/Ho_Chi_Minh",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(string(envContent), expected) {
			t.Errorf(".env content missing %q", expected)
		}
	}
}

func TestRenderAllWithKKEngine(t *testing.T) {
	// Create temp directory for testing
	tmpDir := t.TempDir()

	// Override HOME for N8nDir fallback
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	cfg := N8nConfig{
		Domain:          "n8n.example.com",
		N8nHost:         "n8n.example.com",
		DBUser:          "n8n",
		DBPassword:      "testpassword12345678",
		EncryptionKey:   "test-encryption-key-12345678901234",
		Timezone:        "Asia/Ho_Chi_Minh",
		ConnectKKEngine: true,
	}

	err := RenderAll(cfg)
	if err != nil {
		t.Fatalf("RenderAll() error = %v", err)
	}

	// Verify docker-compose.yml contains kkengine_net
	composePath := filepath.Join(tmpDir, ".kk", "n8n", "docker-compose.yml")
	composeContent, err := os.ReadFile(composePath)
	if err != nil {
		t.Fatalf("Failed to read docker-compose.yml: %v", err)
	}

	if !strings.Contains(string(composeContent), "kkengine_net") {
		t.Error("docker-compose.yml should contain kkengine_net when ConnectKKEngine=true")
	}

	if !strings.Contains(string(composeContent), "external: true") {
		t.Error("docker-compose.yml should contain external: true for kkengine_net")
	}
}

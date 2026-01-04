package compose

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseComposeFile(t *testing.T) {
	t.Run("valid docker-compose.yml", func(t *testing.T) {
		// Create a temporary docker-compose.yml file
		tempDir := t.TempDir()
		composeContent := `
version: '3.8'
services:
  web:
    image: nginx:latest
    ports:
      - "80:80"
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost"]
      interval: 30s
      timeout: 10s
      retries: 5
  db:
    image: postgres:latest
    ports:
      - "5432:5432"
`
		composePath := filepath.Join(tempDir, "docker-compose.yml")
		err := os.WriteFile(composePath, []byte(composeContent), 0644)
		assert.NoError(t, err)

		composeFile, err := ParseComposeFile(tempDir)
		assert.NoError(t, err)
		assert.NotNil(t, composeFile)
		assert.Len(t, composeFile.Services, 2)

		webService, ok := composeFile.Services["web"]
		assert.True(t, ok)
		assert.Equal(t, "nginx:latest", webService.Image)
		assert.Contains(t, webService.Ports, "80:80")
		assert.NotNil(t, webService.HealthCheck)
		assert.Contains(t, webService.HealthCheck.Test, "CMD")
		assert.Equal(t, "30s", webService.HealthCheck.Interval)

		dbService, ok := composeFile.Services["db"]
		assert.True(t, ok)
		assert.Equal(t, "postgres:latest", dbService.Image)
		assert.Contains(t, dbService.Ports, "5432:5432")
		assert.Nil(t, dbService.HealthCheck)
	})

	t.Run("non-existent docker-compose.yml", func(t *testing.T) {
		tempDir := t.TempDir()
		_, err := ParseComposeFile(tempDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no such file or directory")
	})

	t.Run("invalid yaml content", func(t *testing.T) {
		tempDir := t.TempDir()
		composeContent := `
services:
  web:
  - image: invalid:yaml
`
		composePath := filepath.Join(tempDir, "docker-compose.yml")
		err := os.WriteFile(composePath, []byte(composeContent), 0644)
		assert.NoError(t, err)

		_, err = ParseComposeFile(tempDir)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot unmarshal")
	})
}

func TestComposeFile_GetServiceNames(t *testing.T) {
	composeFile := &ComposeFile{
		Services: map[string]Service{
			"web": {},
			"db":  {},
			"app": {},
		},
	}
	names := composeFile.GetServiceNames()
	assert.Len(t, names, 3)
	assert.Contains(t, names, "web")
	assert.Contains(t, names, "db")
	assert.Contains(t, names, "app")
}

func TestComposeFile_HasHealthCheck(t *testing.T) {
	composeFile := &ComposeFile{
		Services: map[string]Service{
			"web": {
				HealthCheck: &HealthCheck{},
			},
			"db": {},
		},
	}
	assert.True(t, composeFile.HasHealthCheck("web"))
	assert.False(t, composeFile.HasHealthCheck("db"))
	assert.False(t, composeFile.HasHealthCheck("nonexistent"))
}

func TestComposeFile_GetServicePorts(t *testing.T) {
	composeFile := &ComposeFile{
		Services: map[string]Service{
			"web": {
				Ports: []string{"80:80", "443:443"},
			},
			"db": {},
		},
	}
	ports := composeFile.GetServicePorts("web")
	assert.Len(t, ports, 2)
	assert.Contains(t, ports, "80:80")
	assert.Contains(t, ports, "443:443")

	ports = composeFile.GetServicePorts("db")
	assert.Empty(t, ports)

	ports = composeFile.GetServicePorts("nonexistent")
	assert.Empty(t, ports)
}

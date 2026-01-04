package compose

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ComposeFile struct {
	Services map[string]Service `yaml:"services"`
}

type Service struct {
	Image       string       `yaml:"image"`
	Ports       []string     `yaml:"ports"`
	HealthCheck *HealthCheck `yaml:"healthcheck"`
	DependsOn   interface{}  `yaml:"depends_on"`
}

type HealthCheck struct {
	Test     []string `yaml:"test"`
	Interval string   `yaml:"interval"`
	Timeout  string   `yaml:"timeout"`
	Retries  int      `yaml:"retries"`
}

// ParseComposeFile reads and parses docker-compose.yml
func ParseComposeFile(dir string) (*ComposeFile, error) {
	composePath := filepath.Join(dir, "docker-compose.yml")
	content, err := os.ReadFile(composePath)
	if err != nil {
		return nil, err
	}

	var compose ComposeFile
	if err := yaml.Unmarshal(content, &compose); err != nil {
		return nil, err
	}

	return &compose, nil
}

// GetServiceNames returns list of service names
func (c *ComposeFile) GetServiceNames() []string {
	var names []string
	for name := range c.Services {
		names = append(names, name)
	}
	return names
}

// HasHealthCheck returns true if service has healthcheck defined
func (c *ComposeFile) HasHealthCheck(serviceName string) bool {
	if svc, ok := c.Services[serviceName]; ok {
		return svc.HealthCheck != nil
	}
	return false
}

// GetServicePorts extracts exposed ports for a service
func (c *ComposeFile) GetServicePorts(serviceName string) []string {
	if svc, ok := c.Services[serviceName]; ok {
		return svc.Ports
	}
	return nil
}

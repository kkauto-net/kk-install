package compose

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

type ComposeFile struct {
	Services map[string]Service `yaml:"services"`
}

type Service struct {
	Image         string       `yaml:"image"`
	ContainerName string       `yaml:"container_name"`
	Ports         []string     `yaml:"ports"`
	HealthCheck   *HealthCheck `yaml:"healthcheck"`
	DependsOn     interface{}  `yaml:"depends_on"`
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
	sort.Strings(names)
	return names
}

// GetServiceImages returns unique non-empty service images in deterministic service-name order.
func (c *ComposeFile) GetServiceImages() []string {
	seen := make(map[string]bool)
	var images []string

	for _, name := range c.GetServiceNames() {
		image := c.Services[name].Image
		if image == "" || seen[image] {
			continue
		}
		seen[image] = true
		images = append(images, image)
	}

	return images
}

// GetServiceContainerName returns configured container_name, or the generated kkengine default.
func (c *ComposeFile) GetServiceContainerName(serviceName string) string {
	if svc, ok := c.Services[serviceName]; ok && svc.ContainerName != "" {
		return svc.ContainerName
	}
	return fmt.Sprintf("kkengine_%s", serviceName)
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

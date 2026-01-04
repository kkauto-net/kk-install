package monitor

import (
	"context"
	"encoding/json"
	"strings"
)

type ServiceStatus struct {
	Name    string
	Status  string
	Health  string
	Ports   string
	Running bool
}

// ComposeExecutor interface for testing
type ComposeExecutor interface {
	Ps(ctx context.Context) (string, error)
}

// GetStatus returns status of all services
func GetStatus(ctx context.Context, executor ComposeExecutor) ([]ServiceStatus, error) {
	output, err := executor.Ps(ctx)
	if err != nil {
		return nil, err
	}

	return parseComposePs(output)
}

// Docker compose ps --format json output structure
type composePsJSON struct {
	Name    string `json:"Name"`
	State   string `json:"State"`
	Health  string `json:"Health"`
	Ports   string `json:"Ports"`
	Service string `json:"Service"`
}

func parseComposePs(output string) ([]ServiceStatus, error) {
	var statuses []ServiceStatus

	// Each line is a JSON object
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		var ps composePsJSON
		if err := json.Unmarshal([]byte(line), &ps); err != nil {
			continue // Skip malformed lines
		}

		status := ServiceStatus{
			Name:    ps.Service,
			Status:  ps.State,
			Health:  ps.Health,
			Ports:   ps.Ports,
			Running: strings.ToLower(ps.State) == "running",
		}

		statuses = append(statuses, status)
	}

	return statuses, nil
}

// IsAllHealthy checks if all services are running/healthy
func IsAllHealthy(statuses []ServiceStatus) bool {
	for _, s := range statuses {
		if !s.Running {
			return false
		}
		// If health check exists, must be healthy
		if s.Health != "" && s.Health != "healthy" {
			return false
		}
	}
	return true
}

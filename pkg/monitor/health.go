package monitor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

const (
	MaxRetries     = 3
	InitialDelay   = 2 * time.Second
	MaxDelay       = 30 * time.Second
	CheckInterval  = 3 * time.Second
)

type HealthStatus struct {
	ServiceName string
	Container   string
	Status      string // healthy, unhealthy, starting, none
	Healthy     bool
	Message     string
}

// DockerClient interface for testing
type DockerClient interface {
	ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error)
	Close() error
}

// HealthMonitor checks container health status
type HealthMonitor struct {
	client DockerClient
}

func NewHealthMonitor() (*HealthMonitor, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("tao Docker client that bai: %w", err)
	}
	return &HealthMonitor{client: cli}, nil
}

func (m *HealthMonitor) Close() {
	m.client.Close()
}

// WaitForHealthy waits for container to become healthy with retry
func (m *HealthMonitor) WaitForHealthy(ctx context.Context, containerName string, hasHealthCheck bool) HealthStatus {
	status := HealthStatus{
		Container: containerName,
	}

	// Extract service name from container name (e.g., kkengine_db -> db)
	parts := strings.Split(containerName, "_")
	if len(parts) > 1 {
		status.ServiceName = parts[len(parts)-1]
	} else {
		status.ServiceName = containerName
	}

	// If no health check defined, just check if running
	if !hasHealthCheck {
		return m.checkRunning(ctx, containerName, status)
	}

	// Wait for health check with retries
	delay := InitialDelay
	for retry := 0; retry < MaxRetries; retry++ {
		result := m.checkHealth(ctx, containerName)
		if result.Healthy {
			return result
		}

		// Wait before retry
		select {
		case <-ctx.Done():
			status.Status = "timeout"
			status.Message = "Da het thoi gian cho"
			return status
		case <-time.After(delay):
			// Exponential backoff
			delay = min(delay*2, MaxDelay)
		}
	}

	// Final check after all retries
	return m.checkHealth(ctx, containerName)
}

func (m *HealthMonitor) checkHealth(ctx context.Context, containerName string) HealthStatus {
	status := HealthStatus{Container: containerName}

	info, err := m.client.ContainerInspect(ctx, containerName)
	if err != nil {
		status.Status = "error"
		status.Message = fmt.Sprintf("Khong kiem tra duoc: %v", err)
		return status
	}

	// Extract service name
	parts := strings.Split(containerName, "_")
	if len(parts) > 1 {
		status.ServiceName = parts[len(parts)-1]
	} else {
		status.ServiceName = containerName
	}

	// Check if health check exists
	if info.State.Health == nil {
		// No health check, just check running status
		if info.State.Running {
			status.Status = "running"
			status.Healthy = true
		} else {
			status.Status = "stopped"
			status.Message = fmt.Sprintf("Exit code: %d", info.State.ExitCode)
		}
		return status
	}

	// Check health status
	status.Status = info.State.Health.Status
	switch info.State.Health.Status {
	case "healthy":
		status.Healthy = true
	case "starting":
		status.Message = "Dang khoi dong..."
	case "unhealthy":
		// Get last health check log
		if len(info.State.Health.Log) > 0 {
			lastLog := info.State.Health.Log[len(info.State.Health.Log)-1]
			status.Message = lastLog.Output
		}
	}

	return status
}

func (m *HealthMonitor) checkRunning(ctx context.Context, containerName string, status HealthStatus) HealthStatus {
	info, err := m.client.ContainerInspect(ctx, containerName)
	if err != nil {
		status.Status = "error"
		status.Message = fmt.Sprintf("Khong kiem tra duoc: %v", err)
		return status
	}

	if info.State.Running {
		status.Status = "running"
		status.Healthy = true
	} else {
		status.Status = "stopped"
		status.Message = fmt.Sprintf("Exit code: %d", info.State.ExitCode)
	}

	return status
}

// MonitorAll waits for all containers to be healthy
func (m *HealthMonitor) MonitorAll(ctx context.Context, containers []ContainerInfo, onProgress func(HealthStatus)) []HealthStatus {
	var results []HealthStatus

	for _, c := range containers {
		// Report starting
		onProgress(HealthStatus{
			ServiceName: c.ServiceName,
			Container:   c.ContainerName,
			Status:      "starting",
			Message:     "Dang kiem tra...",
		})

		status := m.WaitForHealthy(ctx, c.ContainerName, c.HasHealthCheck)
		results = append(results, status)

		// Report result
		onProgress(status)
	}

	return results
}

type ContainerInfo struct {
	ServiceName    string
	ContainerName  string
	HasHealthCheck bool
}

func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

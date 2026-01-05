package monitor

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
)

// MockDockerClient implements DockerClient interface
type MockDockerClient struct {
	mockContainerInspect func(ctx context.Context, containerID string) (types.ContainerJSON, error)
	mockClose            func() error
}

func (m *MockDockerClient) ContainerInspect(ctx context.Context, containerID string) (types.ContainerJSON, error) {
	if m.mockContainerInspect != nil {
		return m.mockContainerInspect(ctx, containerID)
	}
	return types.ContainerJSON{}, errors.New("ContainerInspect not mocked")
}

func (m *MockDockerClient) Close() error {
	if m.mockClose != nil {
		return m.mockClose()
	}
	return nil
}

func TestNewHealthMonitor(t *testing.T) {
	// We can't easily mock NewHealthMonitor without changing the package,
	// so we just test that it doesn't panic and returns an error when Docker is not available
	// In a real environment with Docker, this would succeed
	monitor, err := NewHealthMonitor()
	// Either succeeds or fails gracefully
	if err != nil {
		assert.Contains(t, err.Error(), "client")
		assert.Nil(t, monitor)
	} else {
		assert.NotNil(t, monitor)
		assert.NotNil(t, monitor.client)
		monitor.Close()
	}
}

func TestHealthMonitor_WaitForHealthy_NoHealthCheck(t *testing.T) {
	mockClient := &MockDockerClient{}
	monitor := &HealthMonitor{client: mockClient}
	ctx := context.Background()

	// Running container
	mockClient.mockContainerInspect = func(ctx context.Context, containerID string) (types.ContainerJSON, error) {
		return types.ContainerJSON{
			ContainerJSONBase: &types.ContainerJSONBase{
				State: &types.ContainerState{Running: true, Status: "running"},
			},
		}, nil
	}
	status := monitor.WaitForHealthy(ctx, "kkengine_web", false)
	assert.True(t, status.Healthy)
	assert.Equal(t, "running", status.Status)
	assert.Equal(t, "web", status.ServiceName)

	// Stopped container
	mockClient.mockContainerInspect = func(ctx context.Context, containerID string) (types.ContainerJSON, error) {
		return types.ContainerJSON{
			ContainerJSONBase: &types.ContainerJSONBase{
				State: &types.ContainerState{Running: false, Status: "exited", ExitCode: 0},
			},
		}, nil
	}
	status = monitor.WaitForHealthy(ctx, "kkengine_db", false)
	assert.False(t, status.Healthy)
	assert.Equal(t, "stopped", status.Status)
	assert.Contains(t, status.Message, "Exit code: 0")

	// Error inspecting container
	mockClient.mockContainerInspect = func(ctx context.Context, containerID string) (types.ContainerJSON, error) {
		return types.ContainerJSON{}, errors.New("container inspect error")
	}
	status = monitor.WaitForHealthy(ctx, "kkengine_error_no_health", false)
	assert.False(t, status.Healthy)
	assert.Equal(t, "error", status.Status)
	assert.Contains(t, status.Message, "container inspect error")
}

func TestHealthMonitor_WaitForHealthy_WithHealthCheck(t *testing.T) {
	mockClient := &MockDockerClient{}
	monitor := &HealthMonitor{client: mockClient}
	ctx := context.Background()

	t.Run("becomes healthy eventually", func(t *testing.T) {
		callCount := 0
		mockClient.mockContainerInspect = func(ctx context.Context, containerID string) (types.ContainerJSON, error) {
			callCount++
			if callCount < 2 { // First call is 'starting'
				return types.ContainerJSON{
					ContainerJSONBase: &types.ContainerJSONBase{
						State: &types.ContainerState{
							Health: &container.Health{Status: "starting"},
						},
					},
				}, nil
			}
			// Second call onwards is 'healthy'
			return types.ContainerJSON{
				ContainerJSONBase: &types.ContainerJSONBase{
					State: &types.ContainerState{
						Health: &container.Health{Status: "healthy"},
					},
				},
			}, nil
		}
		status := monitor.WaitForHealthy(ctx, "kkengine_app", true)
		assert.True(t, status.Healthy)
		assert.Equal(t, "healthy", status.Status)
		assert.Equal(t, "app", status.ServiceName)
		assert.GreaterOrEqual(t, callCount, 2)
	})

	t.Run("remains unhealthy", func(t *testing.T) {
		mockClient.mockContainerInspect = func(ctx context.Context, containerID string) (types.ContainerJSON, error) {
			return types.ContainerJSON{
				ContainerJSONBase: &types.ContainerJSONBase{
					State: &types.ContainerState{
						Health: &container.Health{
							Status: "unhealthy",
							Log:    []*types.HealthcheckResult{{Output: "ping failed"}},
						},
					},
				},
			}, nil
		}
		status := monitor.WaitForHealthy(ctx, "kkengine_unhealthy", true)
		assert.False(t, status.Healthy)
		assert.Equal(t, "unhealthy", status.Status)
		assert.Contains(t, status.Message, "ping failed")
	})

	t.Run("context timeout", func(t *testing.T) {
		mockClient.mockContainerInspect = func(ctx context.Context, containerID string) (types.ContainerJSON, error) {
			return types.ContainerJSON{
				ContainerJSONBase: &types.ContainerJSONBase{
					State: &types.ContainerState{
						Health: &container.Health{Status: "starting"},
					},
				},
			}, nil
		}
		// Set a short timeout to ensure it triggers
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		status := monitor.WaitForHealthy(ctx, "kkengine_timeout", true)
		assert.False(t, status.Healthy)
		assert.Equal(t, "timeout", status.Status)
		assert.Contains(t, status.Message, "Da het thoi gian cho")
	})

	t.Run("inspect error during retry", func(t *testing.T) {
		callCount := 0
		mockClient.mockContainerInspect = func(ctx context.Context, containerID string) (types.ContainerJSON, error) {
			callCount++
			if callCount == 1 {
				return types.ContainerJSON{
					ContainerJSONBase: &types.ContainerJSONBase{
						State: &types.ContainerState{Health: &container.Health{Status: "starting"}},
					},
				}, nil
			}
			return types.ContainerJSON{}, errors.New("temporary inspect error")
		}
		status := monitor.WaitForHealthy(ctx, "kkengine_inspect_error", true)
		assert.False(t, status.Healthy)
		assert.Equal(t, "error", status.Status)
		assert.Contains(t, status.Message, "temporary inspect error")
	})
}

func TestHealthMonitor_MonitorAll(t *testing.T) {
	mockClient := &MockDockerClient{}
	monitor := &HealthMonitor{client: mockClient}
	ctx := context.Background()

	containers := []ContainerInfo{
		{ServiceName: "web", ContainerName: "kkengine_web", HasHealthCheck: true},
		{ServiceName: "db", ContainerName: "kkengine_db", HasHealthCheck: false},
		{ServiceName: "unhealthy_svc", ContainerName: "kkengine_unhealthy_svc", HasHealthCheck: true},
	}

	var mu sync.Mutex
	var receivedProgress []HealthStatus

	onProgress := func(status HealthStatus) {
		mu.Lock()
		defer mu.Unlock()
		receivedProgress = append(receivedProgress, status)
	}

	callCountWeb := 0
	mockClient.mockContainerInspect = func(ctx context.Context, containerID string) (types.ContainerJSON, error) {
		if containerID == "kkengine_web" {
			callCountWeb++
			if callCountWeb < 2 { // First call is 'starting'
				return types.ContainerJSON{
					ContainerJSONBase: &types.ContainerJSONBase{
						State: &types.ContainerState{Health: &container.Health{Status: "starting"}},
					},
				}, nil
			}
			return types.ContainerJSON{
				ContainerJSONBase: &types.ContainerJSONBase{
					State: &types.ContainerState{Health: &container.Health{Status: "healthy"}},
				},
			}, nil
		} else if containerID == "kkengine_db" {
			return types.ContainerJSON{
				ContainerJSONBase: &types.ContainerJSONBase{
					State: &types.ContainerState{Running: true, Status: "running"},
				},
			}, nil
		} else if containerID == "kkengine_unhealthy_svc" {
			return types.ContainerJSON{
				ContainerJSONBase: &types.ContainerJSONBase{
					State: &types.ContainerState{Health: &container.Health{Status: "unhealthy", Log: []*types.HealthcheckResult{{Output: "failed check"}}}},
				},
			}, nil
		}
		return types.ContainerJSON{}, errors.New("unexpected container ID in MonitorAll mock")
	}

	results := monitor.MonitorAll(ctx, containers, onProgress)

	assert.Len(t, results, 3)
	assert.True(t, results[0].Healthy)
	assert.Equal(t, "healthy", results[0].Status)
	assert.True(t, results[1].Healthy)
	assert.Equal(t, "running", results[1].Status)
	assert.False(t, results[2].Healthy)
	assert.Equal(t, "unhealthy", results[2].Status)

	// Check progress reports
	mu.Lock()
	defer mu.Unlock()
	assert.Len(t, receivedProgress, 6)
	assert.Contains(t, receivedProgress, HealthStatus{ServiceName: "web", Container: "kkengine_web", Status: "starting", Message: "Dang kiem tra..."})
	assert.Contains(t, receivedProgress, HealthStatus{ServiceName: "web", Container: "kkengine_web", Status: "healthy", Healthy: true})
	assert.Contains(t, receivedProgress, HealthStatus{ServiceName: "db", Container: "kkengine_db", Status: "starting", Message: "Dang kiem tra..."})
	assert.Contains(t, receivedProgress, HealthStatus{ServiceName: "db", Container: "kkengine_db", Status: "running", Healthy: true})
	assert.Contains(t, receivedProgress, HealthStatus{ServiceName: "unhealthy_svc", Container: "kkengine_unhealthy_svc", Status: "starting", Message: "Dang kiem tra..."})
	assert.Contains(t, receivedProgress, HealthStatus{ServiceName: "svc", Container: "kkengine_unhealthy_svc", Status: "unhealthy", Message: "failed check"})
}

func TestMin(t *testing.T) {
	assert.Equal(t, 1*time.Second, min(1*time.Second, 2*time.Second))
	assert.Equal(t, 1*time.Second, min(2*time.Second, 1*time.Second))
	assert.Equal(t, 1*time.Second, min(1*time.Second, 1*time.Second))
}

func TestHealthMonitor_Close(t *testing.T) {
	mockCloseCalled := false
	mockClient := &MockDockerClient{
		mockClose: func() error {
			mockCloseCalled = true
			return nil
		},
	}
	monitor := &HealthMonitor{client: mockClient}
	monitor.Close()
	assert.True(t, mockCloseCalled)
}

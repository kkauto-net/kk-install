package monitor

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockComposeExecutor mocks the compose.Executor for testing
type MockComposeExecutor struct {
	MockPs func(ctx context.Context) (string, error)
}

func (m *MockComposeExecutor) Ps(ctx context.Context) (string, error) {
	if m.MockPs != nil {
		return m.MockPs(ctx)
	}
	return "", errors.New("Ps not mocked")
}

// Implement other methods of compose.Executor if needed for other tests
func (m *MockComposeExecutor) Up(ctx context.Context) error             { return nil }
func (m *MockComposeExecutor) Down(ctx context.Context) error           { return nil }
func (m *MockComposeExecutor) Restart(ctx context.Context) error        { return nil }
func (m *MockComposeExecutor) Pull(ctx context.Context) (string, error) { return "", nil }
func (m *MockComposeExecutor) ForceRecreate(ctx context.Context) error  { return nil }


func TestGetStatus(t *testing.T) {
	t.Run("successful ps output", func(t *testing.T) {
		mockPsOutput := `
{"ID":"1a","Name":"test_web_1","Service":"web","Project":"test","State":"running","Health":"healthy","Ports":"0.0.0.0:80->80/tcp"}
{"ID":"2b","Name":"test_db_1","Service":"db","Project":"test","State":"running","Health":"","Ports":"5432/tcp"}
`
		mockExecutor := &MockComposeExecutor{
			MockPs: func(ctx context.Context) (string, error) {
				return mockPsOutput, nil
			},
		}

		statuses, err := GetStatus(context.Background(), mockExecutor)
		assert.NoError(t, err)
		assert.Len(t, statuses, 2)

		assert.Equal(t, "web", statuses[0].Name)
		assert.Equal(t, "running", statuses[0].Status)
		assert.Equal(t, "healthy", statuses[0].Health)
		assert.Equal(t, "0.0.0.0:80->80/tcp", statuses[0].Ports)
		assert.True(t, statuses[0].Running)

		assert.Equal(t, "db", statuses[1].Name)
		assert.Equal(t, "running", statuses[1].Status)
		assert.Equal(t, "", statuses[1].Health) // No healthcheck is considered healthy if running
		assert.Equal(t, "5432/tcp", statuses[1].Ports)
		assert.True(t, statuses[1].Running)
	})

	t.Run("executor ps returns error", func(t *testing.T) {
		mockExecutor := &MockComposeExecutor{
			MockPs: func(ctx context.Context) (string, error) {
				return "", errors.New("compose ps failed")
			},
		}

		statuses, err := GetStatus(context.Background(), mockExecutor)
		assert.Error(t, err)
		assert.Nil(t, statuses)
		assert.Contains(t, err.Error(), "compose ps failed")
	})

	t.Run("empty ps output", func(t *testing.T) {
		mockPsOutput := ""
		mockExecutor := &MockComposeExecutor{
			MockPs: func(ctx context.Context) (string, error) {
				return mockPsOutput, nil
			},
		}

		statuses, err := GetStatus(context.Background(), mockExecutor)
		assert.NoError(t, err)
		assert.Empty(t, statuses)
	})

	t.Run("malformed json line in ps output", func(t *testing.T) {
		mockPsOutput := `
{"ID":"1a","Name":"test_web_1","Service":"web","Project":"test","State":"running","Health":"healthy","Ports":"0.0.0.0:80->80/tcp"}
THIS IS NOT JSON
{"ID":"2b","Name":"test_db_1","Service":"db","Project":"test","State":"running","Health":"","Ports":"5432/tcp"}
`
		mockExecutor := &MockComposeExecutor{
			MockPs: func(ctx context.Context) (string, error) {
				return mockPsOutput, nil
			},
		}

		statuses, err := GetStatus(context.Background(), mockExecutor)
		assert.NoError(t, err)
		assert.Len(t, statuses, 2) // Malformed line should be skipped
		assert.Equal(t, "web", statuses[0].Name)
		assert.Equal(t, "db", statuses[1].Name)
	})
}

func TestIsAllHealthy(t *testing.T) {
	t.Run("all healthy", func(t *testing.T) {
		statuses := []ServiceStatus{
			{Name: "web", Status: "running", Health: "healthy", Running: true},
			{Name: "db", Status: "running", Health: "", Running: true}, // No healthcheck is considered healthy if running
		}
		assert.True(t, IsAllHealthy(statuses))
	})

	t.Run("one service not running", func(t *testing.T) {
		statuses := []ServiceStatus{
			{Name: "web", Status: "running", Health: "healthy", Running: true},
			{Name: "db", Status: "exited", Health: "", Running: false},
		}
		assert.False(t, IsAllHealthy(statuses))
	})

	t.Run("one service unhealthy", func(t *testing.T) {
		statuses := []ServiceStatus{
			{Name: "web", Status: "running", Health: "healthy", Running: true},
			{Name: "app", Status: "running", Health: "unhealthy", Running: true},
		}
		assert.False(t, IsAllHealthy(statuses))
	})

	t.Run("empty status list", func(t *testing.T) {
		statuses := []ServiceStatus{}
		assert.True(t, IsAllHealthy(statuses))
	})

	t.Run("service in starting state", func(t *testing.T) {
		statuses := []ServiceStatus{
			{Name: "web", Status: "running", Health: "healthy", Running: true},
			{Name: "app", Status: "running", Health: "starting", Running: true},
		}
		assert.False(t, IsAllHealthy(statuses))
	})
}

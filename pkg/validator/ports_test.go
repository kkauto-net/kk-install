package validator

import (
	"testing"
)

func TestCheckPort(t *testing.T) {
	// Test with a port that should be available (high port number)
	t.Run("Available port", func(t *testing.T) {
		status := CheckPort(54321)
		if status.InUse {
			t.Errorf("Expected port 54321 to be available, but it's in use")
		}
	})
}

func TestCheckAllPorts(t *testing.T) {
	t.Run("Check all ports without Caddy", func(t *testing.T) {
		results, _ := CheckAllPorts(false)
		if len(results) < 2 {
			t.Errorf("Expected at least 2 port checks, got %d", len(results))
		}
	})

	t.Run("Check all ports with Caddy", func(t *testing.T) {
		results, _ := CheckAllPorts(true)
		if len(results) < 4 {
			t.Errorf("Expected at least 4 port checks, got %d", len(results))
		}
	})
}

func TestFormatPortConflict(t *testing.T) {
	tests := []struct {
		name     string
		portName string
		status   PortStatus
		expected string
	}{
		{
			name:     "Port with PID and process",
			portName: "MariaDB",
			status:   PortStatus{Port: 3307, InUse: true, PID: 1234, Process: "mysqld"},
			expected: "  - Port 3307 (MariaDB): dang dung boi mysqld (PID 1234). Stop: sudo kill 1234",
		},
		{
			name:     "Port with PID only",
			portName: "kkengine",
			status:   PortStatus{Port: 8019, InUse: true, PID: 5678, Process: ""},
			expected: "  - Port 8019 (kkengine): dang dung boi PID 5678. Stop: sudo kill 5678",
		},
		{
			name:     "Port in use without PID",
			portName: "Caddy",
			status:   PortStatus{Port: 80, InUse: true, PID: 0, Process: ""},
			expected: "  - Port 80 (Caddy): dang duoc su dung. Kiem tra: sudo lsof -i :80",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatPortConflict(tt.portName, tt.status)
			if result != tt.expected {
				t.Errorf("formatPortConflict() = %q, want %q", result, tt.expected)
			}
		})
	}
}

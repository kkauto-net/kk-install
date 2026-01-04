package ui

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// CaptureStdout is a helper function to capture stdout
func CaptureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestSimpleSpinner_Lifecycle(t *testing.T) {
	message := "Loading something..."
	spinner := NewSpinner(message)

	// Redirect stdout to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	spinner.Start()

	// Give it some time to print a few frames
	time.Sleep(50 * time.Millisecond) // Shorten sleep for faster test
	spinner.UpdateMessage("Still loading...")
	time.Sleep(50 * time.Millisecond) // Shorten sleep for faster test
	spinner.Stop(true)

	w.Close()
	os.Stdout = oldStdout

	// Read all remaining output to prevent pipe deadlock, but don't assert its content
	// as it's highly variable due to \r and timing.
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Check if the final "OK" message is present
	assert.Contains(t, output, fmt.Sprintf("  [OK] %s", message))
}

func TestShowServiceProgress(t *testing.T) {
	tests := []struct {
		name        string
		serviceName string
		status      string
		expected    string
	}{
		{
			name:        "starting status",
			serviceName: "web",
			status:      "starting",
			expected:    "  [>] web khoi dong...\n",
		},
		{
			name:        "healthy status",
			serviceName: "db",
			status:      "healthy",
			expected:    "  [OK] db san sang\n",
		},
		{
			name:        "running status",
			serviceName: "app",
			status:      "running",
			expected:    "  [OK] app san sang\n",
		},
		{
			name:        "unhealthy status",
			serviceName: "cache",
			status:      "unhealthy",
			expected:    "  [X] cache khong khoe manh\n",
		},
		{
			name:        "unknown status",
			serviceName: "worker",
			status:      "pending",
			expected:    "  [?] worker: pending\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := CaptureStdout(func() {
				ShowServiceProgress(tt.serviceName, tt.status)
			})
			assert.Equal(t, tt.expected, output)
		})
	}
}

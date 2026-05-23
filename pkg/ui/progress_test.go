package ui

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// CaptureStdout is a helper function to capture stdout
func CaptureStdout(t *testing.T, f func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	f()

	require.NoError(t, w.Close())
	os.Stdout = old

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	return buf.String()
}

func TestSimpleSpinner_Lifecycle(t *testing.T) {
	message := "Loading something..."
	spinner := NewSpinner(message)

	// Redirect stdout to capture output
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	spinner.Start()

	// Give it some time to print a few frames
	time.Sleep(50 * time.Millisecond) // Shorten sleep for faster test
	updatedMessage := "Still loading..."
	spinner.UpdateMessage(updatedMessage)
	time.Sleep(50 * time.Millisecond) // Shorten sleep for faster test
	spinner.Stop(true)

	require.NoError(t, w.Close())
	os.Stdout = oldStdout

	// Read all remaining output to prevent pipe deadlock
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	output := buf.String()

	// Check if the final "OK" message with updated message is present
	assert.Contains(t, output, fmt.Sprintf("  [OK] %s", updatedMessage))
}

func TestShowServiceProgress(t *testing.T) {
	// Test that ShowServiceProgress handles all status types without panicking.
	// We cannot easily capture pterm output, so we verify behavior by ensuring
	// no panic occurs with various inputs.
	testCases := []struct {
		name        string
		serviceName string
		status      string
	}{
		{"starting", "web", "starting"},
		{"healthy", "db", "healthy"},
		{"running", "app", "running"},
		{"unhealthy", "cache", "unhealthy"},
		{"unknown", "worker", "pending"},
		{"empty status", "svc", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Should not panic
			assert.NotPanics(t, func() {
				ShowServiceProgress(tc.serviceName, tc.status)
			})
		})
	}
}

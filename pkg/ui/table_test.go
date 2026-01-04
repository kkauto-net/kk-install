package ui

import (
	"strings"
	"testing"

	"github.com/kkengine/kkcli/pkg/monitor"
	"github.com/stretchr/testify/assert"
)

func TestPrintStatusTable(t *testing.T) {
	tests := []struct {
		name     string
		statuses []monitor.ServiceStatus
		expected string
	}{
		{
			name: "basic services",
			statuses: []monitor.ServiceStatus{
				{Name: "web", Status: "running", Health: "healthy", Ports: "0.0.0.0:80->80/tcp", Running: true},
				{Name: "db", Status: "running", Health: "", Ports: "5432/tcp", Running: true},
			},
			expected: `
Trang thai dich vu:

 Service    Status     Health     Ports                     

 web        [OK] running  healthy    0.0.0.0:80->80/tcp  
 db         [OK] running  -          5432/tcp            


`,
		},
		{
			name: "service not running and unhealthy",
			statuses: []monitor.ServiceStatus{
				{Name: "api", Status: "exited", Health: "unhealthy", Ports: "8080/tcp", Running: false},
			},
			expected: `
Trang thai dich vu:

 Service    Status     Health     Ports                     

 api        [X] exited    unhealthy  8080/tcp            


`,
		},
		{
			name: "long service name and ports",
			statuses: []monitor.ServiceStatus{
				{Name: "verylongservicename", Status: "running", Health: "healthy", Ports: "0.0.0.0:8080->8080/tcp, 0.0.0.0:8443->8443/tcp", Running: true},
			},
			expected: `
Trang thai dich vu:

 Service              Status     Health     Ports                                              

 verylongservicename  [OK] running  healthy    0.0.0.0:8080->8080/tcp, 0.0.0.0:844... 


`,
		},
		{
			name:     "empty statuses",
			statuses: []monitor.ServiceStatus{},
			expected: `
Trang thai dich vu:

 Service    Status     Health     Ports                     



`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := CaptureStdout(func() {
				PrintStatusTable(tt.statuses)
			})
			// Clean up any leading/trailing whitespace from the actual output for comparison
			assert.Equal(t, strings.TrimSpace(tt.expected), strings.TrimSpace(output))
		})
	}
}

func TestPrintAccessInfo(t *testing.T) {
	tests := []struct {
		name     string
		statuses []monitor.ServiceStatus
		expected string
	}{
		{
			name: "standard services",
			statuses: []monitor.ServiceStatus{
				{Name: "kkengine", Status: "running", Ports: "8019/tcp", Running: true},
				{Name: "db", Status: "running", Ports: "3307/tcp", Running: true},
				{Name: "caddy", Status: "running", Ports: "80/tcp, 443/tcp", Running: true},
			},
			expected: `Truy cap:
  - kkengine: http://localhost:8019
  - MariaDB: localhost:3307
  - Web: http://localhost (HTTPS: https://localhost)

`,
		},
		{
			name: "service not running",
			statuses: []monitor.ServiceStatus{
				{Name: "kkengine", Status: "exited", Ports: "8019/tcp", Running: false},
			},
			expected: `Truy cap:

`,
		},
		{
			name: "service with no ports",
			statuses: []monitor.ServiceStatus{
				{Name: "kkengine", Status: "running", Ports: "", Running: true},
			},
			expected: `Truy cap:

`,
		},
		{
			name:     "empty statuses",
			statuses: []monitor.ServiceStatus{},
			expected: `Truy cap:

`,
		},
		{
			name: "other service",
			statuses: []monitor.ServiceStatus{
				{Name: "other-service", Status: "running", Ports: "9000/tcp", Running: true},
			},
			expected: `Truy cap:

`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := CaptureStdout(func() {
				PrintAccessInfo(tt.statuses)
			})
			assert.Equal(t, tt.expected, output)
		})
	}
}

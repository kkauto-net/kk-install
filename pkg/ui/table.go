package ui

import (
	"fmt"
	"strings"

	"github.com/kkauto-net/kk-install/pkg/monitor"
)

// PrintStatusTable displays service status as formatted table
func PrintStatusTable(statuses []monitor.ServiceStatus) {
	// Calculate column widths
	nameWidth := 10
	statusWidth := 10
	healthWidth := 10
	portsWidth := 25

	for _, s := range statuses {
		if len(s.Name) > nameWidth {
			nameWidth = len(s.Name)
		}
	}

	// Print header
	fmt.Println()
	fmt.Println("Trang thai dich vu:")
	fmt.Println(strings.Repeat("─", nameWidth+statusWidth+healthWidth+portsWidth+10))
	fmt.Printf("│ %-*s │ %-*s │ %-*s │ %-*s │\n",
		nameWidth, "Service",
		statusWidth, "Status",
		healthWidth, "Health",
		portsWidth, "Ports")
	fmt.Println(strings.Repeat("─", nameWidth+statusWidth+healthWidth+portsWidth+10))

	// Print rows
	for _, s := range statuses {
		health := s.Health
		if health == "" {
			health = "-"
		}

		ports := s.Ports
		if ports == "" {
			ports = "-"
		}
		// Truncate ports if too long
		if len(ports) > portsWidth {
			ports = ports[:portsWidth-3] + "..."
		}

		statusIcon := "[OK]"
		if !s.Running {
			statusIcon = "[X]"
		}

		fmt.Printf("│ %-*s │ %s %-*s │ %-*s │ %-*s │\n",
			nameWidth, s.Name,
			statusIcon, statusWidth-4, s.Status,
			healthWidth, health,
			portsWidth, ports)
	}

	fmt.Println(strings.Repeat("─", nameWidth+statusWidth+healthWidth+portsWidth+10))
	fmt.Println()
}

// PrintAccessInfo shows access URLs for services
func PrintAccessInfo(statuses []monitor.ServiceStatus) {
	fmt.Println("Truy cap:")
	for _, s := range statuses {
		if !s.Running || s.Ports == "" {
			continue
		}

		// Parse ports to show URLs
		switch s.Name {
		case "kkengine":
			fmt.Printf("  - kkengine: http://localhost:8019\n")
		case "db":
			fmt.Printf("  - MariaDB: localhost:3307\n")
		case "caddy":
			fmt.Printf("  - Web: http://localhost (HTTPS: https://localhost)\n")
		}
	}
	fmt.Println()
}

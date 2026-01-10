package ui

import (
	"github.com/pterm/pterm"

	"github.com/kkauto-net/kk-install/pkg/monitor"
)

// PrintStatusTable displays service status using pterm table
func PrintStatusTable(statuses []monitor.ServiceStatus) {
	pterm.DefaultSection.Println(Msg("service_status"))

	tableData := pterm.TableData{
		{Msg("col_service"), Msg("col_status"), Msg("col_health"), Msg("col_ports")},
	}

	for _, s := range statuses {
		statusText := pterm.Green("● " + Msg("status_running"))
		if !s.Running {
			statusText = pterm.Red("○ " + Msg("status_stopped"))
		}

		health := formatHealth(s.Health)
		ports := truncatePorts(s.Ports, 30)

		tableData = append(tableData, []string{
			s.Name,
			statusText,
			health,
			ports,
		})
	}

	pterm.DefaultTable.
		WithHasHeader(true).
		WithBoxed(true).
		WithData(tableData).
		Render()
}

func formatHealth(health string) string {
	if health == "" {
		return pterm.Gray("-")
	}
	if health == "healthy" {
		return pterm.Green("healthy")
	}
	if health == "unhealthy" {
		return pterm.Red("unhealthy")
	}
	return pterm.Yellow(health)
}

func truncatePorts(ports string, maxLen int) string {
	if ports == "" {
		return "-"
	}
	if len(ports) > maxLen {
		return ports[:maxLen-3] + "..."
	}
	return ports
}

// PrintAccessInfo shows access URLs for services
func PrintAccessInfo(statuses []monitor.ServiceStatus) {
	pterm.DefaultSection.Println(Msg("access_info"))

	tableData := pterm.TableData{
		{Msg("col_service"), Msg("col_url")},
	}

	for _, s := range statuses {
		if !s.Running {
			continue
		}
		url := getServiceURL(s.Name, s.Ports)
		if url != "" {
			tableData = append(tableData, []string{s.Name, url})
		}
	}

	if len(tableData) > 1 {
		pterm.DefaultTable.WithHasHeader(true).WithData(tableData).Render()
	}
}

func getServiceURL(name, _ string) string {
	switch name {
	case "kkengine":
		return "http://localhost:8019"
	case "db":
		return "localhost:3307"
	case "caddy":
		return "http://localhost (HTTPS: https://localhost)"
	default:
		return ""
	}
}

package ui

import (
	"github.com/pterm/pterm"

	"github.com/kkauto-net/kk-install/pkg/monitor"
)

// Table display constants
const (
	DigestTruncateLen = 12 // Length to truncate Docker image digests
	PortsTruncateLen  = 30 // Maximum length for ports display
)

// ImageUpdate represents an image update information for display.
type ImageUpdate struct {
	Image     string // Docker image name
	OldDigest string // Current image digest
	NewDigest string // New available digest
}

// PrintUpdatesTable displays available Docker image updates as a boxed table.
func PrintUpdatesTable(updates []ImageUpdate) {
	if len(updates) == 0 {
		return
	}

	tableData := pterm.TableData{
		{Msg("col_image"), Msg("col_current"), Msg("col_new")},
	}

	for _, u := range updates {
		old := truncateDigest(u.OldDigest, DigestTruncateLen)
		new := truncateDigest(u.NewDigest, DigestTruncateLen)
		tableData = append(tableData, []string{u.Image, old, new})
	}

	pterm.DefaultSection.Println(Msg("updates_available"))
	pterm.DefaultTable.
		WithHasHeader(true).
		WithBoxed(true).
		WithData(tableData).
		Render()
}

func truncateDigest(digest string, maxLen int) string {
	if len(digest) > maxLen {
		return digest[:maxLen] + "..."
	}
	return digest
}

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
		ports := truncatePorts(s.Ports, PortsTruncateLen)

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

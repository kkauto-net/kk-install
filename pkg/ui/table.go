package ui

import (
	"fmt"
	"strings"

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

// PrintStatusTable displays service status using pterm table with title
func PrintStatusTable(statuses []monitor.ServiceStatus) {
	tableData := pterm.TableData{
		{Msg("col_service"), Msg("col_status"), Msg("col_health"), Msg("col_ports")},
	}

	running := 0
	for _, s := range statuses {
		statusText := pterm.Green(IconRunning + " " + Msg("status_running"))
		if !s.Running {
			statusText = pterm.Red(IconStopped + " " + Msg("status_stopped"))
		} else {
			running++
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

	// Render table to string first, trim trailing newline
	tableStr, _ := pterm.DefaultTable.
		WithHasHeader(true).
		WithData(tableData).
		Srender()
	tableStr = strings.TrimSuffix(tableStr, "\n")

	// Print as boxed panel with title
	pterm.DefaultBox.
		WithTitle(pterm.Bold.Sprint("kk status")).
		WithTitleTopLeft().
		Print(tableStr)

	// Print summary box
	fmt.Println()
	var summaryMsg string
	var summaryColor pterm.Color
	if running == 0 {
		summaryMsg = MsgF("status_summary_stopped", len(statuses))
		summaryColor = pterm.FgYellow
	} else if running == len(statuses) {
		summaryMsg = MsgF("all_running", running)
		summaryColor = pterm.FgGreen
	} else {
		summaryMsg = MsgF("some_running", running, len(statuses))
		summaryColor = pterm.FgYellow
	}

	pterm.DefaultBox.
		WithTitle(pterm.Bold.Sprint(Msg("summary"))).
		WithTitleTopCenter().
		WithBoxStyle(pterm.NewStyle(summaryColor)).
		Println(summaryMsg)
}

func formatHealth(health string) string {
	switch health {
	case "":
		return pterm.Gray("-")
	case "healthy":
		return pterm.Green(IconHealthy + " healthy")
	case "unhealthy":
		return pterm.Red(IconUnhealthy + " unhealthy")
	case "starting":
		return pterm.Blue(IconStarting + " starting")
	default:
		return pterm.Yellow(IconWarning + " " + health)
	}
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

// PrintAccessInfo shows access URLs for services.
// domain: the configured SYSTEM_DOMAIN from .env (optional)
func PrintAccessInfo(statuses []monitor.ServiceStatus, domain string) {
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

	// Add domain URLs if domain is configured and not localhost
	if domain != "" && domain != "localhost" {
		tableData = append(tableData, []string{Msg("main_url"), "https://" + domain})
		tableData = append(tableData, []string{Msg("manager_system"), "https://" + domain + "/wtadmin/"})
	}

	if len(tableData) > 1 {
		fmt.Println() // Add spacing
		pterm.DefaultTable.WithHasHeader(true).WithBoxed(true).WithData(tableData).Render()
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

// PrintCommandResult displays service status table with command-specific title and summary.
// cmdName: command name for box title (e.g., "kk start")
// successMsgKey: i18n key for success message (e.g., "start_summary_success")
// partialMsgKey: i18n key for partial success (e.g., "start_summary_partial")
func PrintCommandResult(statuses []monitor.ServiceStatus, cmdName, successMsgKey, partialMsgKey string) {
	tableData := pterm.TableData{
		{Msg("col_service"), Msg("col_status"), Msg("col_health"), Msg("col_ports")},
	}

	running := 0
	for _, s := range statuses {
		statusText := pterm.Green(IconRunning + " " + Msg("status_running"))
		if !s.Running {
			statusText = pterm.Red(IconStopped + " " + Msg("status_stopped"))
		} else {
			running++
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

	// Render table to string first, trim trailing newline
	tableStr, _ := pterm.DefaultTable.
		WithHasHeader(true).
		WithData(tableData).
		Srender()
	tableStr = strings.TrimSuffix(tableStr, "\n")

	// Print as boxed panel with command title
	pterm.DefaultBox.
		WithTitle(pterm.Bold.Sprint(cmdName)).
		WithTitleTopLeft().
		Print(tableStr)

	// Print summary box
	fmt.Println()
	var summaryMsg string
	var summaryColor pterm.Color
	if running == len(statuses) && running > 0 {
		summaryMsg = MsgF(successMsgKey, running)
		summaryColor = pterm.FgGreen
	} else if running == 0 {
		// All services failed/stopped - use red
		summaryMsg = MsgF(partialMsgKey, running, len(statuses))
		summaryColor = pterm.FgRed
	} else {
		// Partial success - use yellow
		summaryMsg = MsgF(partialMsgKey, running, len(statuses))
		summaryColor = pterm.FgYellow
	}

	pterm.DefaultBox.
		WithTitle(pterm.Bold.Sprint(Msg("summary"))).
		WithTitleTopCenter().
		WithBoxStyle(pterm.NewStyle(summaryColor)).
		Println(summaryMsg)
}

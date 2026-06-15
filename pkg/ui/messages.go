package ui

import (
	"github.com/pterm/pterm"
)

// Icons for UI elements (Unicode emoji for compatibility)
const (
	IconLanguage = "🌐"  // Language selection
	IconDocker   = "🐳"  // Docker
	IconConfig   = "⚙️" // Config
	IconFolder   = "📁"  // Directory
	IconStorage  = "💾"  // SeaweedFS
	IconWeb      = "🌐"  // Caddy
	IconLink     = "🔗"  // Domain
	IconWrite    = "✍️" // Generating
	IconComplete = "✅"  // Complete
	IconCheck    = "✅"  // Success (same as complete)
	IconKey      = "🔑"  // License key
	IconClock    = "🕐"  // Timezone
)

// Status icons for service/health states
const (
	IconRunning   = "●" // Green - service running
	IconStopped   = "○" // Red - service stopped
	IconStarting  = "◐" // Blue - service starting
	IconHealthy   = "✓" // Green - health check passed
	IconUnhealthy = "✗" // Red - health check failed
	IconWarning   = "⚠" // Yellow - warning state
	IconUnknown   = "?" // Gray - unknown state
)

// Message functions using i18n
// These functions are kept for backward compatibility
func MsgCheckingDocker() string     { return Msg("checking_docker") }
func MsgDockerOK() string           { return Msg("docker_ok") }
func MsgCreated(file string) string { return MsgF("created", file) }
func MsgInitComplete() string       { return Msg("init_complete") }
func MsgDockerNotInstalled() string { return Msg("docker_not_installed") }
func MsgDockerNotRunning() string   { return Msg("docker_not_running") }
func MsgNextSteps() string          { return Msg("next_steps") }

// Progress indicators using pterm
func ShowSuccess(msg string) {
	pterm.Success.Println(msg)
}

func ShowError(msg string) {
	pterm.Error.Println(msg)
}

func ShowInfo(msg string) {
	pterm.Info.Println(msg)
}

func ShowWarning(msg string) {
	pterm.Warning.Println(msg)
}

// ShowWarningf prints a formatted warning using pterm.
func ShowWarningf(format string, args ...any) {
	pterm.Warning.Printfln(format, args...)
}

// ShowSuccessMsg prints a localized success message.
func ShowSuccessMsg(key string) {
	ShowSuccess(Msg(key))
}

// ShowNote prints an indented informational note.
func ShowNote(msg string) {
	pterm.Println("  " + msg)
}

// ShowOK prints a success line with consistent formatting.
func ShowOK(msg string) {
	pterm.Success.Println(msg)
}

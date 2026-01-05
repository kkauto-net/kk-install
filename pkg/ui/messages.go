package ui

import (
	"github.com/pterm/pterm"
)

// Icons for UI elements (Unicode emoji for compatibility)
const (
	IconLanguage = "🌐" // Language selection
	IconDocker   = "🐳" // Docker
	IconConfig   = "⚙️"  // Config
	IconFolder   = "📁" // Directory
	IconStorage  = "💾" // SeaweedFS
	IconWeb      = "🌐" // Caddy
	IconLink     = "🔗" // Domain
	IconWrite    = "✍️"  // Generating
	IconComplete = "✅" // Complete
	IconCheck    = "✅" // Success (same as complete)
)

// Message functions using i18n
// These functions are kept for backward compatibility
func MsgCheckingDocker() string         { return Msg("checking_docker") }
func MsgDockerOK() string               { return Msg("docker_ok") }
func MsgCreated(file string) string     { return MsgF("created", file) }
func MsgInitComplete() string           { return Msg("init_complete") }
func MsgDockerNotInstalled() string     { return Msg("docker_not_installed") }
func MsgDockerNotRunning() string       { return Msg("docker_not_running") }
func MsgNextSteps() string              { return Msg("next_steps") }

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

// Package ui provides color definitions following ClaudeKit CLI design guidelines.
package ui

import "github.com/pterm/pterm"

// Color palette hex codes (reference only - pterm uses terminal colors)
const (
	// Background
	HexBackground = "#1E1E28" // Dark charcoal/purple-grey

	// Text colors
	HexTextPrimary   = "#C5C6C7" // Light grey - main body text
	HexTextSecondary = "#ADAEAF" // Dimmed grey - descriptions
	HexTextHighlight = "#FFFFFF" // White - section headers

	// Accent colors
	HexCommand     = "#36C38A" // Teal green - commands, options
	HexPlaceholder = "#F5A97F" // Peach/orange - placeholders
	HexBrand       = "#66E0D2" // Cyan/aqua - logo, brand

	// Status colors
	HexSuccess = "#36C38A" // Green
	HexWarning = "#F5A97F" // Orange
	HexError   = "#FF6B6B" // Red
	HexInfo    = "#66E0D2" // Cyan
)

// pterm style presets for consistent styling across CLI
var (
	// Text styles
	StyleCommand     = pterm.NewStyle(pterm.FgLightGreen)              // Commands, options
	StylePlaceholder = pterm.NewStyle(pterm.FgLightYellow)             // <placeholders>
	StyleHeader      = pterm.NewStyle(pterm.FgWhite, pterm.Bold)       // Section headers
	StyleDescription = pterm.NewStyle(pterm.FgGray)                    // Body text
	StyleHint        = pterm.NewStyle(pterm.FgDarkGray)                // Hints, notes
	StyleBrand       = pterm.NewStyle(pterm.FgCyan)                    // Logo, brand accent
	StylePath        = pterm.NewStyle(pterm.FgCyan)                    // File paths
	StyleValue       = pterm.NewStyle(pterm.FgDefault)                 // Values, numbers
	StyleSuccess     = pterm.NewStyle(pterm.FgLightGreen)              // Success messages
	StyleWarning     = pterm.NewStyle(pterm.FgLightYellow)             // Warnings
	StyleError       = pterm.NewStyle(pterm.FgLightRed)                // Errors
	StyleInfo        = pterm.NewStyle(pterm.FgCyan)                    // Info messages
	StyleStepHeader  = pterm.NewStyle(pterm.FgWhite, pterm.BgDarkGray) // Step headers
)

// Color helper functions for inline styling
func Command(s string) string     { return StyleCommand.Sprint(s) }
func Placeholder(s string) string { return StylePlaceholder.Sprint(s) }
func Header(s string) string      { return StyleHeader.Sprint(s) }
func Description(s string) string { return StyleDescription.Sprint(s) }
func Hint(s string) string        { return StyleHint.Sprint(s) }
func Brand(s string) string       { return StyleBrand.Sprint(s) }
func Path(s string) string        { return StylePath.Sprint(s) }
func Value(s string) string       { return StyleValue.Sprint(s) }
func Success(s string) string     { return StyleSuccess.Sprint(s) }
func Warning(s string) string     { return StyleWarning.Sprint(s) }
func Error(s string) string       { return StyleError.Sprint(s) }
func Info(s string) string        { return StyleInfo.Sprint(s) }

// Logo returns the KKauto.net ASCII art logo in brand color (Cyan)
func Logo() string {
	logo := ` ██╗  ██╗██╗  ██╗ █████╗ ██╗   ██╗████████╗ ██████╗    ███╗   ██╗███████╗████████╗
 ██║ ██╔╝██║ ██╔╝██╔══██╗██║   ██║╚══██╔══╝██╔═══██╗   ████╗  ██║██╔════╝╚══██╔══╝
 █████╔╝ █████╔╝ ███████║██║   ██║   ██║   ██║   ██║   ██╔██╗ ██║█████╗     ██║
 ██╔═██╗ ██╔═██╗ ██╔══██║██║   ██║   ██║   ██║   ██║   ██║╚██╗██║██╔══╝     ██║
 ██║  ██╗██║  ██╗██║  ██║╚██████╔╝   ██║   ╚██████╔╝██╗██║ ╚████║███████╗   ██║
 ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝    ╚═╝    ╚═════╝ ╚═╝╚═╝  ╚═══╝╚══════╝   ╚═╝`
	return StyleBrand.Sprint(logo)
}

// LogoCompact returns a smaller single-line logo
func LogoCompact() string {
	return StyleBrand.Sprint("KKauto.net")
}

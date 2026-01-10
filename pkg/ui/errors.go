package ui

import "github.com/pterm/pterm"

// ErrorSuggestion contains error information and a suggested fix.
type ErrorSuggestion struct {
	Title      string // Error title displayed in box header
	Message    string // Error message body
	Suggestion string // How to fix the error
	Command    string // Optional command to run for fixing
}

// ShowBoxedError displays an error in a red box with optional fix suggestions.
// The error is displayed with a red border and icon for visibility.
func ShowBoxedError(err ErrorSuggestion) {
	content := err.Message
	if err.Suggestion != "" {
		content += "\n\n" + Msg("to_fix") + ":\n  " + err.Suggestion
	}
	if err.Command != "" {
		content += "\n\n" + Msg("then_run") + ": " + err.Command
	}

	pterm.DefaultBox.
		WithTitle(pterm.Red("❌ " + err.Title)).
		WithTitleTopLeft().
		WithBoxStyle(pterm.NewStyle(pterm.FgRed)).
		Println(content)
}

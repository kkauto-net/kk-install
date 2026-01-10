// Package ui provides user interface components for CLI output.
package ui

import "github.com/pterm/pterm"

// ShowCommandBanner displays a boxed header for a command.
// cmd is the command name (e.g., "kk status")
// description is a brief description shown inside the box.
func ShowCommandBanner(cmd, description string) {
	pterm.DefaultBox.
		WithTitle(pterm.Cyan(cmd)).
		WithTitleTopCenter().
		Println(description)
	pterm.Println() // spacing
}

// ShowCompletionBanner displays a boxed footer indicating success or failure.
// success determines the color (green for success, red for failure)
// title is shown as the box title
// content is the message displayed inside the box.
func ShowCompletionBanner(success bool, title, content string) {
	style := pterm.NewStyle(pterm.FgGreen)
	if !success {
		style = pterm.NewStyle(pterm.FgRed)
	}
	pterm.DefaultBox.
		WithTitle(title).
		WithTitleTopCenter().
		WithBoxStyle(style).
		Println(content)
}

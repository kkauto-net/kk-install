package ui

import (
	"fmt"
	"strings"

	"github.com/pterm/pterm"
)

// ErrorBoxMaxWidth is the maximum width for error box content
const ErrorBoxMaxWidth = 70

// ErrorSuggestion contains error information and a suggested fix.
type ErrorSuggestion struct {
	Title      string // Error title displayed in box header
	Message    string // Error message body
	Suggestion string // How to fix the error
	Command    string // Optional command to run for fixing
}

// wrapText wraps text to maxWidth characters per line
func wrapText(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return text
	}

	var result strings.Builder
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		if i > 0 {
			result.WriteString("\n")
		}

		// If line is shorter than max width, keep as is
		if len(line) <= maxWidth {
			result.WriteString(line)
			continue
		}

		// Wrap long lines
		words := strings.Fields(line)
		currentLine := ""
		for _, word := range words {
			if currentLine == "" {
				currentLine = word
			} else if len(currentLine)+1+len(word) <= maxWidth {
				currentLine += " " + word
			} else {
				result.WriteString(currentLine + "\n")
				currentLine = word
			}
		}
		if currentLine != "" {
			result.WriteString(currentLine)
		}
	}

	return result.String()
}

// ShowBoxedError displays an error in a red box with optional fix suggestions.
// The error is displayed with a red border and icon for visibility.
func ShowBoxedError(err ErrorSuggestion) {
	// Wrap the message to prevent overly wide boxes
	content := wrapText(err.Message, ErrorBoxMaxWidth)
	if err.Suggestion != "" {
		content += "\n\n" + Msg("to_fix") + ":\n  " + wrapText(err.Suggestion, ErrorBoxMaxWidth-2)
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

// ShowBoxedErrors displays multiple errors in a single red box.
// Useful for grouping related errors like preflight check failures.
func ShowBoxedErrors(title string, errors []ErrorSuggestion) {
	if len(errors) == 0 {
		return
	}

	var content strings.Builder
	for i, err := range errors {
		content.WriteString(fmt.Sprintf("%d. %s\n", i+1, err.Message))
		if err.Suggestion != "" {
			content.WriteString(fmt.Sprintf("   → %s\n", err.Suggestion))
		}
		if err.Command != "" {
			content.WriteString(fmt.Sprintf("   → %s: %s\n", Msg("then_run"), err.Command))
		}
		if i < len(errors)-1 {
			content.WriteString("\n")
		}
	}

	pterm.DefaultBox.
		WithTitle(pterm.Red("❌ " + title)).
		WithTitleTopLeft().
		WithBoxStyle(pterm.NewStyle(pterm.FgRed)).
		Println(content.String())
}

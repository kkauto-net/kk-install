package validator

import (
	"fmt"
	"strings"

	"github.com/kkauto-net/kk-install/pkg/ui"
)

// ErrorKey constants for translation
const (
	ErrDockerNotInstalled = "docker_not_installed"
	ErrDockerNotRunning   = "docker_not_running"
	ErrPortConflict       = "port_conflict"
	ErrEnvMissing         = "env_missing"
	ErrEnvMissingVars     = "env_missing_vars"
	ErrComposeMissing     = "compose_missing"
	ErrComposeSyntax      = "compose_syntax_error"
	ErrDiskLow            = "disk_low"
)

var errorSuggestionKeys = map[string]string{
	ErrDockerNotInstalled: "docker_install_suggestion",
	ErrDockerNotRunning:   "docker_start_suggestion",
	ErrPortConflict:       "port_conflict_suggestion",
	ErrEnvMissing:         "env_missing_suggestion",
	ErrEnvMissingVars:     "env_missing_vars_suggestion",
	ErrComposeMissing:     "env_missing_suggestion",
	ErrComposeSyntax:      "compose_syntax_error_suggestion",
	ErrDiskLow:            "preflight_fix_stop_conflicting",
}

// UserErrorMessage returns the localized message body for a UserError.
func UserErrorMessage(err error) string {
	if ue, ok := err.(*UserError); ok {
		if ue.Message != "" {
			return ue.Message
		}
		if ue.Key != "" {
			if len(ue.Args) > 0 {
				return ui.MsgF(ue.Key, ue.Args...)
			}
			return ui.Msg(ue.Key)
		}
	}
	if err != nil {
		return err.Error()
	}
	return ""
}

// UserErrorDetail returns optional technical detail attached to a UserError.
func UserErrorDetail(err error) string {
	if ue, ok := err.(*UserError); ok {
		return ue.Detail
	}
	return ""
}

// UserErrorSuggestion returns the localized suggestion for a UserError.
func UserErrorSuggestion(err error) string {
	if ue, ok := err.(*UserError); ok {
		if ue.Suggestion != "" {
			return ue.Suggestion
		}
		if ue.Key != "" {
			if suggestionKey, ok := errorSuggestionKeys[ue.Key]; ok {
				if ue.Key == ErrEnvMissingVars && len(ue.Args) > 0 {
					return ui.MsgF(suggestionKey, ue.Args[0])
				}
				if ue.Key == "compose_version_old" && len(ue.Args) > 0 {
					return ui.Msg("compose_version_old_suggestion")
				}
				return ui.Msg(suggestionKey)
			}
			suggestionKey := ue.Key + "_suggestion"
			if ui.Msg(suggestionKey) != suggestionKey {
				return ui.Msg(suggestionKey)
			}
		}
	}
	return ""
}

// FormatUserErrorForBox builds message text for boxed errors, including trimmed detail.
func FormatUserErrorForBox(err error) string {
	msg := UserErrorMessage(err)
	if detail := UserErrorDetail(err); detail != "" {
		trimmed := trimErrorDetail(detail)
		if trimmed != "" {
			return msg + "\n\n" + trimmed
		}
	}
	return msg
}

func trimErrorDetail(detail string) string {
	lines := strings.Split(strings.TrimSpace(detail), "\n")
	if len(lines) > 3 {
		lines = lines[len(lines)-3:]
	}
	result := strings.Join(lines, "\n")
	if len(result) > 240 {
		return result[:237] + "..."
	}
	return result
}

// IsDockerInstallError reports whether err is a Docker auto-install failure.
func IsDockerInstallError(err error) bool {
	key := UserErrorKey(err)
	if key == "docker_install_failed" {
		return true
	}
	return strings.HasPrefix(key, "docker_install_err_")
}

// TranslateError converts technical error to user-friendly localized text.
func TranslateError(err error) string {
	if _, ok := err.(*UserError); ok {
		msg := UserErrorMessage(err)
		suggestion := UserErrorSuggestion(err)
		if suggestion != "" {
			return fmt.Sprintf("%s\n  → %s", msg, suggestion)
		}
		return msg
	}
	return ui.MsgF("err_unknown", err)
}

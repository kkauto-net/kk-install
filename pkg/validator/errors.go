package validator

import (
	"fmt"

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

// TranslateError converts technical error to user-friendly localized text.
func TranslateError(err error) string {
	if ue, ok := err.(*UserError); ok {
		msg := ue.Message
		if msg == "" && ue.Key != "" {
			msg = ui.Msg(ue.Key)
		}
		suggestion := ue.Suggestion
		if suggestion == "" && ue.Key != "" {
			if suggestionKey, ok := errorSuggestionKeys[ue.Key]; ok {
				if ue.Key == ErrEnvMissingVars && len(ue.Args) > 0 {
					suggestion = ui.MsgF(suggestionKey, ue.Args[0])
				} else if ue.Key == "compose_version_old" && len(ue.Args) > 0 {
					suggestion = ui.Msg("compose_version_old_suggestion")
					msg = ui.MsgF(ue.Key, ue.Args[0])
				} else {
					suggestion = ui.Msg(suggestionKey)
				}
			} else if suggestionKey := ue.Key + "_suggestion"; ui.Msg(suggestionKey) != suggestionKey {
				suggestion = ui.Msg(suggestionKey)
			}
		}
		if suggestion != "" {
			return fmt.Sprintf("%s\n  → %s", msg, suggestion)
		}
		return msg
	}
	return ui.MsgF("err_unknown", err)
}

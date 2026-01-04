package validator

import "fmt"

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

// ErrorMessages maps error keys to Vietnamese messages
var ErrorMessages = map[string]struct {
	Message    string
	Suggestion string
}{
	ErrDockerNotInstalled: {
		Message:    "Docker chua cai dat",
		Suggestion: "Cai Docker tai: https://docs.docker.com/get-docker/",
	},
	ErrDockerNotRunning: {
		Message:    "Docker daemon khong chay",
		Suggestion: "Khoi dong Docker: sudo systemctl start docker",
	},
	ErrPortConflict: {
		Message:    "Co port dang bi su dung",
		Suggestion: "Xem chi tiet ben duoi",
	},
	ErrEnvMissing: {
		Message:    "File .env khong ton tai",
		Suggestion: "Chay: kk init",
	},
	ErrEnvMissingVars: {
		Message:    "Thieu bien moi truong bat buoc",
		Suggestion: "Xem chi tiet ben duoi",
	},
	ErrComposeMissing: {
		Message:    "File docker-compose.yml khong ton tai",
		Suggestion: "Chay: kk init",
	},
	ErrComposeSyntax: {
		Message:    "Loi cu phap trong docker-compose.yml",
		Suggestion: "Kiem tra YAML: indentation, colons, quotes",
	},
	ErrDiskLow: {
		Message:    "Disk space thap",
		Suggestion: "Don dep disk hoac mo rong storage",
	},
}

// TranslateError converts technical error to user-friendly
func TranslateError(err error) string {
	if ue, ok := err.(*UserError); ok {
		return fmt.Sprintf("%s\n  → %s", ue.Message, ue.Suggestion)
	}
	// Fallback for unknown errors
	return fmt.Sprintf("Loi: %v", err)
}

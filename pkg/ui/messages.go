package ui

import (
	"fmt"

	"github.com/pterm/pterm"
)

// Success messages
func MsgCheckingDocker() string { return "Dang kiem tra Docker..." }
func MsgDockerOK() string       { return "Docker da san sang" }
func MsgCreated(file string) string {
	return fmt.Sprintf("Da tao: %s", file)
}
func MsgInitComplete() string { return "Khoi tao hoan tat!" }

// Error messages
func MsgDockerNotInstalled() string { return "Docker chua cai dat" }
func MsgDockerNotRunning() string   { return "Docker daemon khong chay" }

// Next steps
func MsgNextSteps() string {
	return `
Buoc tiep theo:
  1. Kiem tra va chinh sua .env neu can
  2. Chay: kk start
`
}

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

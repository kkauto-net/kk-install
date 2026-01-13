package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/kkauto-net/kk-install/pkg/n8n"
	"github.com/kkauto-net/kk-install/pkg/ui"
)

var n8nLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View n8n logs",
	Long:  `Tail logs from n8n container. Use -f to follow in real-time.`,
	RunE:  runN8nLogs,
}

var (
	n8nLogsFollow bool
	n8nLogsTail   string
	n8nLogsAll    bool
)

func init() {
	n8nLogsCmd.Flags().BoolVarP(&n8nLogsFollow, "follow", "f", false,
		"Follow log output")
	n8nLogsCmd.Flags().StringVarP(&n8nLogsTail, "tail", "n", "100",
		"Number of lines to show")
	n8nLogsCmd.Flags().BoolVarP(&n8nLogsAll, "all", "a", false,
		"Show logs from all containers (n8n + postgres)")
	n8nCmd.AddCommand(n8nLogsCmd)
}

func runN8nLogs(cmd *cobra.Command, args []string) error {
	if !n8n.IsInstalled() {
		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:      "n8n Not Installed",
			Message:    ui.Msg("n8n_not_installed"),
			Suggestion: "Install n8n first",
			Command:    "kk n8n install",
		})
		return fmt.Errorf("n8n not installed")
	}

	// Build docker compose logs command
	cmdArgs := []string{"compose", "-f", n8n.ComposePath(), "logs"}

	if n8nLogsFollow {
		cmdArgs = append(cmdArgs, "-f")
	}

	cmdArgs = append(cmdArgs, "--tail", n8nLogsTail)

	// Service filter - only n8n by default
	if !n8nLogsAll {
		cmdArgs = append(cmdArgs, "n8n")
	}

	// Execute directly (stream to stdout)
	execCmd := exec.Command("docker", cmdArgs...)
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	execCmd.Stdin = os.Stdin

	return execCmd.Run()
}

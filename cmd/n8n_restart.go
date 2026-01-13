package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kkauto-net/kk-install/pkg/compose"
	"github.com/kkauto-net/kk-install/pkg/n8n"
	"github.com/kkauto-net/kk-install/pkg/ui"
)

var n8nRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart n8n services",
	Long:  `Restart n8n and PostgreSQL database containers.`,
	RunE:  runN8nRestart,
}

func init() {
	n8nCmd.AddCommand(n8nRestartCmd)
}

func runN8nRestart(cmd *cobra.Command, args []string) error {
	if !n8n.IsInstalled() {
		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:      "n8n Not Installed",
			Message:    ui.Msg("n8n_not_installed"),
			Suggestion: "Install n8n first",
			Command:    "kk n8n install",
		})
		return fmt.Errorf("n8n not installed")
	}

	ui.ShowStepHeader(1, 1, ui.Msg("restarting"))

	ctx, cancel := context.WithTimeout(context.Background(), compose.DefaultTimeout)
	defer cancel()

	executor := compose.NewExecutor(n8n.N8nDir())
	spinner := ui.StartPtermSpinner(ui.Msg("restarting"))

	if err := executor.Restart(ctx); err != nil {
		spinner.Fail(ui.Msg("restart_failed"))
		return err
	}
	spinner.Success(ui.Msg("restart_complete"))

	return nil
}

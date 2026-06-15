package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/kkauto-net/kk-install/pkg/compose"
	"github.com/kkauto-net/kk-install/pkg/monitor"
	"github.com/kkauto-net/kk-install/pkg/n8n"
	"github.com/kkauto-net/kk-install/pkg/ui"
)

var n8nStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show n8n service status",
	Long:  `Display status of n8n and PostgreSQL containers.`,
	RunE:  runN8nStatus,
}

func init() {
	n8nCmd.AddCommand(n8nStatusCmd)
}

func runN8nStatus(cmd *cobra.Command, args []string) error {
	if !n8n.IsInstalled() {
		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:      ui.Msg("err_title_n8n_not_installed"),
			Message:    ui.Msg("n8n_not_installed"),
			Suggestion: ui.Msg("n8n_run_install_hint"),
			Command:    "kk n8n install",
		})
		return fmt.Errorf("n8n not installed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	executor := compose.NewExecutor(n8n.N8nDir())

	definedServices := []string{"n8n", "n8n-postgres"}
	spinner := ui.StartPtermSpinner(ui.Msg("get_status_failed"))
	statuses, err := monitor.GetStatusWithServices(ctx, executor, definedServices)
	if err != nil {
		spinner.Fail(ui.Msg("get_status_failed"))
		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:      ui.Msg("get_status_failed"),
			Message:    ui.SanitizeError(err),
			Suggestion: ui.Msg("err_check_docker_running"),
			Command:    ui.Msg("docker_start_command"),
		})
		return err
	}
	spinner.Success(ui.Msg("status_desc"))

	ui.PrintStatusTable(statuses)

	for _, s := range statuses {
		if s.Name == "n8n" && s.Running {
			ui.ShowNote(ui.MsgF("n8n_access_url", n8n.AccessURL()))
			break
		}
	}

	return nil
}

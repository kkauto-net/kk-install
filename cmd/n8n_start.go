package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kkauto-net/kk-install/pkg/compose"
	"github.com/kkauto-net/kk-install/pkg/n8n"
	"github.com/kkauto-net/kk-install/pkg/ui"
	"github.com/kkauto-net/kk-install/pkg/validator"
)

var n8nStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start n8n services",
	Long:  `Start n8n and PostgreSQL database containers.`,
	RunE:  runN8nStart,
}

func init() {
	n8nCmd.AddCommand(n8nStartCmd)
}

func runN8nStart(cmd *cobra.Command, args []string) error {
	return runN8nStartInternal()
}

func runN8nStartInternal() error {
	if !n8n.IsInstalled() {
		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:      ui.Msg("err_title_n8n_not_installed"),
			Message:    ui.Msg("n8n_not_installed"),
			Suggestion: ui.Msg("err_install_n8n_first"),
			Command:    "kk n8n install",
		})
		return fmt.Errorf("n8n not installed")
	}

	ui.ShowStepHeader(1, 3, ui.Msg("step_preflight"))
	dv := validator.NewDockerValidator()
	if err := dv.CheckDockerDaemon(); err != nil {
		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:      ui.Msg("docker_not_running"),
			Message:    ui.SanitizeError(err),
			Suggestion: ui.Msg("docker_start_suggestion"),
			Command:    ui.Msg("docker_start_command"),
		})
		return err
	}

	portStatus := validator.CheckPort(5678)
	if portStatus.InUse && !portStatus.UsedByKKEngine {
		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:      ui.Msg("err_title_port_in_use"),
			Message:    ui.MsgF("err_port_in_use_msg", portStatus.Process, portStatus.PID),
			Suggestion: ui.Msg("err_port_in_use_suggestion"),
		})
		return fmt.Errorf("port 5678 in use")
	}
	ui.ShowSuccess(ui.Msg("docker_ok"))

	ui.ShowStepHeader(2, 3, ui.Msg("step_start_services"))
	ctx, cancel := context.WithTimeout(context.Background(), compose.DefaultTimeout)
	defer cancel()

	executor := compose.NewExecutor(n8n.N8nDir())
	spinner := ui.StartPtermSpinner(ui.Msg("n8n_starting"))

	if err := executor.Up(ctx); err != nil {
		spinner.Fail(ui.Msg("start_failed"))
		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:      ui.Msg("err_title_n8n_start_failed"),
			Message:    ui.SanitizeError(err),
			Suggestion: ui.Msg("err_check_docker_logs"),
			Command:    ui.Msg("docker_compose_logs_command"),
		})
		return err
	}
	spinner.Success(ui.Msg("n8n_started"))

	ui.ShowStepHeader(3, 3, ui.Msg("step_status"))
	accessURL := n8n.AccessURL()
	ui.ShowNote(ui.MsgF("n8n_running_at", accessURL))
	ui.ShowNote(ui.Msg("n8n_run_status_hint"))

	return nil
}

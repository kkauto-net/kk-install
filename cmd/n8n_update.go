package cmd

import (
	"context"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/kkauto-net/kk-install/pkg/compose"
	"github.com/kkauto-net/kk-install/pkg/n8n"
	"github.com/kkauto-net/kk-install/pkg/ui"
	"github.com/kkauto-net/kk-install/pkg/updater"
)

var n8nUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update n8n to latest version",
	Long:  `Pull latest n8n and PostgreSQL images and recreate containers.`,
	RunE:  runN8nUpdate,
}

var forceN8nUpdate bool

func init() {
	n8nUpdateCmd.Flags().BoolVarP(&forceN8nUpdate, "force", "f", false,
		"Skip confirmation")
	n8nCmd.AddCommand(n8nUpdateCmd)
}

func runN8nUpdate(cmd *cobra.Command, args []string) error {
	if !n8n.IsInstalled() {
		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:      ui.Msg("err_title_n8n_not_installed"),
			Message:    ui.Msg("n8n_not_installed"),
			Suggestion: ui.Msg("err_install_n8n_first"),
			Command:    "kk n8n install",
		})
		return fmt.Errorf("n8n not installed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), compose.DefaultTimeout)
	defer cancel()

	executor := compose.NewExecutor(n8n.N8nDir())

	ui.ShowStepHeader(1, 3, ui.Msg("step_pull_images"))
	spinner := ui.StartPtermSpinner(ui.Msg("pulling_images"))

	output, err := executor.Pull(ctx)
	if err != nil {
		spinner.Fail(ui.Msg("pull_failed"))
		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:      ui.Msg("pull_failed"),
			Message:    ui.SanitizeError(err),
			Suggestion: ui.Msg("err_check_docker_pull"),
		})
		return err
	}
	spinner.Success(ui.Msg("pulling_images"))

	updates := updater.ParsePullOutput(output)

	if len(updates) == 0 {
		ui.ShowOK(ui.Msg("images_up_to_date"))
		return nil
	}

	ui.ShowStepHeader(2, 3, ui.Msg("updates_available"))
	uiUpdates := make([]ui.ImageUpdate, len(updates))
	for i, u := range updates {
		uiUpdates[i] = ui.ImageUpdate{
			Image:     u.Image,
			OldDigest: u.OldDigest,
			NewDigest: u.NewDigest,
		}
	}
	ui.PrintUpdatesTable(uiUpdates)

	if !forceN8nUpdate {
		var confirm bool
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title(ui.Msg("confirm_restart")).
					Value(&confirm),
			),
		)
		if err := form.Run(); err != nil {
			return err
		}
		if !confirm {
			ui.ShowInfo(ui.Msg("update_cancelled"))
			return nil
		}
	}

	ui.ShowStepHeader(3, 3, ui.Msg("step_recreate"))
	spinner = ui.StartPtermSpinner(ui.Msg("recreating"))

	if err := executor.ForceRecreate(ctx); err != nil {
		spinner.Fail(ui.Msg("recreate_failed"))
		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:      ui.Msg("recreate_failed"),
			Message:    ui.SanitizeError(err),
			Suggestion: ui.Msg("err_check_docker_logs"),
			Command:    ui.Msg("docker_compose_logs_command"),
		})
		return err
	}
	spinner.Success(ui.Msg("update_complete"))

	ui.ShowSuccess(ui.Msg("n8n_update_success"))
	ui.ShowNote(ui.MsgF("n8n_access_url", n8n.AccessURL()))

	return nil
}

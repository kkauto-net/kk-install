package cmd

import (
	"context"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/kkauto-net/kk-install/pkg/compose"
	"github.com/kkauto-net/kk-install/pkg/n8n"
	"github.com/kkauto-net/kk-install/pkg/ui"
)

var n8nRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove n8n containers",
	Long:  `Remove n8n and PostgreSQL containers. Use -v to also remove data volumes.`,
	RunE:  runN8nRemove,
}

var n8nRemoveVolumes bool

func init() {
	n8nRemoveCmd.Flags().BoolVarP(&n8nRemoveVolumes, "volumes", "v", false,
		"Also remove data volumes (WARNING: deletes all data!)")
	n8nCmd.AddCommand(n8nRemoveCmd)
}

func runN8nRemove(cmd *cobra.Command, args []string) error {
	if !n8n.IsInstalled() {
		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:      ui.Msg("err_title_n8n_not_installed"),
			Message:    ui.Msg("n8n_not_installed"),
			Suggestion: ui.Msg("err_nothing_to_remove"),
		})
		return fmt.Errorf("n8n not installed")
	}

	if n8nRemoveVolumes {
		var confirm bool
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title(ui.Msg("n8n_remove_volume_warning")).
					Description(ui.Msg("n8n_remove_volume_warning_desc")).
					Value(&confirm),
			),
		)
		if err := form.Run(); err != nil {
			return err
		}
		if !confirm {
			ui.ShowInfo(ui.Msg("err_cancelled"))
			return nil
		}
	}

	ui.ShowStepHeader(1, 1, ui.Msg("step_remove_services"))

	ctx, cancel := context.WithTimeout(context.Background(), compose.DefaultTimeout)
	defer cancel()

	executor := compose.NewExecutor(n8n.N8nDir())
	spinner := ui.StartPtermSpinner(ui.Msg("removing_services"))

	var err error
	if n8nRemoveVolumes {
		err = executor.DownWithVolumes(ctx)
	} else {
		err = executor.Down(ctx)
	}

	if err != nil {
		spinner.Fail(ui.Msg("remove_failed"))
		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:      ui.Msg("remove_failed"),
			Message:    ui.SanitizeError(err),
			Suggestion: ui.Msg("err_check_docker_logs"),
			Command:    ui.Msg("docker_compose_logs_command"),
		})
		return err
	}

	if n8nRemoveVolumes {
		spinner.Success(ui.Msg("removed_with_volumes"))
	} else {
		spinner.Success(ui.Msg("removed_containers"))
		ui.ShowNote(ui.MsgF("n8n_data_preserved", n8n.N8nDir()))
	}

	return nil
}

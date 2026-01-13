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
			Title:      "n8n Not Installed",
			Message:    ui.Msg("n8n_not_installed"),
			Suggestion: "Nothing to remove",
		})
		return fmt.Errorf("n8n not installed")
	}

	// Warn about data deletion if volumes flag is set
	if n8nRemoveVolumes {
		var confirm bool
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("WARNING: This will delete all n8n data!").
					Description("Workflows, credentials, and database will be permanently lost.").
					Value(&confirm),
			),
		)
		if err := form.Run(); err != nil {
			return err
		}
		if !confirm {
			fmt.Println("Cancelled.")
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
		return err
	}

	if n8nRemoveVolumes {
		spinner.Success(ui.Msg("removed_with_volumes"))
	} else {
		spinner.Success(ui.Msg("removed_containers"))
		fmt.Println("  Data preserved in", n8n.N8nDir())
	}

	return nil
}

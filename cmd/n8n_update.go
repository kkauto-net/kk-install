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
			Title:      "n8n Not Installed",
			Message:    ui.Msg("n8n_not_installed"),
			Suggestion: "Install n8n first",
			Command:    "kk n8n install",
		})
		return fmt.Errorf("n8n not installed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), compose.DefaultTimeout)
	defer cancel()

	executor := compose.NewExecutor(n8n.N8nDir())

	// Step 1: Pull images
	ui.ShowStepHeader(1, 3, ui.Msg("step_pull_images"))
	spinner := ui.StartPtermSpinner(ui.Msg("pulling_images"))

	output, err := executor.Pull(ctx)
	if err != nil {
		spinner.Fail(ui.Msg("pull_failed"))
		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:   "Failed to pull images",
			Message: err.Error(),
		})
		return err
	}
	spinner.Success(ui.Msg("pulling_images"))

	// Step 2: Parse updates
	updates := updater.ParsePullOutput(output)

	if len(updates) == 0 {
		fmt.Println()
		ui.ShowSuccess(ui.Msg("images_up_to_date"))
		return nil
	}

	// Show updates
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
	fmt.Println()

	// Confirm
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
			fmt.Println(ui.Msg("update_cancelled"))
			return nil
		}
	}

	// Step 3: Recreate
	ui.ShowStepHeader(3, 3, ui.Msg("step_recreate"))
	spinner = ui.StartPtermSpinner(ui.Msg("recreating"))

	if err := executor.ForceRecreate(ctx); err != nil {
		spinner.Fail(ui.Msg("recreate_failed"))
		return err
	}
	spinner.Success(ui.Msg("update_complete"))

	fmt.Println()
	fmt.Println("  n8n updated successfully!")
	fmt.Println("  Access at: http://localhost:5678")

	return nil
}

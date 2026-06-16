package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/kkauto-net/kk-install/pkg/compose"
	"github.com/kkauto-net/kk-install/pkg/config"
	"github.com/kkauto-net/kk-install/pkg/monitor"
	"github.com/kkauto-net/kk-install/pkg/ui"
)

var updateCmd = &cobra.Command{
	Use:         "update",
	Short:       "Pull latest images and restart services",
	Long:        `Pull new images when available and optionally restart services to run the new version.`,
	Annotations: map[string]string{"group": "management"},
	RunE:        runUpdate,
}

var forceUpdate bool

func init() {
	updateCmd.Flags().BoolVarP(&forceUpdate, "force", "f", false, "Skip confirmation prompts")
	rootCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	cwd, err := config.EnsureProjectDir()
	if err != nil {
		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:      ui.Msg("project_not_configured"),
			Message:    ui.SanitizeError(err),
			Suggestion: ui.Msg("run_init_to_configure"),
			Command:    "kk init",
		})
		return err
	}

	ui.ShowCommandBanner(ui.Msg("cmd_update_title"), ui.Msg("update_desc"))

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n\n" + ui.Msg("stopping"))
		cancel()
	}()

	executor := compose.NewExecutor(cwd)
	imageState, err := prepareUpdateImageState(ctx, cwd)
	if err != nil {
		showUpdatePreparationError(err)
		return err
	}

	// Step 1: Pull new images
	ui.ShowStepHeader(1, 4, ui.Msg("step_pull_images"))
	spinner := ui.StartPtermSpinner(ui.Msg("pulling_images"))

	pullCtx, pullCancel := context.WithTimeout(ctx, compose.DefaultTimeout)
	defer pullCancel()

	_, err = executor.Pull(pullCtx)
	if err != nil {
		spinner.Fail(ui.Msg("pull_failed"))

		suggestion := ui.Msg("err_pull_internet")
		command := ""
		if ui.IsDockerPermissionError(err) {
			suggestion, command = ui.DockerPermissionSuggestion()
		}

		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:      ui.Msg("pull_failed"),
			Message:    ui.SanitizeError(err),
			Suggestion: suggestion,
			Command:    command,
		})
		return err
	}
	spinner.Success(ui.Msg("pulling_images"))

	updates, err := detectUpdatesAfterPull(ctx, imageState)
	if err != nil {
		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:      ui.Msg("pull_failed"),
			Message:    ui.SanitizeError(err),
			Suggestion: ui.Msg("err_check_docker_pull"),
		})
		return err
	}

	if len(updates) == 0 {
		ui.ShowOK(ui.Msg("images_up_to_date"))
		return nil
	}

	// Step 2: Show available updates
	ui.ShowStepHeader(2, 4, ui.Msg("updates_available"))
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

	confirmed, confirmErr := confirmUpdateRestart()
	if confirmErr != nil {
		return confirmErr
	}
	if !confirmed {
		return nil
	}

	// Step 3: Restart services with new images
	ui.ShowStepHeader(3, 4, ui.Msg("step_recreate"))

	recreateCtx, recreateCancel := context.WithTimeout(ctx, compose.DefaultTimeout)
	defer recreateCancel()

	restartSpinner := ui.StartPtermSpinner(ui.Msg("restarting"))
	err = executor.ForceRecreate(recreateCtx)
	if err != nil {
		restartSpinner.Fail(ui.Msg("restart_failed"))
		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:      ui.Msg("restart_failed"),
			Message:    ui.SanitizeError(err),
			Suggestion: ui.Msg("err_check_docker_logs"),
			Command:    ui.Msg("docker_compose_logs_command"),
		})
		return fmt.Errorf("%s: %w", ui.Msg("restart_failed"), err)
	}
	restartSpinner.Success(ui.Msg("restart_complete"))

	definedServices := imageState.composeFile.GetServiceNames()
	monitorUpdateHealth(recreateCtx, imageState.composeFile)

	// Step 4: Show status
	ui.ShowStepHeader(4, 4, ui.Msg("step_status"))

	// Show status
	statusCtx, statusCancel := context.WithTimeout(ctx, compose.DefaultTimeout)
	defer statusCancel()

	statuses, err := monitor.GetStatusWithServices(statusCtx, executor, definedServices)
	if err == nil {
		ui.PrintCommandResult(statuses, ui.Msg("cmd_update_title"), "update_summary_success", "update_summary_partial")
	}

	return nil
}

func confirmUpdateRestart() (bool, error) {
	if forceUpdate {
		return true, nil
	}

	var confirm bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(ui.Msg("confirm_restart")).
				Value(&confirm),
		),
	)
	if err := form.Run(); err != nil {
		return false, err
	}
	if !confirm {
		ui.ShowInfo(ui.Msg("update_cancelled"))
		return false, nil
	}
	return true, nil
}

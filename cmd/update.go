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
	"github.com/kkauto-net/kk-install/pkg/updater"
)

var updateCmd = &cobra.Command{
	Use:         "update",
	Short:       "Pull latest images and recreate containers",
	Long:        `Check and download new images from Docker Hub, then restart services.`,
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
			Message:    err.Error(),
			Suggestion: ui.Msg("run_init_to_configure"),
			Command:    "kk init",
		})
		return err
	}

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

	// Step 1: Pull new images
	ui.ShowStepHeader(1, 4, ui.Msg("step_pull_images"))
	spinner := ui.StartPtermSpinner(ui.Msg("pulling_images"))

	pullCtx, pullCancel := context.WithTimeout(ctx, compose.DefaultTimeout)
	defer pullCancel()

	output, err := executor.Pull(pullCtx)
	if err != nil {
		spinner.Fail(ui.Msg("pull_failed"))

		suggestion := "Check internet connection or Docker Hub status"
		command := ""
		if ui.IsDockerPermissionError(err) {
			suggestion, command = ui.DockerPermissionSuggestion()
		}

		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:      ui.Msg("pull_failed"),
			Message:    err.Error(),
			Suggestion: suggestion,
			Command:    command,
		})
		return err
	}
	spinner.Success(ui.Msg("pulling_images"))

	// Step 2: Parse pull output
	updates := updater.ParsePullOutput(output)

	if len(updates) == 0 {
		fmt.Println("\n[OK] " + ui.Msg("images_up_to_date"))
		return nil
	}

	// Step 2: Show updates with boxed table
	ui.ShowStepHeader(2, 4, ui.Msg("step_status"))
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

	// Confirm restart
	if !forceUpdate {
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

	// Step 3: Recreate containers
	ui.ShowStepHeader(3, 4, ui.Msg("step_recreate"))

	recreateCtx, recreateCancel := context.WithTimeout(ctx, compose.DefaultTimeout)
	defer recreateCancel()

	if err := executor.ForceRecreate(recreateCtx); err != nil {
		return fmt.Errorf("%s: %w", ui.Msg("recreate_failed"), err)
	}

	// Monitor health
	composeFile, err := compose.ParseComposeFile(cwd)
	var definedServices []string
	if err == nil {
		for name := range composeFile.Services {
			definedServices = append(definedServices, name)
		}

		healthMonitor, err := monitor.NewHealthMonitor()
		if err == nil {
			defer healthMonitor.Close()

			var containers []monitor.ContainerInfo
			for name := range composeFile.Services {
				containers = append(containers, monitor.ContainerInfo{
					ServiceName:    name,
					ContainerName:  fmt.Sprintf("kkengine_%s", name),
					HasHealthCheck: composeFile.HasHealthCheck(name),
				})
			}

			healthMonitor.MonitorAll(recreateCtx, containers, func(status monitor.HealthStatus) {
				ui.ShowServiceProgress(status.ServiceName, status.Status)
			})
		}
	}

	// Step 4: Show status
	ui.ShowStepHeader(4, 4, ui.Msg("step_status"))

	// Show status
	statusCtx, statusCancel := context.WithTimeout(ctx, compose.DefaultTimeout)
	defer statusCancel()

	statuses, err := monitor.GetStatusWithServices(statusCtx, executor, definedServices)
	if err == nil {
		ui.PrintCommandResult(statuses, "kk update", "update_summary_success", "update_summary_partial")
	}

	return nil
}

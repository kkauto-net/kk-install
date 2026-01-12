package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/kkauto-net/kk-install/pkg/compose"
	"github.com/kkauto-net/kk-install/pkg/config"
	"github.com/kkauto-net/kk-install/pkg/ui"
)

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove all containers, networks, and volumes",
	Long: `Remove all containers, networks, and volumes created by docker-compose.
This is useful when you need to clean up containers from other directories
or resolve container name conflicts.`,
	Annotations: map[string]string{"group": "management"},
	RunE:        runRemove,
}

var removeVolumes bool

func init() {
	removeCmd.Flags().BoolVarP(&removeVolumes, "volumes", "v", false, "Also remove volumes (WARNING: deletes data)")
	rootCmd.AddCommand(removeCmd)
}

func runRemove(cmd *cobra.Command, args []string) error {
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

	// Step 1: Remove containers and networks
	ui.ShowStepHeader(1, 1, ui.Msg("step_remove_services"))

	executor := compose.NewExecutor(cwd)

	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, compose.DefaultTimeout)
	defer timeoutCancel()

	spinner := ui.StartPtermSpinner(ui.Msg("removing_services"))

	var execErr error
	if removeVolumes {
		execErr = executor.DownWithVolumes(timeoutCtx)
	} else {
		execErr = executor.Down(timeoutCtx)
	}

	if execErr != nil {
		spinner.Fail(ui.Msg("remove_failed"))

		suggestion := "Check Docker logs for details"
		command := "docker compose logs"
		if ui.IsDockerPermissionError(execErr) {
			suggestion, command = ui.DockerPermissionSuggestion()
		}

		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:      ui.Msg("remove_failed"),
			Message:    execErr.Error(),
			Suggestion: suggestion,
			Command:    command,
		})
		return execErr
	}
	spinner.Success(ui.Msg("remove_complete"))

	if removeVolumes {
		ui.ShowSuccess(ui.Msg("removed_with_volumes"))
	} else {
		ui.ShowSuccess(ui.Msg("removed_containers"))
	}

	return nil
}

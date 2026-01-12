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
	"github.com/kkauto-net/kk-install/pkg/monitor"
	"github.com/kkauto-net/kk-install/pkg/ui"
)

var restartCmd = &cobra.Command{
	Use:         "restart",
	Short:       "Restart all services",
	Long:        `Restart all containers in the stack.`,
	Annotations: map[string]string{"group": "management"},
	RunE:        runRestart,
}

func init() {
	rootCmd.AddCommand(restartCmd)
}

func runRestart(cmd *cobra.Command, args []string) error {
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

	// Step 1: Restart services
	ui.ShowStepHeader(1, 3, ui.Msg("step_start_services"))

	executor := compose.NewExecutor(cwd)

	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, compose.DefaultTimeout)
	defer timeoutCancel()

	spinner := ui.StartPtermSpinner(ui.Msg("restarting"))
	if err := executor.Restart(timeoutCtx); err != nil {
		spinner.Fail(ui.Msg("restart_failed"))

		suggestion := "Check if services are running"
		command := "kk status"
		if ui.IsDockerPermissionError(err) {
			suggestion, command = ui.DockerPermissionSuggestion()
		}

		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:      ui.Msg("restart_failed"),
			Message:    err.Error(),
			Suggestion: suggestion,
			Command:    command,
		})
		return err
	}
	spinner.Success(ui.Msg("restart_complete"))

	// Step 2: Monitor health
	ui.ShowStepHeader(2, 3, ui.Msg("step_health_check"))
	composeFile, err := compose.ParseComposeFile(cwd)
	var definedServices []string
	if err == nil {
		for name := range composeFile.Services {
			definedServices = append(definedServices, name)
		}

		healthMonitor, err := monitor.NewHealthMonitor()
		if err == nil {
			defer healthMonitor.Close()

			fmt.Println("\n" + ui.Msg("health_checking"))

			var containers []monitor.ContainerInfo
			for name := range composeFile.Services {
				containers = append(containers, monitor.ContainerInfo{
					ServiceName:    name,
					ContainerName:  fmt.Sprintf("kkengine_%s", name),
					HasHealthCheck: composeFile.HasHealthCheck(name),
				})
			}

			healthMonitor.MonitorAll(timeoutCtx, containers, func(status monitor.HealthStatus) {
				ui.ShowServiceProgress(status.ServiceName, status.Status)
			})
		}
	}

	// Step 3: Show final status
	ui.ShowStepHeader(3, 3, ui.Msg("step_status"))
	statuses, err := monitor.GetStatusWithServices(timeoutCtx, executor, definedServices)
	if err == nil {
		ui.PrintCommandResult(statuses, "kk restart", "restart_summary_success", "restart_summary_partial")
	}

	return nil
}

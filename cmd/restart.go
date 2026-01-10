package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/kkauto-net/kk-install/pkg/compose"
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
	// Command banner
	ui.ShowCommandBanner("kk restart", ui.Msg("restart_desc"))

	cwd, err := os.Getwd()
	if err != nil {
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
		return fmt.Errorf("%s: %w", ui.Msg("restart_failed"), err)
	}
	spinner.Success(ui.Msg("restart_complete"))

	// Step 2: Monitor health
	ui.ShowStepHeader(2, 3, ui.Msg("step_health_check"))
	composeFile, err := compose.ParseComposeFile(cwd)
	if err == nil {
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
	statuses, err := monitor.GetStatus(timeoutCtx, executor)
	if err == nil {
		ui.PrintStatusTable(statuses)
	}

	return nil
}

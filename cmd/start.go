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
	"github.com/kkauto-net/kk-install/pkg/validator"
)

var startCmd = &cobra.Command{
	Use:         "start",
	Short:       "Start all services with preflight checks",
	Long:        `Run preflight checks, then start all services.`,
	Annotations: map[string]string{"group": "core"},
	RunE:        runStart,
}

func init() {
	rootCmd.AddCommand(startCmd)
}

func runStart(cmd *cobra.Command, args []string) error {
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

	// Step 1: Detect if Caddy is enabled
	composeFile, err := compose.ParseComposeFile(cwd)
	includeCaddy := false
	var definedServices []string
	if err == nil {
		_, includeCaddy = composeFile.Services["caddy"]
		for name := range composeFile.Services {
			definedServices = append(definedServices, name)
		}
	}

	// Step 1: Run preflight checks
	ui.ShowStepHeader(1, 4, ui.Msg("step_preflight"))
	results, err := validator.RunPreflight(cwd, includeCaddy)
	validator.PrintPreflightResults(results)

	if err != nil {
		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:      ui.Msg("preflight_failed"),
			Message:    ui.Msg("preflight_checks_failed"),
			Suggestion: "Fix the issues above and try again",
		})
		return err
	}

	// Step 2: Start docker-compose
	ui.ShowStepHeader(2, 4, ui.Msg("step_start_services"))
	executor := compose.NewExecutor(cwd)

	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, compose.DefaultTimeout)
	defer timeoutCancel()

	spinner := ui.StartPtermSpinner(ui.Msg("starting_services"))
	if err := executor.Up(timeoutCtx); err != nil {
		spinner.Fail(ui.Msg("start_failed"))

		suggestion := "Check Docker logs for details"
		command := "docker compose logs"

		// Check for specific error types and provide better suggestions
		if ui.IsDockerPermissionError(err) {
			suggestion, command = ui.DockerPermissionSuggestion()
		} else if ui.IsContainerConflictError(err) {
			suggestion, command = ui.ContainerConflictSuggestion()
		}

		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:      ui.Msg("start_failed"),
			Message:    err.Error(),
			Suggestion: suggestion,
			Command:    command,
		})
		return err
	}
	spinner.Success(ui.Msg("services_started"))

	// Step 3: Monitor health
	ui.ShowStepHeader(3, 4, ui.Msg("step_health_check"))

	healthMonitor, err := monitor.NewHealthMonitor()
	if err != nil {
		// Can't monitor, but services may still be running
		fmt.Printf("  [!] %s: %v\n", ui.Msg("health_failed"), err)
	} else {
		defer healthMonitor.Close()

		// Build container list
		var containers []monitor.ContainerInfo
		for name := range composeFile.Services {
			containers = append(containers, monitor.ContainerInfo{
				ServiceName:    name,
				ContainerName:  fmt.Sprintf("kkengine_%s", name),
				HasHealthCheck: composeFile.HasHealthCheck(name),
			})
		}

		// Monitor with progress callback
		healthResults := healthMonitor.MonitorAll(timeoutCtx, containers, func(status monitor.HealthStatus) {
			ui.ShowServiceProgress(status.ServiceName, status.Status)
		})

		// Check if all healthy
		allHealthy := true
		for _, r := range healthResults {
			if !r.Healthy {
				allHealthy = false
				break
			}
		}

		if !allHealthy {
			fmt.Println("\n[!] " + ui.Msg("some_not_ready"))
		}
	}

	// Step 4: Show status
	ui.ShowStepHeader(4, 4, ui.Msg("step_status"))

	statuses, err := monitor.GetStatusWithServices(timeoutCtx, executor, definedServices)
	if err == nil {
		ui.PrintCommandResult(statuses, "kk start", "start_summary_success", "start_summary_partial")
		domain := config.ReadEnvValue(cwd, "SYSTEM_DOMAIN")
		ui.PrintAccessInfo(statuses, domain)
	}

	return nil
}

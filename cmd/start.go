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

	// Step 1: Detect if Caddy is enabled
	composeFile, err := compose.ParseComposeFile(cwd)
	includeCaddy := false
	if err == nil {
		_, includeCaddy = composeFile.Services["caddy"]
	}

	// Step 2: Run preflight checks
	fmt.Println("\n" + ui.Msg("preflight_checking"))
	results, err := validator.RunPreflight(cwd, includeCaddy)
	validator.PrintPreflightResults(results)

	if err != nil {
		return fmt.Errorf("%s", ui.Msg("preflight_failed"))
	}

	// Step 3: Start docker-compose
	fmt.Println(ui.Msg("starting_services"))
	executor := compose.NewExecutor(cwd)

	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, compose.DefaultTimeout)
	defer timeoutCancel()

	if err := executor.Up(timeoutCtx); err != nil {
		return fmt.Errorf("%s: %w", ui.Msg("start_failed"), err)
	}

	// Step 4: Monitor health
	fmt.Println("\n" + ui.Msg("health_checking"))

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

	// Step 5: Show status
	fmt.Println("\n[OK] " + ui.Msg("start_complete"))

	statuses, err := monitor.GetStatus(timeoutCtx, executor)
	if err == nil {
		ui.PrintStatusTable(statuses)
		ui.PrintAccessInfo(statuses)
	}

	return nil
}

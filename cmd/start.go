package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/kkengine/kkcli/pkg/compose"
	"github.com/kkengine/kkcli/pkg/monitor"
	"github.com/kkengine/kkcli/pkg/ui"
	"github.com/kkengine/kkcli/pkg/validator"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Khoi dong kkengine Docker stack",
	Long:  `Chay preflight checks, sau do khoi dong tat ca services.`,
	RunE:  runStart,
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
		fmt.Println("\n\nDang dung lai...")
		cancel()
	}()

	// Step 1: Detect if Caddy is enabled
	composeFile, err := compose.ParseComposeFile(cwd)
	includeCaddy := false
	if err == nil {
		_, includeCaddy = composeFile.Services["caddy"]
	}

	// Step 2: Run preflight checks
	fmt.Println("\nKiem tra truoc khi chay...")
	results, err := validator.RunPreflight(cwd, includeCaddy)
	validator.PrintPreflightResults(results)

	if err != nil {
		return fmt.Errorf("preflight checks that bai. Vui long sua loi tren")
	}

	// Step 3: Start docker-compose
	fmt.Println("Khoi dong services...")
	executor := compose.NewExecutor(cwd)

	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, compose.DefaultTimeout)
	defer timeoutCancel()

	if err := executor.Up(timeoutCtx); err != nil {
		return fmt.Errorf("khoi dong that bai: %w", err)
	}

	// Step 4: Monitor health
	fmt.Println("\nDang kiem tra suc khoe dich vu...")

	healthMonitor, err := monitor.NewHealthMonitor()
	if err != nil {
		// Can't monitor, but services may still be running
		fmt.Printf("  [!] Khong the theo doi health: %v\n", err)
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
			fmt.Println("\n[!] Mot so dich vu chua san sang. Kiem tra: kk status")
		}
	}

	// Step 5: Show status
	fmt.Println("\n[OK] Khoi dong hoan tat!")

	statuses, err := monitor.GetStatus(timeoutCtx, executor)
	if err == nil {
		ui.PrintStatusTable(statuses)
		ui.PrintAccessInfo(statuses)
	}

	return nil
}

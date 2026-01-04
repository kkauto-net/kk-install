package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/kkengine/kkcli/pkg/compose"
	"github.com/kkengine/kkcli/pkg/monitor"
	"github.com/kkengine/kkcli/pkg/ui"
	"github.com/kkengine/kkcli/pkg/updater"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Cap nhat images moi nhat",
	Long:  `Kiem tra va tai images moi tu Docker Hub, sau do restart services.`,
	RunE:  runUpdate,
}

var forceUpdate bool

func init() {
	updateCmd.Flags().BoolVarP(&forceUpdate, "force", "f", false, "Khong hoi xac nhan")
	rootCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
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

	executor := compose.NewExecutor(cwd)

	// Step 1: Pull new images
	fmt.Println("Dang kiem tra cap nhat...")
	spinner := ui.NewSpinner("Dang tai images...")
	spinner.Start()

	pullCtx, pullCancel := context.WithTimeout(ctx, compose.DefaultTimeout)
	defer pullCancel()

	output, err := executor.Pull(pullCtx)
	spinner.Stop(err == nil)

	if err != nil {
		return fmt.Errorf("khong tai duoc images: %w", err)
	}

	// Step 2: Parse pull output
	updates := updater.ParsePullOutput(output)

	if len(updates) == 0 {
		fmt.Println("\n[OK] Tat ca images da la phien ban moi nhat.")
		return nil
	}

	// Step 3: Show updates
	fmt.Println("\nCo cap nhat:")
	for _, u := range updates {
		fmt.Printf("  - %s\n", u.Image)
		if u.OldDigest != "" && u.NewDigest != "" {
			oldDigest := u.OldDigest
			if len(oldDigest) > 12 {
				oldDigest = oldDigest[:12]
			}
			newDigest := u.NewDigest
			if len(newDigest) > 12 {
				newDigest = newDigest[:12]
			}
			fmt.Printf("    %s -> %s\n", oldDigest, newDigest)
		}
	}
	fmt.Println()

	// Step 4: Confirm restart
	if !forceUpdate {
		var confirm bool
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Khoi dong lai services voi images moi?").
					Value(&confirm),
			),
		)

		if err := form.Run(); err != nil {
			return err
		}

		if !confirm {
			fmt.Println("Huy cap nhat. Images da duoc tai, chay 'kk restart' de ap dung.")
			return nil
		}
	}

	// Step 5: Recreate containers
	fmt.Println("Dang khoi dong lai voi images moi...")

	recreateCtx, recreateCancel := context.WithTimeout(ctx, compose.DefaultTimeout)
	defer recreateCancel()

	if err := executor.ForceRecreate(recreateCtx); err != nil {
		return fmt.Errorf("recreate that bai: %w", err)
	}

	// Step 6: Monitor health
	composeFile, err := compose.ParseComposeFile(cwd)
	if err == nil {
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

	fmt.Println("\n[OK] Cap nhat hoan tat!")

	// Show status
	statusCtx, statusCancel := context.WithTimeout(ctx, compose.DefaultTimeout)
	defer statusCancel()

	statuses, err := monitor.GetStatus(statusCtx, executor)
	if err == nil {
		ui.PrintStatusTable(statuses)
	}

	return nil
}

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
	"github.com/kkauto-net/kk-install/pkg/monitor"
	"github.com/kkauto-net/kk-install/pkg/ui"
	"github.com/kkauto-net/kk-install/pkg/updater"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Cap nhat images moi nhat",
	Long:  `Kiem tra va tai images moi tu Docker Hub, sau do restart services.`,
	RunE:  runUpdate,
}

var forceUpdate bool

func init() {
	updateCmd.Flags().BoolVarP(&forceUpdate, "force", "f", false, "Skip confirmation prompts")
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
		fmt.Println("\n\n" + ui.Msg("stopping"))
		cancel()
	}()

	executor := compose.NewExecutor(cwd)

	// Step 1: Pull new images
	fmt.Println(ui.Msg("checking_updates"))
	spinner := ui.NewSpinner(ui.Msg("pulling_images"))
	spinner.Start()

	pullCtx, pullCancel := context.WithTimeout(ctx, compose.DefaultTimeout)
	defer pullCancel()

	output, err := executor.Pull(pullCtx)
	spinner.Stop(err == nil)

	if err != nil {
		return fmt.Errorf("%s: %w", ui.Msg("pull_failed"), err)
	}

	// Step 2: Parse pull output
	updates := updater.ParsePullOutput(output)

	if len(updates) == 0 {
		fmt.Println("\n[OK] " + ui.Msg("images_up_to_date"))
		return nil
	}

	// Step 3: Show updates
	fmt.Println("\n" + ui.Msg("updates_available"))
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

	// Step 5: Recreate containers
	fmt.Println(ui.Msg("recreating"))

	recreateCtx, recreateCancel := context.WithTimeout(ctx, compose.DefaultTimeout)
	defer recreateCancel()

	if err := executor.ForceRecreate(recreateCtx); err != nil {
		return fmt.Errorf("%s: %w", ui.Msg("recreate_failed"), err)
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

	fmt.Println("\n[OK] " + ui.Msg("update_complete"))

	// Show status
	statusCtx, statusCancel := context.WithTimeout(ctx, compose.DefaultTimeout)
	defer statusCancel()

	statuses, err := monitor.GetStatus(statusCtx, executor)
	if err == nil {
		ui.PrintStatusTable(statuses)
	}

	return nil
}

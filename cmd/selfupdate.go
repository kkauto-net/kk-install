package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/kkauto-net/kk-install/pkg/selfupdate"
	"github.com/kkauto-net/kk-install/pkg/ui"
)

var selfupdateCmd = &cobra.Command{
	Use:         "selfupdate",
	Short:       "Update kk CLI to the latest version",
	Long:        `Check for and install the latest version of kk CLI from GitHub releases.`,
	Annotations: map[string]string{"group": "management"},
	RunE:        runSelfupdate,
}

var (
	checkOnly        bool
	forceSelfupdate  bool
)

func init() {
	selfupdateCmd.Flags().BoolVarP(&checkOnly, "check", "c", false, "Only check for updates, don't install")
	selfupdateCmd.Flags().BoolVarP(&forceSelfupdate, "force", "f", false, "Skip confirmation prompts")
	rootCmd.AddCommand(selfupdateCmd)
}

func runSelfupdate(cmd *cobra.Command, args []string) error {
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

	// Step 1: Check for updates
	ui.ShowStepHeader(1, 2, ui.Msg("step_check_update"))
	spinner := ui.StartPtermSpinner(ui.Msg("checking_cli_update"))

	checkCtx, checkCancel := context.WithTimeout(ctx, 30*time.Second)
	defer checkCancel()

	result, err := selfupdate.CheckForUpdate(checkCtx, Version)
	if err != nil {
		spinner.Fail(ui.Msg("check_update_failed"))
		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:      ui.Msg("check_update_failed"),
			Message:    err.Error(),
			Suggestion: "Check internet connection or try again later",
		})
		return err
	}
	spinner.Success(ui.Msg("checking_cli_update"))

	// Show version info
	fmt.Println()
	fmt.Printf("  %s: %s\n", ui.Msg("current_version"), result.CurrentVersion)
	fmt.Printf("  %s: %s\n", ui.Msg("latest_version"), result.LatestVersion)
	fmt.Println()

	if !result.UpdateNeeded {
		ui.ShowSuccess(ui.Msg("cli_up_to_date"))
		return nil
	}

	// Update available
	ui.ShowInfo(ui.Msg("update_available"))

	if checkOnly {
		fmt.Println()
		fmt.Printf("  %s: kk selfupdate\n", ui.Msg("to_update_run"))
		return nil
	}

	// Confirm update
	if !forceSelfupdate {
		var confirm bool
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title(ui.Msg("confirm_cli_update")).
					Value(&confirm),
			),
		)

		if err := form.Run(); err != nil {
			return err
		}

		if !confirm {
			fmt.Println(ui.Msg("selfupdate_cancelled"))
			return nil
		}
	}

	// Step 2: Download and install
	ui.ShowStepHeader(2, 2, ui.Msg("step_install_update"))
	spinner = ui.StartPtermSpinner(ui.Msg("downloading_update"))

	updateCtx, updateCancel := context.WithTimeout(ctx, 5*time.Minute)
	defer updateCancel()

	if err := selfupdate.Update(updateCtx, result); err != nil {
		spinner.Fail(ui.Msg("update_install_failed"))
		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:      ui.Msg("update_install_failed"),
			Message:    err.Error(),
			Suggestion: "Try running with sudo or check permissions",
		})
		return err
	}
	spinner.Success(ui.Msg("downloading_update"))

	// Success message
	fmt.Println()
	ui.ShowSuccess(fmt.Sprintf("%s %s → %s", ui.Msg("selfupdate_complete"), result.CurrentVersion, result.LatestVersion))

	return nil
}

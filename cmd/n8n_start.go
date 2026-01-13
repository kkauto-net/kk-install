package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kkauto-net/kk-install/pkg/compose"
	"github.com/kkauto-net/kk-install/pkg/n8n"
	"github.com/kkauto-net/kk-install/pkg/ui"
	"github.com/kkauto-net/kk-install/pkg/validator"
)

var n8nStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start n8n services",
	Long:  `Start n8n and PostgreSQL database containers.`,
	RunE:  runN8nStart,
}

func init() {
	n8nCmd.AddCommand(n8nStartCmd)
}

func runN8nStart(cmd *cobra.Command, args []string) error {
	return runN8nStartInternal()
}

// runN8nStartInternal starts n8n services (called from install and start commands)
func runN8nStartInternal() error {
	// Check installation
	if !n8n.IsInstalled() {
		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:      "n8n Not Installed",
			Message:    ui.Msg("n8n_not_installed"),
			Suggestion: "Install n8n first",
			Command:    "kk n8n install",
		})
		return fmt.Errorf("n8n not installed")
	}

	// Step 1: Preflight check
	ui.ShowStepHeader(1, 3, ui.Msg("step_preflight"))
	dv := validator.NewDockerValidator()
	if err := dv.CheckDockerDaemon(); err != nil {
		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:      ui.Msg("docker_not_running"),
			Message:    err.Error(),
			Suggestion: ui.Msg("docker_start_suggestion"),
			Command:    "sudo systemctl start docker",
		})
		return err
	}

	// Check port 5678 availability
	portStatus := validator.CheckPort(5678)
	if portStatus.InUse && !portStatus.UsedByKKEngine {
		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:      "Port 5678 In Use",
			Message:    fmt.Sprintf("Port 5678 is being used by %s (PID %d)", portStatus.Process, portStatus.PID),
			Suggestion: "Stop the service using port 5678 or change n8n port",
		})
		return fmt.Errorf("port 5678 in use")
	}
	ui.ShowSuccess(ui.Msg("docker_ok"))

	// Step 2: Start services
	ui.ShowStepHeader(2, 3, ui.Msg("step_start_services"))
	ctx, cancel := context.WithTimeout(context.Background(), compose.DefaultTimeout)
	defer cancel()

	executor := compose.NewExecutor(n8n.N8nDir())
	spinner := ui.StartPtermSpinner(ui.Msg("n8n_starting"))

	if err := executor.Up(ctx); err != nil {
		spinner.Fail(ui.Msg("start_failed"))
		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:   "Failed to start n8n",
			Message: err.Error(),
		})
		return err
	}
	spinner.Success(ui.Msg("n8n_started"))

	// Step 3: Show status
	ui.ShowStepHeader(3, 3, ui.Msg("step_status"))
	fmt.Println()
	fmt.Println("  n8n is running at: http://localhost:5678")
	fmt.Println("  Run 'kk n8n status' for details")

	return nil
}

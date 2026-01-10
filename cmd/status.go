package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/kkauto-net/kk-install/pkg/compose"
	"github.com/kkauto-net/kk-install/pkg/monitor"
	"github.com/kkauto-net/kk-install/pkg/ui"
)

var statusCmd = &cobra.Command{
	Use:         "status",
	Short:       "View service status and health",
	Long:        `Display status of all containers in the stack.`,
	Annotations: map[string]string{"group": "core"},
	RunE:        runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Parse compose file to get defined services
	composeFile, err := compose.ParseComposeFile(cwd)
	if err != nil {
		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:      ui.Msg("get_status_failed"),
			Message:    err.Error(),
			Suggestion: "Make sure docker-compose.yml exists",
			Command:    "kk init",
		})
		return err
	}

	definedServices := composeFile.GetServiceNames()
	if len(definedServices) == 0 {
		fmt.Println(ui.Msg("no_services_defined"))
		fmt.Println(ui.Msg("run_init"))
		return nil
	}

	executor := compose.NewExecutor(cwd)
	statuses, err := monitor.GetStatusWithServices(ctx, executor, definedServices)
	if err != nil {
		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:      ui.Msg("get_status_failed"),
			Message:    err.Error(),
			Suggestion: "Check if Docker is running",
			Command:    "docker ps",
		})
		return err
	}

	ui.PrintStatusTable(statuses)

	// Show access info if any services running
	for _, s := range statuses {
		if s.Running {
			ui.PrintAccessInfo(statuses)
			break
		}
	}

	return nil
}

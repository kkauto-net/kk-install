package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/kkauto-net/kk-install/pkg/compose"
	"github.com/kkauto-net/kk-install/pkg/config"
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
		errMsg := err.Error()
		suggestion := "Check if Docker is running"
		command := "sudo systemctl start docker"

		// Check for permission denied error
		if strings.Contains(strings.ToLower(errMsg), "permission denied") {
			suggestion = "Add user to docker group"
			command = "sudo usermod -aG docker $USER && newgrp docker"
		}

		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:      ui.Msg("get_status_failed"),
			Message:    errMsg,
			Suggestion: suggestion,
			Command:    command,
		})
		return err
	}

	ui.PrintStatusTable(statuses)

	// Check for unhealthy services and show hint
	var unhealthyServices []string
	for _, s := range statuses {
		if s.Running && s.Health == "unhealthy" {
			unhealthyServices = append(unhealthyServices, s.Name)
		}
	}
	if len(unhealthyServices) > 0 {
		fmt.Println()
		fmt.Printf("  [!] %d service(s) unhealthy: %s\n", len(unhealthyServices), strings.Join(unhealthyServices, ", "))
		fmt.Printf("      View logs: docker compose logs %s\n", unhealthyServices[0])
	}

	// Show access info if any services running
	for _, s := range statuses {
		if s.Running {
			ui.PrintAccessInfo(statuses)
			break
		}
	}

	return nil
}

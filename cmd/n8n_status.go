package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/kkauto-net/kk-install/pkg/compose"
	"github.com/kkauto-net/kk-install/pkg/monitor"
	"github.com/kkauto-net/kk-install/pkg/n8n"
	"github.com/kkauto-net/kk-install/pkg/ui"
)

var n8nStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show n8n service status",
	Long:  `Display status of n8n and PostgreSQL containers.`,
	RunE:  runN8nStatus,
}

func init() {
	n8nCmd.AddCommand(n8nStatusCmd)
}

func runN8nStatus(cmd *cobra.Command, args []string) error {
	if !n8n.IsInstalled() {
		fmt.Println(ui.Msg("n8n_not_installed"))
		fmt.Println("Run 'kk n8n install' to set up n8n.")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	executor := compose.NewExecutor(n8n.N8nDir())

	// Get status for n8n services
	definedServices := []string{"n8n", "n8n-postgres"}
	statuses, err := monitor.GetStatusWithServices(ctx, executor, definedServices)
	if err != nil {
		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:   ui.Msg("get_status_failed"),
			Message: err.Error(),
		})
		return err
	}

	ui.PrintStatusTable(statuses)

	// Show access info if n8n is running
	for _, s := range statuses {
		if s.Name == "n8n" && s.Running {
			fmt.Println()
			fmt.Println("  Access n8n at: http://localhost:5678")
			break
		}
	}

	return nil
}

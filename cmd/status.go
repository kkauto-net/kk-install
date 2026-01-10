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

	executor := compose.NewExecutor(cwd)
	statuses, err := monitor.GetStatus(ctx, executor)
	if err != nil {
		return fmt.Errorf("%s: %w", ui.Msg("get_status_failed"), err)
	}

	if len(statuses) == 0 {
		fmt.Println(ui.Msg("no_services"))
		fmt.Println(ui.Msg("run_start"))
		return nil
	}

	ui.PrintStatusTable(statuses)
	ui.PrintAccessInfo(statuses)

	// Summary
	running := 0
	for _, s := range statuses {
		if s.Running {
			running++
		}
	}

	if running == len(statuses) {
		ui.ShowSuccess(ui.MsgF("all_running", running))
	} else {
		ui.ShowWarning(ui.MsgF("some_running", running, len(statuses)))
	}

	return nil
}

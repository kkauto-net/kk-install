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
	Use:   "status",
	Short: "Xem trang thai dich vu",
	Long:  `Hien thi trang thai tat ca containers trong stack.`,
	RunE:  runStatus,
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
		fmt.Printf("[OK] "+ui.Msg("all_running")+"\n", running)
	} else {
		fmt.Printf("[!] "+ui.Msg("some_running")+"\n", running, len(statuses))
	}

	return nil
}

package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/kkengine/kkcli/pkg/compose"
	"github.com/kkengine/kkcli/pkg/monitor"
	"github.com/kkengine/kkcli/pkg/ui"
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
		return fmt.Errorf("khong lay duoc trang thai: %w", err)
	}

	if len(statuses) == 0 {
		fmt.Println("Khong co dich vu nao dang chay.")
		fmt.Println("Chay: kk start")
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
		fmt.Printf("[OK] Tat ca %d dich vu dang chay.\n", running)
	} else {
		fmt.Printf("[!] %d/%d dich vu dang chay.\n", running, len(statuses))
	}

	return nil
}

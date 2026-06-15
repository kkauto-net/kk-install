package cmd

import (
	"github.com/spf13/cobra"

	"github.com/kkauto-net/kk-install/pkg/config"
	"github.com/kkauto-net/kk-install/pkg/ui"
)

var configCmd = &cobra.Command{
	Use:         "config",
	Short:       "Manage CLI configuration",
	Long:        `View and manage kk CLI configuration settings.`,
	Annotations: map[string]string{"group": "core"},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  `Display the current CLI configuration including language and project directory.`,
	RunE:  runConfigShow,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:      ui.Msg("config_load_failed"),
			Message:    ui.SanitizeError(err),
			Suggestion: ui.Msg("run_init_to_configure"),
			Command:    "kk init",
		})
		return err
	}

	ui.PrintConfigSummary(cfg)
	return nil
}

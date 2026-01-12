package cmd

import (
	"fmt"

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
			Message:    err.Error(),
			Suggestion: ui.Msg("run_init_to_configure"),
			Command:    "kk init",
		})
		return err
	}

	// Display configuration
	fmt.Println()
	fmt.Println(ui.Msg("config_title"))
	fmt.Println()

	// Language
	langDisplay := cfg.Language
	if langDisplay == "vi" {
		langDisplay = "Tiếng Việt"
	} else {
		langDisplay = "English"
	}
	fmt.Printf("  %s: %s\n", ui.Msg("config_language"), langDisplay)

	// Project directory
	projectDir := cfg.ProjectDir
	if projectDir == "" {
		projectDir = ui.Msg("config_not_set")
	}
	fmt.Printf("  %s: %s\n", ui.Msg("config_project_dir"), projectDir)

	// Config file path
	fmt.Printf("  %s: %s\n", ui.Msg("config_file_path"), config.ConfigPath())

	fmt.Println()
	return nil
}

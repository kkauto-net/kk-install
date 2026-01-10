package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/kkauto-net/kk-install/pkg/config"
	"github.com/kkauto-net/kk-install/pkg/ui"
)

var Version = "0.1.0"

var rootCmd = &cobra.Command{
	Use:   "kk",
	Short: "🚀 Manage your kkengine Docker stack effortlessly",
	Long:  `🚀 Manage your kkengine Docker stack effortlessly.`,
}

func Execute() {
	// Apply custom help templates (after all subcommands are registered)
	ui.ApplyTemplates(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Version = Version

	// Load language preference from config
	cfg, err := config.Load()
	if err == nil && cfg != nil {
		ui.SetLanguage(ui.Language(cfg.Language))
	}
	// If load fails, ui package already defaults to English
}

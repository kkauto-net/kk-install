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
	Short: "KK CLI - Docker Compose management for kkengine",
	Long:  `KK CLI giup ban quan ly kkengine Docker stack de dang.`,
}

func Execute() {
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

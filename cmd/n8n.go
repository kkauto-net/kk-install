package cmd

import (
	"github.com/spf13/cobra"
)

var n8nCmd = &cobra.Command{
	Use:   "n8n",
	Short: "Manage n8n workflow automation",
	Long:  `Install, start, stop, and manage n8n workflow automation platform.`,
}

func init() {
	rootCmd.AddCommand(n8nCmd)
}

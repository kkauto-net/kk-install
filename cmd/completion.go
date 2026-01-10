package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for bash, zsh, or fish.

Bash:
  $ source <(kk completion bash)
  # Or add to ~/.bashrc:
  $ kk completion bash > /etc/bash_completion.d/kk

Zsh:
  $ source <(kk completion zsh)
  # Or add to ~/.zshrc:
  $ kk completion zsh > "${fpath[1]}/_kk"

Fish:
  $ kk completion fish | source
  # Or save to:
  $ kk completion fish > ~/.config/fish/completions/kk.fish
`,
	Annotations:           map[string]string{"group": "additional"},
	ValidArgs:             []string{"bash", "zsh", "fish"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	DisableFlagsInUseLine: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}

package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish]",
	Short: "Tao shell completion script",
	Long: `Tao shell completion script cho bash, zsh, hoac fish.

Bash:
  $ source <(kk completion bash)
  # Hoac them vao ~/.bashrc:
  $ kk completion bash > /etc/bash_completion.d/kk

Zsh:
  $ source <(kk completion zsh)
  # Hoac them vao ~/.zshrc:
  $ kk completion zsh > "${fpath[1]}/_kk"

Fish:
  $ kk completion fish | source
  # Hoac luu vao:
  $ kk completion fish > ~/.config/fish/completions/kk.fish
`,
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

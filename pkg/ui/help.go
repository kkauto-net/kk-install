package ui

import (
	"strings"

	"github.com/spf13/cobra"
)

// CommandGroup represents a group of commands
type CommandGroup struct {
	Title    string
	Commands []*cobra.Command
}

// HelpTemplate is the custom help template (GitHub CLI style)
const HelpTemplate = `{{with .Long}}{{. | trim}}{{else}}{{.Short | trim}}{{end}}

USAGE
  {{.UseLine}}
{{if .HasAvailableSubCommands}}
{{- range $group := groupCommands .Commands}}
{{$group.Title}}{{range $group.Commands}}
  {{rpad .Name 12}}{{.Short}}{{end}}

{{end}}{{end}}{{if .HasAvailableLocalFlags}}FLAGS
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}

{{end}}LEARN MORE
  Use '{{.CommandPath}} <command> --help' for more information
`

// UsageTemplate is the custom usage template
const UsageTemplate = `USAGE
  {{.UseLine}}{{if .HasAvailableSubCommands}}

Use '{{.CommandPath}} <command> --help' for more information about a command.{{end}}
`

// SubcommandHelpTemplate for individual commands
const SubcommandHelpTemplate = `{{with .Long}}{{. | trim}}{{else}}{{.Short | trim}}{{end}}

USAGE
  {{.UseLine}}{{if .HasAvailableFlags}}

FLAGS
{{.Flags.FlagUsages | trimTrailingWhitespaces}}{{end}}
`

// ApplyTemplates applies custom help/usage templates to root command
func ApplyTemplates(rootCmd *cobra.Command) {
	cobra.AddTemplateFunc("trim", strings.TrimSpace)
	cobra.AddTemplateFunc("groupCommands", groupCommands)

	rootCmd.SetHelpTemplate(HelpTemplate)
	rootCmd.SetUsageTemplate(UsageTemplate)

	// Apply subcommand template to all subcommands
	for _, cmd := range rootCmd.Commands() {
		cmd.SetHelpTemplate(SubcommandHelpTemplate)
	}
}

// groupCommands groups commands by their "group" annotation
func groupCommands(commands []*cobra.Command) []CommandGroup {
	groups := map[string][]*cobra.Command{
		"core":       {},
		"management": {},
		"additional": {},
	}

	groupOrder := []string{"core", "management", "additional"}
	groupTitles := map[string]string{
		"core":       "CORE COMMANDS",
		"management": "MANAGEMENT COMMANDS",
		"additional": "ADDITIONAL COMMANDS",
	}

	for _, cmd := range commands {
		if !cmd.IsAvailableCommand() || cmd.IsAdditionalHelpTopicCommand() {
			continue
		}

		group := cmd.Annotations["group"]
		if group == "" {
			group = "additional" // default group
		}

		groups[group] = append(groups[group], cmd)
	}

	var result []CommandGroup
	for _, g := range groupOrder {
		if len(groups[g]) > 0 {
			result = append(result, CommandGroup{
				Title:    groupTitles[g],
				Commands: groups[g],
			})
		}
	}

	return result
}

// Package ui provides help templates with ClaudeKit CLI-style colors.
package ui

import (
	"fmt"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

// CommandGroup represents a group of commands
type CommandGroup struct {
	Title    string
	Commands []*cobra.Command
}

// ApplyTemplates applies custom help/usage templates to root command
func ApplyTemplates(rootCmd *cobra.Command) {
	cobra.AddTemplateFunc("trim", strings.TrimSpace)
	cobra.AddTemplateFunc("groupCommands", groupCommands)

	// Use custom help function for colored output
	rootCmd.SetHelpFunc(customHelpFunc)
	rootCmd.SetUsageFunc(customUsageFunc)

	// Apply subcommand help to all subcommands
	for _, cmd := range rootCmd.Commands() {
		cmd.SetHelpFunc(subcommandHelpFunc)
	}
}

// customHelpFunc renders the root help with colors (ClaudeKit CLI style)
func customHelpFunc(cmd *cobra.Command, args []string) {
	// Logo (only for root command)
	if cmd.Parent() == nil {
		fmt.Println(Logo())
		fmt.Println()
	}

	// Description
	desc := cmd.Long
	if desc == "" {
		desc = cmd.Short
	}
	fmt.Println(strings.TrimSpace(desc))
	fmt.Println()

	// Usage section
	printSectionHeader("USAGE")
	fmt.Printf("  %s\n", cmd.UseLine())
	fmt.Println()

	// Command groups (if has subcommands)
	if cmd.HasAvailableSubCommands() {
		groups := groupCommands(cmd.Commands())
		for _, group := range groups {
			printSectionHeader(group.Title)
			for _, c := range group.Commands {
				printCommandRow(c.Name(), c.Short)
			}
			fmt.Println()
		}
	}

	// Flags section
	if cmd.HasAvailableLocalFlags() {
		printSectionHeader("FLAGS")
		printFlags(cmd.LocalFlags().FlagUsages())
		fmt.Println()
	}

	// Learn more
	printSectionHeader("LEARN MORE")
	fmt.Printf("  Use '%s %s' for more information\n",
		Command(cmd.CommandPath()+" <command>"),
		Placeholder("--help"))
}

// subcommandHelpFunc renders help for subcommands
func subcommandHelpFunc(cmd *cobra.Command, args []string) {
	// Description
	desc := cmd.Long
	if desc == "" {
		desc = cmd.Short
	}
	fmt.Println(strings.TrimSpace(desc))
	fmt.Println()

	// Usage section
	printSectionHeader("USAGE")
	fmt.Printf("  %s\n", cmd.UseLine())

	// Flags section
	if cmd.HasAvailableFlags() {
		fmt.Println()
		printSectionHeader("FLAGS")
		printFlags(cmd.Flags().FlagUsages())
	}
}

// customUsageFunc renders usage with colors
func customUsageFunc(cmd *cobra.Command) error {
	printSectionHeader("USAGE")
	fmt.Printf("  %s\n", cmd.UseLine())

	if cmd.HasAvailableSubCommands() {
		fmt.Println()
		fmt.Printf("  Use '%s %s' for more information about a command.\n",
			Command(cmd.CommandPath()+" <command>"),
			Placeholder("--help"))
	}
	return nil
}

// printSectionHeader prints a section header in white/bold
func printSectionHeader(title string) {
	pterm.Println(pterm.White(title))
}

// printCommandRow prints a command with description (two-column layout)
func printCommandRow(name, description string) {
	// Command name in green, 12-char padding, description in default
	fmt.Printf("  %s  %s\n",
		pterm.LightGreen(fmt.Sprintf("%-12s", name)),
		description)
}

// printFlags prints flag usages with colored flag names
func printFlags(flagUsages string) {
	lines := strings.Split(strings.TrimSpace(flagUsages), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		// Color the flag portion (before the description)
		// Format: "  -f, --flag   description"
		parts := strings.SplitN(strings.TrimSpace(line), "   ", 2)
		if len(parts) == 2 {
			// Flag with description
			fmt.Printf("  %s   %s\n",
				pterm.LightGreen(parts[0]),
				strings.TrimSpace(parts[1]))
		} else {
			// Just flag or continuation
			colored := colorFlagLine(line)
			fmt.Println(colored)
		}
	}
}

// colorFlagLine colors flags in a line
func colorFlagLine(line string) string {
	// Simple approach: color anything starting with - as green
	words := strings.Fields(line)
	for i, word := range words {
		if strings.HasPrefix(word, "-") {
			words[i] = pterm.LightGreen(word)
		}
	}
	return "  " + strings.Join(words, " ")
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

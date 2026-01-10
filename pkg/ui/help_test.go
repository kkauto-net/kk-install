package ui

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestGroupCommands_EmptyCommands(t *testing.T) {
	result := groupCommands([]*cobra.Command{})
	assert.Empty(t, result)
}

func TestGroupCommands_DefaultsToAdditional(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	cmd := &cobra.Command{Use: "test", Short: "Test command", Run: func(cmd *cobra.Command, args []string) {}}
	root.AddCommand(cmd)
	result := groupCommands(root.Commands())

	assert.Len(t, result, 1)
	assert.Equal(t, "ADDITIONAL COMMANDS", result[0].Title)
	assert.Len(t, result[0].Commands, 1)
}

func TestGroupCommands_CoreGroup(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	cmd := &cobra.Command{
		Use:         "init",
		Short:       "Initialize",
		Annotations: map[string]string{"group": "core"},
		Run:         func(cmd *cobra.Command, args []string) {},
	}
	root.AddCommand(cmd)
	result := groupCommands(root.Commands())

	assert.Len(t, result, 1)
	assert.Equal(t, "CORE COMMANDS", result[0].Title)
}

func TestGroupCommands_ManagementGroup(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	cmd := &cobra.Command{
		Use:         "restart",
		Short:       "Restart",
		Annotations: map[string]string{"group": "management"},
		Run:         func(cmd *cobra.Command, args []string) {},
	}
	root.AddCommand(cmd)
	result := groupCommands(root.Commands())

	assert.Len(t, result, 1)
	assert.Equal(t, "MANAGEMENT COMMANDS", result[0].Title)
}

func TestGroupCommands_MultipleGroups(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	run := func(cmd *cobra.Command, args []string) {}
	commands := []*cobra.Command{
		{Use: "init", Short: "Init", Annotations: map[string]string{"group": "core"}, Run: run},
		{Use: "restart", Short: "Restart", Annotations: map[string]string{"group": "management"}, Run: run},
		{Use: "completion", Short: "Completion", Annotations: map[string]string{"group": "additional"}, Run: run},
	}
	for _, cmd := range commands {
		root.AddCommand(cmd)
	}
	result := groupCommands(root.Commands())

	// Should have 3 groups in order: core, management, additional
	assert.Len(t, result, 3)
	assert.Equal(t, "CORE COMMANDS", result[0].Title)
	assert.Equal(t, "MANAGEMENT COMMANDS", result[1].Title)
	assert.Equal(t, "ADDITIONAL COMMANDS", result[2].Title)
}

func TestGroupCommands_SkipsHiddenCommands(t *testing.T) {
	root := &cobra.Command{Use: "root"}
	cmd := &cobra.Command{
		Use:    "hidden",
		Short:  "Hidden command",
		Hidden: true,
	}
	root.AddCommand(cmd)
	result := groupCommands(root.Commands())

	assert.Empty(t, result)
}

func TestApplyTemplates(t *testing.T) {
	rootCmd := &cobra.Command{Use: "test"}
	subCmd := &cobra.Command{Use: "sub", Short: "Sub command"}
	rootCmd.AddCommand(subCmd)

	// Should not panic
	ApplyTemplates(rootCmd)

	// Verify templates were set (check that help template contains our custom format)
	assert.NotEmpty(t, rootCmd.HelpTemplate())
	assert.Contains(t, rootCmd.HelpTemplate(), "USAGE")
	assert.Contains(t, rootCmd.HelpTemplate(), "LEARN MORE")
}

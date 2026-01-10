---
phase: 02
title: Custom Cobra Help Templates
status: pending
effort: 45m
---

# Phase 02: Custom Help Templates

## Context

- Parent plan: [plan.md](plan.md)
- Dependencies: Phase 01 (for language loading)

## Overview

Create GitHub CLI style help templates using Cobra's `SetUsageTemplate` and `SetHelpTemplate`.

## Requirements

1. UPPERCASE section headers (USAGE, CORE COMMANDS, FLAGS, etc.)
2. Grouped commands by annotation
3. Colon separator between command and description
4. LEARN MORE section at bottom

## Target Output

```
🚀 Manage your kkengine Docker stack effortlessly.

USAGE
  kk <command> [flags]

CORE COMMANDS
  init:       Initialize Docker stack with interactive setup
  start:      Start all services with preflight checks
  status:     View service status and health

MANAGEMENT COMMANDS
  restart:    Restart all services
  update:     Pull latest images and recreate containers

ADDITIONAL COMMANDS
  completion: Generate shell completion scripts

FLAGS
  -h, --help      Show help for command
  -v, --version   Show version

LEARN MORE
  Use 'kk <command> --help' for more information
```

## Implementation Steps

### 1. Create pkg/ui/help.go

```go
package ui

import (
    "bytes"
    "fmt"
    "strings"
    "text/template"

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
{{$group.Title | upper}}{{range $group.Commands}}
  {{rpad .Name 12}}{{.Short}}{{end}}
{{end}}{{end}}{{if .HasAvailableLocalFlags}}
FLAGS
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}
{{end}}
LEARN MORE
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
    cobra.AddTemplateFunc("upper", strings.ToUpper)
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
```

### 2. Update cmd/root.go

```go
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
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}

func init() {
    rootCmd.Version = Version

    // Load language preference
    cfg, _ := config.Load()
    ui.SetLanguage(ui.Language(cfg.Language))

    // Apply custom templates
    ui.ApplyTemplates(rootCmd)
}
```

## Todo List

- [ ] Create pkg/ui/help.go with templates
- [ ] Add template functions (upper, trim, groupCommands)
- [ ] Update cmd/root.go to apply templates
- [ ] Test help output format

## Success Criteria

- [ ] `kk --help` shows GitHub CLI style output
- [ ] Commands grouped by annotation
- [ ] UPPERCASE headers display correctly
- [ ] Subcommand help uses simplified template

## Files Changed

| File | Action |
|------|--------|
| `pkg/ui/help.go` | CREATE |
| `cmd/root.go` | MODIFY |

package ui

import (
	"github.com/kkauto-net/kk-install/pkg/config"
	"github.com/pterm/pterm"
)

// PrintConfigSummary displays CLI configuration in a boxed table.
func PrintConfigSummary(cfg *config.Config) {
	langDisplay := Msg("lang_english")
	if cfg.Language == "vi" {
		langDisplay = Msg("lang_vietnamese")
	}

	projectDir := cfg.ProjectDir
	if projectDir == "" {
		projectDir = Msg("config_not_set")
	}

	tableData := pterm.TableData{
		{Msg("col_setting"), Msg("col_value")},
		{Msg("config_language"), langDisplay},
		{Msg("config_project_dir"), projectDir},
		{Msg("config_file_path"), config.ConfigPath()},
	}

	pterm.DefaultSection.Println(Msg("config_title"))
	renderTable(pterm.DefaultTable.
		WithHasHeader(true).
		WithBoxed(true).
		WithData(tableData))
}

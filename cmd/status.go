package cmd

import (
	"context"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/kkauto-net/kk-install/pkg/compose"
	"github.com/kkauto-net/kk-install/pkg/config"
	"github.com/kkauto-net/kk-install/pkg/monitor"
	"github.com/kkauto-net/kk-install/pkg/ui"
)

var statusCmd = &cobra.Command{
	Use:         "status",
	Short:       "View service status and health",
	Long:        `Display status of all containers in the stack.`,
	Annotations: map[string]string{"group": "core"},
	RunE:        runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	cwd, err := config.EnsureProjectDir()
	if err != nil {
		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:      ui.Msg("project_not_configured"),
			Message:    ui.SanitizeError(err),
			Suggestion: ui.Msg("run_init_to_configure"),
			Command:    "kk init",
		})
		return err
	}

	ui.ShowCommandBanner(ui.Msg("cmd_status_title"), ui.Msg("status_desc"))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	composeFile, err := compose.ParseComposeFile(cwd)
	if err != nil {
		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:      ui.Msg("get_status_failed"),
			Message:    ui.SanitizeError(err),
			Suggestion: ui.Msg("err_compose_file_missing"),
			Command:    "kk init",
		})
		return err
	}

	definedServices := composeFile.GetServiceNames()
	if len(definedServices) == 0 {
		ui.ShowWarning(ui.Msg("no_services_defined"))
		ui.ShowNote(ui.Msg("run_init"))
		return nil
	}

	executor := compose.NewExecutor(cwd)
	spinner := ui.StartPtermSpinner(ui.Msg("get_status_failed"))
	statuses, err := monitor.GetStatusWithServices(ctx, executor, definedServices)
	if err != nil {
		spinner.Fail(ui.Msg("get_status_failed"))

		suggestion := ui.Msg("err_check_docker_running")
		command := ui.Msg("docker_start_command")

		if ui.IsDockerPermissionError(err) {
			suggestion, command = ui.DockerPermissionSuggestion()
		}

		ui.ShowBoxedError(ui.ErrorSuggestion{
			Title:      ui.Msg("get_status_failed"),
			Message:    ui.SanitizeError(err),
			Suggestion: suggestion,
			Command:    command,
		})
		return err
	}
	spinner.Success(ui.Msg("status_desc"))

	ui.PrintStatusTable(statuses)

	var unhealthyServices []string
	for _, s := range statuses {
		if s.Running && s.Health == "unhealthy" {
			unhealthyServices = append(unhealthyServices, s.Name)
		}
	}
	if len(unhealthyServices) > 0 {
		ui.ShowWarning(ui.MsgF("unhealthy_services_hint", len(unhealthyServices), strings.Join(unhealthyServices, ", ")))
		ui.ShowNote(ui.MsgF("unhealthy_services_logs", unhealthyServices[0]))
	}

	for _, s := range statuses {
		if s.Running {
			domain := config.ReadEnvValue(cwd, "SYSTEM_DOMAIN")
			ui.PrintAccessInfo(statuses, domain)
			break
		}
	}

	return nil
}

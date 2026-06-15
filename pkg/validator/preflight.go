package validator

import (
	"fmt"

	"github.com/kkauto-net/kk-install/pkg/ui"
	"github.com/pterm/pterm"
)

type PreflightResult struct {
	CheckName  string
	Passed     bool
	Error      error
	Warning    string
	Fix        string // Fix suggestion for the error
	FixCommand string // Command to run to fix the error
}

// RunPreflight executes all validation checks
func RunPreflight(dir string, includeCaddy bool) ([]PreflightResult, error) {
	var results []PreflightResult
	var hasBlockingError bool

	// Create docker validator instance
	dockerValidator := NewDockerValidator()

	// 1. Docker installed
	err := dockerValidator.CheckDockerInstalled()
	results = append(results, PreflightResult{
		CheckName:  ui.Msg("preflight_check_docker_installed"),
		Passed:     err == nil,
		Error:      err,
		Fix:        ui.Msg("preflight_fix_install_docker"),
		FixCommand: "https://docs.docker.com/get-docker/",
	})
	if err != nil {
		hasBlockingError = true
	}

	// 2. Docker daemon running (only if installed)
	if !hasBlockingError {
		err = dockerValidator.CheckDockerDaemon()
		results = append(results, PreflightResult{
			CheckName:  ui.Msg("preflight_check_docker_daemon"),
			Passed:     err == nil,
			Error:      err,
			Fix:        ui.Msg("preflight_fix_start_docker"),
			FixCommand: "systemctl start docker",
		})
		if err != nil {
			hasBlockingError = true
		}
	}

	// 3. Port conflicts
	_, err = CheckAllPorts(includeCaddy)
	results = append(results, PreflightResult{
		CheckName:  ui.Msg("preflight_check_ports"),
		Passed:     err == nil,
		Error:      err,
		Fix:        ui.Msg("preflight_fix_stop_conflicting"),
		FixCommand: "",
	})
	if err != nil {
		hasBlockingError = true
	}

	// 4. Environment file
	err = ValidateEnvFile(dir)
	results = append(results, PreflightResult{
		CheckName:  ui.Msg("preflight_check_env"),
		Passed:     err == nil,
		Error:      err,
		Fix:        ui.Msg("preflight_fix_create_env"),
		FixCommand: "kk init",
	})
	if err != nil {
		hasBlockingError = true
	}

	// 5. Docker compose syntax
	err = ValidateDockerCompose(dir)
	results = append(results, PreflightResult{
		CheckName:  ui.Msg("preflight_check_compose"),
		Passed:     err == nil,
		Error:      err,
		Fix:        ui.Msg("preflight_fix_create_compose"),
		FixCommand: "kk init",
	})
	if err != nil {
		hasBlockingError = true
	}

	// 6. Caddyfile (if enabled)
	if includeCaddy {
		err = ValidateCaddyfile(dir)
		results = append(results, PreflightResult{
			CheckName:  ui.Msg("preflight_check_caddyfile"),
			Passed:     err == nil,
			Error:      err,
			Fix:        ui.Msg("preflight_fix_create_caddyfile"),
			FixCommand: "kk init",
		})
		if err != nil {
			hasBlockingError = true
		}
	}

	// 7. Disk space (warning only)
	availableGB, err := CheckDiskSpace(dir)
	if err == nil && availableGB < MinDiskSpaceGB {
		results = append(results, PreflightResult{
			CheckName: ui.Msg("preflight_check_disk"),
			Passed:    true, // Warning only
			Warning:   ui.MsgF("preflight_disk_warning", availableGB, MinDiskSpaceGB),
		})
	} else {
		results = append(results, PreflightResult{
			CheckName: ui.Msg("preflight_check_disk"),
			Passed:    true,
		})
	}

	// Return error if any blocking check failed
	if hasBlockingError {
		return results, fmt.Errorf("preflight checks failed")
	}

	return results, nil
}

// PrintPreflightResults displays preflight check results as pterm table
func PrintPreflightResults(results []PreflightResult) {
	tableData := pterm.TableData{
		{ui.Msg("check"), ui.Msg("result")},
	}

	for _, r := range results {
		var status string
		if r.Passed {
			if r.Warning != "" {
				status = pterm.Yellow("⚠ " + r.Warning)
			} else {
				status = pterm.Green("✓ " + ui.Msg("preflight_pass"))
			}
		} else {
			if r.Error != nil {
				errMsg := TranslateError(r.Error)
				status = pterm.Red("✗ " + errMsg)
				if r.Fix != "" {
					status += "\n  → " + r.Fix
				}
				if r.FixCommand != "" {
					status += ": " + r.FixCommand
				}
			} else {
				status = pterm.Red("✗ " + ui.Msg("preflight_result_failed"))
			}
		}
		tableData = append(tableData, []string{r.CheckName, status})
	}

	if err := pterm.DefaultTable.
		WithHasHeader(true).
		WithBoxed(true).
		WithData(tableData).
		Render(); err != nil {
		ui.ShowWarningf(ui.Msg("preflight_table_render_failed"), err)
	}
}

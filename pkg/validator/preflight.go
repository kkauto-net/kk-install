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
		CheckName:  "Docker cai dat",
		Passed:     err == nil,
		Error:      err,
		Fix:        "Install Docker",
		FixCommand: "https://docs.docker.com/get-docker/",
	})
	if err != nil {
		hasBlockingError = true
	}

	// 2. Docker daemon running (only if installed)
	if !hasBlockingError {
		err = dockerValidator.CheckDockerDaemon()
		results = append(results, PreflightResult{
			CheckName:  "Docker daemon",
			Passed:     err == nil,
			Error:      err,
			Fix:        "Start Docker daemon",
			FixCommand: "systemctl start docker",
		})
		if err != nil {
			hasBlockingError = true
		}
	}

	// 3. Port conflicts
	_, err = CheckAllPorts(includeCaddy)
	results = append(results, PreflightResult{
		CheckName:  "Cong mang (ports)",
		Passed:     err == nil,
		Error:      err,
		Fix:        "Stop conflicting services or change ports",
		FixCommand: "",
	})
	if err != nil {
		hasBlockingError = true
	}

	// 4. Environment file
	err = ValidateEnvFile(dir)
	results = append(results, PreflightResult{
		CheckName:  "File .env",
		Passed:     err == nil,
		Error:      err,
		Fix:        "Create .env file",
		FixCommand: "kk init",
	})
	if err != nil {
		hasBlockingError = true
	}

	// 5. Docker compose syntax
	err = ValidateDockerCompose(dir)
	results = append(results, PreflightResult{
		CheckName:  "docker-compose.yml",
		Passed:     err == nil,
		Error:      err,
		Fix:        "Create or fix docker-compose.yml",
		FixCommand: "kk init",
	})
	if err != nil {
		hasBlockingError = true
	}

	// 6. Caddyfile (if enabled)
	if includeCaddy {
		err = ValidateCaddyfile(dir)
		results = append(results, PreflightResult{
			CheckName:  "Caddyfile",
			Passed:     err == nil,
			Error:      err,
			Fix:        "Create or fix Caddyfile",
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
			CheckName: "Disk space",
			Passed:    true, // Warning only
			Warning:   fmt.Sprintf("Chi con %.1fGB, recommend >= %dGB", availableGB, MinDiskSpaceGB),
		})
	} else {
		results = append(results, PreflightResult{
			CheckName: "Disk space",
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
				status = pterm.Green("✓ Pass")
			}
		} else {
			if r.Error != nil {
				errMsg := TranslateError(r.Error)
				status = pterm.Red("✗ " + errMsg)
				// Add fix suggestion on new line if available
				if r.Fix != "" {
					status += "\n  → " + r.Fix
				}
				if r.FixCommand != "" {
					status += ": " + r.FixCommand
				}
			} else {
				status = pterm.Red("✗ Failed")
			}
		}
		tableData = append(tableData, []string{r.CheckName, status})
	}

	pterm.DefaultTable.
		WithHasHeader(true).
		WithBoxed(true).
		WithData(tableData).
		Render()
}

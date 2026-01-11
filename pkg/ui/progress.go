package ui

import (
	"fmt"
	"sync"
	"time"

	"github.com/pterm/pterm"
)

// SimpleSpinner provides basic spinner animation for progress indication.
// Deprecated: Use StartPtermSpinner for better terminal support.
type SimpleSpinner struct {
	frames  []string
	current int
	message string
	mu      sync.RWMutex // Protects message field
	done    chan bool
}

// NewSpinner creates a new SimpleSpinner with the given message.
// Deprecated: Use StartPtermSpinner for better terminal support.
func NewSpinner(message string) *SimpleSpinner {
	return &SimpleSpinner{
		frames:  []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		message: message,
		done:    make(chan bool, 1), // Buffered to prevent deadlock
	}
}

// Start begins the spinner animation in a goroutine.
func (s *SimpleSpinner) Start() {
	go func() {
		for {
			select {
			case <-s.done:
				return
			default:
				s.mu.RLock()
				msg := s.message
				s.mu.RUnlock()
				fmt.Printf("\r  %s %s ", s.frames[s.current], msg)
				s.current = (s.current + 1) % len(s.frames)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
}

// Stop halts the spinner and shows final status.
// If success is true, shows [OK]; otherwise shows [X].
func (s *SimpleSpinner) Stop(success bool) {
	s.done <- true
	s.mu.RLock()
	msg := s.message
	s.mu.RUnlock()
	if success {
		fmt.Printf("\r  [OK] %s\n", msg)
	} else {
		fmt.Printf("\r  [X] %s\n", msg)
	}
}

// UpdateMessage changes the spinner message while it's running.
func (s *SimpleSpinner) UpdateMessage(msg string) {
	s.mu.Lock()
	s.message = msg
	s.mu.Unlock()
}

// StartPtermSpinner creates and starts a pterm spinner with the given message.
// Returns a SpinnerPrinter that can be controlled with Success(), Fail(), etc.
func StartPtermSpinner(msg string) *pterm.SpinnerPrinter {
	spinner, _ := pterm.DefaultSpinner.Start(msg)
	return spinner
}

// ShowServiceProgress displays service startup status using pterm formatting.
// Status can be: "starting", "healthy", "running", "unhealthy", or any custom value.
func ShowServiceProgress(serviceName, status string) {
	switch status {
	case "starting":
		pterm.Info.Printfln("%s %s %s", IconStarting, serviceName, Msg("starting"))
	case "healthy", "running":
		pterm.Success.Printfln("%s %s %s", IconHealthy, serviceName, Msg("ready"))
	case "unhealthy":
		pterm.Error.Printfln("%s %s %s", IconUnhealthy, serviceName, Msg("unhealthy"))
	default:
		pterm.Warning.Printfln("%s %s: %s", IconWarning, serviceName, status)
	}
}

// ShowStepHeader displays a step progress indicator (e.g., "Step 1/4: Title").
func ShowStepHeader(current, total int, title string) {
	stepText := fmt.Sprintf("Step %d/%d", current, total)
	pterm.DefaultSection.
		WithLevel(2).
		Println(fmt.Sprintf("%s: %s", stepText, title))
}

// PrintInitSummary shows configuration summary and created files after kk init.
func PrintInitSummary(enableSeaweedFS, enableCaddy bool, domain string, createdFiles []string, installDir string) {
	// 1. Configuration Summary
	pterm.DefaultSection.Println(Msg("config_summary"))

	configData := pterm.TableData{
		{Msg("col_setting"), Msg("col_value")},
		{"SeaweedFS", boolToStatus(enableSeaweedFS)},
		{"Caddy", boolToStatus(enableCaddy)},
	}
	if enableCaddy && domain != "" {
		configData = append(configData, []string{Msg("domain"), domain})
	}

	pterm.DefaultTable.
		WithHasHeader(true).
		WithBoxed(true).
		WithData(configData).
		Render()

	// 2. Created Files
	fmt.Println()
	pterm.DefaultSection.Println(Msg("created_files"))

	fileData := pterm.TableData{{Msg("col_file")}}
	for _, f := range createdFiles {
		fileData = append(fileData, []string{pterm.Green("✓ " + f)})
	}

	pterm.DefaultTable.
		WithHasHeader(true).
		WithBoxed(true).
		WithData(fileData).
		Render()

	// 3. Installation Location
	fmt.Println()
	pterm.DefaultSection.Println(Msg("install_location"))
	pterm.DefaultTable.
		WithBoxed(true).
		WithData(pterm.TableData{
			{IconFolder + " " + installDir},
		}).
		Render()

	// 4. Data Directories
	fmt.Println()
	pterm.DefaultSection.Println(Msg("data_directories"))
	pterm.DefaultTable.
		WithBoxed(true).
		WithData(pterm.TableData{
			{IconStorage + " SYSTEM_DATABASE", "./data_database"},
			{IconStorage + " SYSTEM_FILESTORE", "./data_storage"},
		}).
		Render()
}

// boolToStatus returns colored enabled/disabled status
func boolToStatus(b bool) string {
	if b {
		return pterm.Green("✓ " + Msg("enabled"))
	}
	return pterm.Gray("○ " + Msg("disabled"))
}

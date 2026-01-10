package ui

import (
	"fmt"
	"sync"
	"time"

	"github.com/pterm/pterm"
)

// SimpleSpinner provides basic spinner animation
type SimpleSpinner struct {
	frames  []string
	current int
	message string
	mu      sync.RWMutex // Protects message field
	done    chan bool
}

func NewSpinner(message string) *SimpleSpinner {
	return &SimpleSpinner{
		frames:  []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		message: message,
		done:    make(chan bool, 1), // Buffered to prevent deadlock
	}
}

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

func (s *SimpleSpinner) UpdateMessage(msg string) {
	s.mu.Lock()
	s.message = msg
	s.mu.Unlock()
}

// ProgressIndicator for service startup
func ShowServiceProgress(serviceName, status string) {
	switch status {
	case "starting":
		fmt.Printf("  [>] %s khoi dong...\n", serviceName)
	case "healthy", "running":
		fmt.Printf("  [OK] %s san sang\n", serviceName)
	case "unhealthy":
		fmt.Printf("  [X] %s khong khoe manh\n", serviceName)
	default:
		fmt.Printf("  [?] %s: %s\n", serviceName, status)
	}
}

// ShowStepHeader displays step progress indicator
func ShowStepHeader(current, total int, title string) {
	stepText := fmt.Sprintf("Step %d/%d", current, total)
	pterm.DefaultSection.
		WithLevel(2).
		Println(fmt.Sprintf("%s: %s", stepText, title))
}

// PrintInitSummary shows configuration summary and created files
func PrintInitSummary(enableSeaweedFS, enableCaddy bool, domain string, createdFiles []string) {
	// Configuration Summary
	pterm.DefaultSection.Println(Msg("config_summary"))

	configData := pterm.TableData{
		{Msg("col_setting"), Msg("col_value")},
		{"SeaweedFS", boolToStatus(enableSeaweedFS)},
		{"Caddy", boolToStatus(enableCaddy)},
	}
	if enableCaddy && domain != "" {
		configData = append(configData, []string{Msg("domain"), domain})
	}

	pterm.DefaultTable.WithHasHeader(true).WithData(configData).Render()

	// Created Files
	fmt.Println()
	pterm.DefaultSection.Println(Msg("created_files"))

	for _, f := range createdFiles {
		pterm.Success.Println(f)
	}
}

// boolToStatus returns colored enabled/disabled status
func boolToStatus(b bool) string {
	if b {
		return pterm.Green("✓ " + Msg("enabled"))
	}
	return pterm.Gray("○ " + Msg("disabled"))
}

package ui

import (
	"fmt"
	"time"
)

// SimpleSpinner provides basic spinner animation
type SimpleSpinner struct {
	frames  []string
	current int
	message string
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
				fmt.Printf("\r  %s %s ", s.frames[s.current], s.message)
				s.current = (s.current + 1) % len(s.frames)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
}

func (s *SimpleSpinner) Stop(success bool) {
	s.done <- true
	if success {
		fmt.Printf("\r  [OK] %s\n", s.message)
	} else {
		fmt.Printf("\r  [X] %s\n", s.message)
	}
}

func (s *SimpleSpinner) UpdateMessage(msg string) {
	s.message = msg
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

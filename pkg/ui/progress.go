package ui

import (
	"fmt"
	"sync"
	"time"
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

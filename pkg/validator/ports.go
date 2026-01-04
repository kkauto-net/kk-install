package validator

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type PortStatus struct {
	Port    int
	InUse   bool
	PID     int
	Process string
}

// RequiredPorts defines ports needed by kkengine stack
var RequiredPorts = map[string]int{
	"MariaDB":  3307,
	"kkengine": 8019,
}

var OptionalPorts = map[string]int{
	"Caddy HTTP":  80,
	"Caddy HTTPS": 443,
}

// CheckPort uses net.Listen to check if port is available
func CheckPort(port int) PortStatus {
	status := PortStatus{Port: port}

	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		status.InUse = true
		// Try to find which process is using it
		pid, process := findProcessUsingPort(port)
		status.PID = pid
		status.Process = process
		return status
	}
	listener.Close()
	return status
}

// CheckAllPorts validates all required ports
func CheckAllPorts(includeCaddy bool) ([]PortStatus, error) {
	var results []PortStatus
	var conflicts []string

	// Check required ports
	for name, port := range RequiredPorts {
		status := CheckPort(port)
		results = append(results, status)
		if status.InUse {
			conflicts = append(conflicts, formatPortConflict(name, status))
		}
	}

	// Check optional Caddy ports if enabled
	if includeCaddy {
		for name, port := range OptionalPorts {
			status := CheckPort(port)
			results = append(results, status)
			if status.InUse {
				conflicts = append(conflicts, formatPortConflict(name, status))
			}
		}
	}

	if len(conflicts) > 0 {
		return results, &UserError{
			Key:        "port_conflict",
			Message:    "Xung dot port",
			Suggestion: strings.Join(conflicts, "\n"),
		}
	}
	return results, nil
}

// findProcessUsingPort attempts to find PID using the port (Linux)
func findProcessUsingPort(port int) (int, string) {
	// Try /proc/net/tcp first (Linux-specific, no external command)
	pid, process := findFromProcNet(port)
	if pid > 0 {
		return pid, process
	}

	// Fallback to lsof (works on most Unix systems)
	return findFromLsof(port)
}

func findFromProcNet(port int) (int, string) {
	// /proc/net/tcp uses hex port numbers
	hexPort := fmt.Sprintf(":%04X", port)

	file, err := os.Open("/proc/net/tcp")
	if err != nil {
		return 0, ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, hexPort) {
			// Extract inode, then find PID from /proc/*/fd
			// Simplified: return 0 and let lsof handle it
			return 0, ""
		}
	}
	return 0, ""
}

func findFromLsof(port int) (int, string) {
	cmd := exec.Command("lsof", "-i", fmt.Sprintf(":%d", port), "-t", "-sTCP:LISTEN")
	output, err := cmd.Output()
	if err != nil {
		return 0, ""
	}

	pidStr := strings.TrimSpace(string(output))
	if pidStr == "" {
		return 0, ""
	}

	// Get first PID if multiple
	pids := strings.Split(pidStr, "\n")
	pid, err := strconv.Atoi(pids[0])
	if err != nil {
		return 0, ""
	}

	// Get process name from /proc/PID/comm
	commPath := fmt.Sprintf("/proc/%d/comm", pid)
	comm, err := os.ReadFile(commPath)
	if err != nil {
		return pid, ""
	}

	return pid, strings.TrimSpace(string(comm))
}

func formatPortConflict(name string, status PortStatus) string {
	if status.PID > 0 {
		if status.Process != "" {
			return fmt.Sprintf("  - Port %d (%s): dang dung boi %s (PID %d). Stop: sudo kill %d",
				status.Port, name, status.Process, status.PID, status.PID)
		}
		return fmt.Sprintf("  - Port %d (%s): dang dung boi PID %d. Stop: sudo kill %d",
			status.Port, name, status.PID, status.PID)
	}
	return fmt.Sprintf("  - Port %d (%s): dang duoc su dung. Kiem tra: sudo lsof -i :%d",
		status.Port, name, status.Port)
}

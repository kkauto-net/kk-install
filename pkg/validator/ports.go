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
	Port           int
	InUse          bool
	PID            int
	Process        string
	UsedByKKEngine bool // true if port is used by kkengine container
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
// For privileged ports (<1024), uses alternative methods if not root
func CheckPort(port int) PortStatus {
	status := PortStatus{Port: port}

	// For privileged ports, check if we're root first
	if port < 1024 && os.Getuid() != 0 {
		// Not root - use ss/netstat to check instead of net.Listen
		inUse, pid, process := checkPortWithSS(port)
		if inUse {
			status.InUse = true
			status.PID = pid
			status.Process = process
			status.UsedByKKEngine = isPortUsedByKKEngine(port)
		}
		return status
	}

	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		status.InUse = true
		// Try to find which process is using it
		pid, process := findProcessUsingPort(port)
		status.PID = pid
		status.Process = process
		status.UsedByKKEngine = isPortUsedByKKEngine(port)
		return status
	}
	listener.Close()
	return status
}

// isPortUsedByKKEngine checks if port is being used by a kkengine container
func isPortUsedByKKEngine(port int) bool {
	// Use docker ps to check if any kkengine_ container is using this port
	cmd := exec.Command("docker", "ps", "--filter", "name=kkengine_", "--format", "{{.Ports}}")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	portStr := fmt.Sprintf(":%d->", port)
	return strings.Contains(string(output), portStr)
}

// checkPortWithSS uses ss command to check if port is in use (works without root)
func checkPortWithSS(port int) (inUse bool, pid int, process string) {
	// Try ss first (more common on modern Linux)
	cmd := exec.Command("ss", "-tlnp", fmt.Sprintf("sport = :%d", port))
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		// First line is header, check if there's actual data
		for _, line := range lines[1:] {
			if strings.TrimSpace(line) != "" && strings.Contains(line, fmt.Sprintf(":%d", port)) {
				// Port is in use
				return true, 0, ""
			}
		}
		return false, 0, ""
	}

	// Fallback to netstat
	cmd = exec.Command("netstat", "-tlnp")
	output, err = cmd.Output()
	if err == nil {
		portStr := fmt.Sprintf(":%d", port)
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, portStr) && strings.Contains(line, "LISTEN") {
				return true, 0, ""
			}
		}
		return false, 0, ""
	}

	// Can't determine - assume available
	return false, 0, ""
}

// CheckAllPorts validates all required ports
func CheckAllPorts(includeCaddy bool) ([]PortStatus, error) {
	var results []PortStatus
	var conflicts []string

	// Check required ports
	for name, port := range RequiredPorts {
		status := CheckPort(port)
		results = append(results, status)
		// Only report conflict if port is NOT used by our own kkengine containers
		if status.InUse && !status.UsedByKKEngine {
			conflicts = append(conflicts, formatPortConflict(name, status))
		}
	}

	// Check optional Caddy ports if enabled
	if includeCaddy {
		for name, port := range OptionalPorts {
			status := CheckPort(port)
			results = append(results, status)
			// Only report conflict if port is NOT used by our own kkengine containers
			if status.InUse && !status.UsedByKKEngine {
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

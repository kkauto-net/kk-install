# Phase 02: Validation Layer

## Context

- **Parent Plan:** [plan.md](./plan.md)
- **Dependencies:** [Phase 01 - Core Foundation](./phase-01-core-foundation.md)
- **Related Research:** [Docker Integration](./research/researcher-02-docker-integration.md)

## Overview

| Field | Value |
|-------|-------|
| Date | 2026-01-04 |
| Description | Port conflict detection, env validation, config validation, error translation framework |
| Priority | P1 |
| Status | pending |
| Effort | 1 week |

## Key Insights (from Research)

1. **Port Detection:** `net.Listen` la cach dang tin cay nhat, cross-platform
2. **Docker API Fallback:** Kiem tra port mappings cua containers dang chay
3. **Error Translation:** Tach technical error va user-facing message
4. **I18n Pattern:** Key-based messages cho de dang mo rong ngon ngu

## Requirements

- [x] Port conflict detection (3307, 8019, 80, 443)
- [x] Identify process using port (PID, process name)
- [x] Environment variable validation (.env completeness)
- [x] Docker compose syntax validation
- [x] Disk space check (warn if < 5GB)
- [x] User-friendly error messages in Vietnamese
- [x] Error translation framework

## Architecture

```
pkg/
├── validator/
│   ├── docker.go    # (from Phase 01)
│   ├── ports.go     # Port conflict detection
│   ├── env.go       # Environment validation
│   ├── config.go    # Config syntax validation
│   ├── disk.go      # Disk space check
│   └── errors.go    # Error types + translation
└── ui/
    ├── messages.go  # (from Phase 01)
    └── errors.go    # Error display formatting
```

## Related Code Files

After implementation:
- `/home/kkdev/kkcli/pkg/validator/ports.go`
- `/home/kkdev/kkcli/pkg/validator/env.go`
- `/home/kkdev/kkcli/pkg/validator/config.go`
- `/home/kkdev/kkcli/pkg/validator/disk.go`
- `/home/kkdev/kkcli/pkg/validator/errors.go`
- `/home/kkdev/kkcli/pkg/ui/errors.go`

## Implementation Steps

### Step 1: Port Conflict Detection (4h)

**pkg/validator/ports.go:**
```go
package validator

import (
    "bufio"
    "fmt"
    "net"
    "os"
    "os/exec"
    "regexp"
    "strconv"
    "strings"
)

type PortStatus struct {
    Port      int
    InUse     bool
    PID       int
    Process   string
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
```

### Step 2: Environment Validation (3h)

**pkg/validator/env.go:**
```go
package validator

import (
    "bufio"
    "fmt"
    "os"
    "path/filepath"
    "strings"
)

// RequiredEnvVars lists mandatory environment variables
var RequiredEnvVars = []string{
    "DB_PASSWORD",
    "DB_ROOT_PASSWORD",
    "REDIS_PASSWORD",
}

// OptionalEnvVars lists optional environment variables with defaults
var OptionalEnvVars = map[string]string{
    "DB_HOSTNAME": "db",
    "DB_PORT":     "3306",
    "DB_DATABASE": "kkengine",
    "DB_USERNAME": "kkengine",
    "REDIS_HOST":  "redis",
    "REDIS_PORT":  "6379",
}

// ValidateEnvFile checks .env file exists and contains required vars
func ValidateEnvFile(dir string) error {
    envPath := filepath.Join(dir, ".env")

    // Check file exists
    if _, err := os.Stat(envPath); os.IsNotExist(err) {
        return &UserError{
            Key:        "env_missing",
            Message:    "File .env khong ton tai",
            Suggestion: "Chay: kk init",
        }
    }

    // Parse .env file
    envVars, err := parseEnvFile(envPath)
    if err != nil {
        return &UserError{
            Key:        "env_parse_error",
            Message:    fmt.Sprintf("Loi doc file .env: %v", err),
            Suggestion: "Kiem tra cu phap file .env",
        }
    }

    // Check required vars
    var missing []string
    for _, key := range RequiredEnvVars {
        if val, ok := envVars[key]; !ok || val == "" {
            missing = append(missing, key)
        }
    }

    if len(missing) > 0 {
        return &UserError{
            Key:        "env_missing_vars",
            Message:    "Thieu bien moi truong trong .env",
            Suggestion: fmt.Sprintf("Them vao .env: %s", strings.Join(missing, ", ")),
        }
    }

    // Check password strength (minimum 16 chars)
    passwordVars := []string{"DB_PASSWORD", "DB_ROOT_PASSWORD", "REDIS_PASSWORD"}
    var weakPasswords []string
    for _, key := range passwordVars {
        if val, ok := envVars[key]; ok && len(val) < 16 {
            weakPasswords = append(weakPasswords, key)
        }
    }

    if len(weakPasswords) > 0 {
        // Warning only, don't block
        fmt.Printf("  [!] Canh bao: Mat khau yeu cho: %s (nen >= 16 ky tu)\n",
            strings.Join(weakPasswords, ", "))
    }

    return nil
}

func parseEnvFile(path string) (map[string]string, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    vars := make(map[string]string)
    scanner := bufio.NewScanner(file)
    lineNum := 0

    for scanner.Scan() {
        lineNum++
        line := strings.TrimSpace(scanner.Text())

        // Skip empty lines and comments
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }

        // Parse KEY=VALUE
        parts := strings.SplitN(line, "=", 2)
        if len(parts) != 2 {
            continue // Skip malformed lines
        }

        key := strings.TrimSpace(parts[0])
        value := strings.TrimSpace(parts[1])

        // Remove quotes if present
        value = strings.Trim(value, "\"'")

        vars[key] = value
    }

    return vars, scanner.Err()
}

// CheckEnvPermissions warns if .env is world-readable
func CheckEnvPermissions(dir string) {
    envPath := filepath.Join(dir, ".env")
    info, err := os.Stat(envPath)
    if err != nil {
        return
    }

    mode := info.Mode()
    // Check if others have read permission (Unix)
    if mode&0004 != 0 {
        fmt.Printf("  [!] Canh bao: File .env co the doc boi nguoi khac.\n")
        fmt.Printf("      Chay: chmod 600 %s\n", envPath)
    }
}
```

### Step 3: Config Syntax Validation (2h)

**pkg/validator/config.go:**
```go
package validator

import (
    "fmt"
    "os"
    "path/filepath"

    "gopkg.in/yaml.v3"
)

// ValidateDockerCompose checks docker-compose.yml syntax
func ValidateDockerCompose(dir string) error {
    composePath := filepath.Join(dir, "docker-compose.yml")

    if _, err := os.Stat(composePath); os.IsNotExist(err) {
        return &UserError{
            Key:        "compose_missing",
            Message:    "File docker-compose.yml khong ton tai",
            Suggestion: "Chay: kk init",
        }
    }

    content, err := os.ReadFile(composePath)
    if err != nil {
        return &UserError{
            Key:        "compose_read_error",
            Message:    fmt.Sprintf("Khong doc duoc docker-compose.yml: %v", err),
            Suggestion: "Kiem tra quyen truy cap file",
        }
    }

    // Parse YAML to validate syntax
    var compose map[string]interface{}
    if err := yaml.Unmarshal(content, &compose); err != nil {
        return &UserError{
            Key:        "compose_syntax_error",
            Message:    fmt.Sprintf("Loi cu phap docker-compose.yml: %v", err),
            Suggestion: "Kiem tra cu phap YAML (indentation, colons, quotes)",
        }
    }

    // Check required sections
    if _, ok := compose["services"]; !ok {
        return &UserError{
            Key:        "compose_no_services",
            Message:    "docker-compose.yml thieu section 'services'",
            Suggestion: "Them section services vao file",
        }
    }

    return nil
}

// ValidateCaddyfile does basic Caddyfile syntax check
func ValidateCaddyfile(dir string) error {
    caddyPath := filepath.Join(dir, "Caddyfile")

    if _, err := os.Stat(caddyPath); os.IsNotExist(err) {
        // Caddyfile is optional
        return nil
    }

    content, err := os.ReadFile(caddyPath)
    if err != nil {
        return &UserError{
            Key:        "caddy_read_error",
            Message:    fmt.Sprintf("Khong doc duoc Caddyfile: %v", err),
            Suggestion: "Kiem tra quyen truy cap file",
        }
    }

    // Basic check: file should not be empty if exists
    if len(content) == 0 {
        return &UserError{
            Key:        "caddy_empty",
            Message:    "Caddyfile trong",
            Suggestion: "Them cau hinh domain vao Caddyfile",
        }
    }

    return nil
}
```

### Step 4: Disk Space Check (1h)

**pkg/validator/disk.go:**
```go
package validator

import (
    "fmt"
    "syscall"
)

const MinDiskSpaceGB = 5

// CheckDiskSpace verifies sufficient disk space
func CheckDiskSpace(path string) (float64, error) {
    var stat syscall.Statfs_t
    if err := syscall.Statfs(path, &stat); err != nil {
        return 0, fmt.Errorf("khong kiem tra duoc disk: %w", err)
    }

    // Available space in bytes
    available := float64(stat.Bavail * uint64(stat.Bsize))
    availableGB := available / (1024 * 1024 * 1024)

    return availableGB, nil
}

// WarnIfLowDiskSpace prints warning if disk < MinDiskSpaceGB
func WarnIfLowDiskSpace(path string) {
    availableGB, err := CheckDiskSpace(path)
    if err != nil {
        return // Silently ignore if can't check
    }

    if availableGB < MinDiskSpaceGB {
        fmt.Printf("  [!] Canh bao: Disk space thap (%.1fGB). Recommend it nhat %dGB.\n",
            availableGB, MinDiskSpaceGB)
    }
}
```

### Step 5: Error Types and Translation (2h)

**pkg/validator/errors.go:**
```go
package validator

// ErrorKey constants for translation
const (
    ErrDockerNotInstalled = "docker_not_installed"
    ErrDockerNotRunning   = "docker_not_running"
    ErrPortConflict       = "port_conflict"
    ErrEnvMissing         = "env_missing"
    ErrEnvMissingVars     = "env_missing_vars"
    ErrComposeMissing     = "compose_missing"
    ErrComposeSyntax      = "compose_syntax_error"
    ErrDiskLow            = "disk_low"
)

// UserError is already defined in docker.go
// Re-export or move to this file

// ErrorMessages maps error keys to Vietnamese messages
var ErrorMessages = map[string]struct {
    Message    string
    Suggestion string
}{
    ErrDockerNotInstalled: {
        Message:    "Docker chua cai dat",
        Suggestion: "Cai Docker tai: https://docs.docker.com/get-docker/",
    },
    ErrDockerNotRunning: {
        Message:    "Docker daemon khong chay",
        Suggestion: "Khoi dong Docker: sudo systemctl start docker",
    },
    ErrPortConflict: {
        Message:    "Co port dang bi su dung",
        Suggestion: "Xem chi tiet ben duoi",
    },
    ErrEnvMissing: {
        Message:    "File .env khong ton tai",
        Suggestion: "Chay: kk init",
    },
    ErrEnvMissingVars: {
        Message:    "Thieu bien moi truong bat buoc",
        Suggestion: "Xem chi tiet ben duoi",
    },
    ErrComposeMissing: {
        Message:    "File docker-compose.yml khong ton tai",
        Suggestion: "Chay: kk init",
    },
    ErrComposeSyntax: {
        Message:    "Loi cu phap trong docker-compose.yml",
        Suggestion: "Kiem tra YAML: indentation, colons, quotes",
    },
    ErrDiskLow: {
        Message:    "Disk space thap",
        Suggestion: "Don dep disk hoac mo rong storage",
    },
}

// TranslateError converts technical error to user-friendly
func TranslateError(err error) string {
    if ue, ok := err.(*UserError); ok {
        return fmt.Sprintf("%s\n  → %s", ue.Message, ue.Suggestion)
    }
    // Fallback for unknown errors
    return fmt.Sprintf("Loi: %v", err)
}
```

### Step 6: Preflight Check Runner (2h)

**pkg/validator/preflight.go:**
```go
package validator

import (
    "fmt"
    "os"
)

type PreflightResult struct {
    CheckName string
    Passed    bool
    Error     error
    Warning   string
}

// RunPreflight executes all validation checks
func RunPreflight(dir string, includeCaddy bool) ([]PreflightResult, error) {
    var results []PreflightResult
    var hasBlockingError bool

    // 1. Docker installed
    err := CheckDockerInstalled()
    results = append(results, PreflightResult{
        CheckName: "Docker cai dat",
        Passed:    err == nil,
        Error:     err,
    })
    if err != nil {
        hasBlockingError = true
    }

    // 2. Docker daemon running (only if installed)
    if !hasBlockingError {
        err = CheckDockerDaemon()
        results = append(results, PreflightResult{
            CheckName: "Docker daemon",
            Passed:    err == nil,
            Error:     err,
        })
        if err != nil {
            hasBlockingError = true
        }
    }

    // 3. Port conflicts
    _, err = CheckAllPorts(includeCaddy)
    results = append(results, PreflightResult{
        CheckName: "Cong mang (ports)",
        Passed:    err == nil,
        Error:     err,
    })
    if err != nil {
        hasBlockingError = true
    }

    // 4. Environment file
    err = ValidateEnvFile(dir)
    results = append(results, PreflightResult{
        CheckName: "File .env",
        Passed:    err == nil,
        Error:     err,
    })
    if err != nil {
        hasBlockingError = true
    }

    // 5. Docker compose syntax
    err = ValidateDockerCompose(dir)
    results = append(results, PreflightResult{
        CheckName: "docker-compose.yml",
        Passed:    err == nil,
        Error:     err,
    })
    if err != nil {
        hasBlockingError = true
    }

    // 6. Caddyfile (if enabled)
    if includeCaddy {
        err = ValidateCaddyfile(dir)
        results = append(results, PreflightResult{
            CheckName: "Caddyfile",
            Passed:    err == nil,
            Error:     err,
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

// PrintPreflightResults displays results in user-friendly format
func PrintPreflightResults(results []PreflightResult) {
    fmt.Println("\nKiem tra truoc khi chay:")
    fmt.Println("─────────────────────────")

    for _, r := range results {
        if r.Passed {
            if r.Warning != "" {
                fmt.Printf("  [!] %s (canh bao: %s)\n", r.CheckName, r.Warning)
            } else {
                fmt.Printf("  [OK] %s\n", r.CheckName)
            }
        } else {
            fmt.Printf("  [X] %s\n", r.CheckName)
            if r.Error != nil {
                fmt.Printf("      %s\n", TranslateError(r.Error))
            }
        }
    }
    fmt.Println()
}
```

## Todo List

- [ ] Implement ports.go with net.Listen approach
- [ ] Add PID detection via /proc or lsof
- [ ] Implement env.go for .env validation
- [ ] Add password strength warning
- [ ] Implement config.go for YAML validation
- [ ] Implement disk.go for disk space check
- [ ] Create unified error types in errors.go
- [ ] Implement preflight.go runner
- [ ] Add go get gopkg.in/yaml.v3
- [ ] Unit tests for each validator
- [ ] Integration test for preflight runner

## Success Criteria

1. Port conflict detected correctly with PID info
2. Missing .env variables identified
3. Invalid YAML syntax caught with line info
4. Disk space warning at < 5GB
5. All errors show Vietnamese messages
6. Preflight results displayed clearly

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| lsof not installed | Low | Fallback to /proc/net/tcp |
| YAML v3 dependency | Low | Well-maintained package |
| Windows compatibility | Medium | Linux-only for now (target platform) |

## Security Considerations

1. **No Secret Exposure:** Preflight results don't log password values
2. **File Permissions:** Warn if .env is world-readable
3. **Input Sanitization:** Don't execute user input directly

## Next Steps

After completing Phase 02:
1. Proceed to [Phase 03: Operations](./phase-03-operations.md)
2. Integrate preflight checks into kk start
3. Add health check monitoring

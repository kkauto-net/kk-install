# Phase 03: Operations

## Context

- **Parent Plan:** [plan.md](./plan.md)
- **Dependencies:** [Phase 01](./phase-01-core-foundation.md), [Phase 02](./phase-02-validation-layer.md)
- **Related Research:** [Docker Integration](./research/researcher-02-docker-integration.md)

## Overview

| Field | Value |
|-------|-------|
| Date | 2026-01-04 |
| Description | kk start with monitoring, health checks, kk status, kk restart, progress indicators |
| Priority | P1 |
| Status | pending |
| Effort | 1 week |

## Key Insights (from Research)

1. **Health Check Monitoring:** Docker SDK `ContainerInspect()` -> `State.Health.Status`
2. **Retry Strategy:** Exponential backoff with jitter, max 3 retries
3. **Timeout:** `context.WithTimeout` for all Docker operations
4. **Progress Indicators:** pterm or spinner for visual feedback

## Requirements

- [x] `kk start` command with preflight + docker-compose up
- [x] Health check monitoring with auto-retry (3x)
- [x] Progress indicators during operations
- [x] `kk status` with formatted table output
- [x] `kk restart` command
- [x] Graceful handling of SIGINT/SIGTERM
- [x] Service status table with access URLs

## Architecture

```
pkg/
├── compose/
│   ├── executor.go   # docker-compose wrapper
│   └── parser.go     # Parse compose file for service list
├── monitor/
│   ├── health.go     # Health check with retry
│   └── status.go     # Get container status
└── ui/
    ├── progress.go   # Spinners, progress bars
    └── table.go      # Status table formatting
cmd/
├── start.go
├── status.go
└── restart.go
```

## Related Code Files

After implementation:
- `/home/kkdev/kkcli/cmd/start.go`
- `/home/kkdev/kkcli/cmd/status.go`
- `/home/kkdev/kkcli/cmd/restart.go`
- `/home/kkdev/kkcli/pkg/compose/executor.go`
- `/home/kkdev/kkcli/pkg/compose/parser.go`
- `/home/kkdev/kkcli/pkg/monitor/health.go`
- `/home/kkdev/kkcli/pkg/monitor/status.go`
- `/home/kkdev/kkcli/pkg/ui/progress.go`
- `/home/kkdev/kkcli/pkg/ui/table.go`

## Implementation Steps

### Step 1: Docker Compose Executor (3h)

**pkg/compose/executor.go:**
```go
package compose

import (
    "bytes"
    "context"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "time"
)

// Executor wraps docker-compose commands
type Executor struct {
    WorkDir     string
    ComposeFile string
}

func NewExecutor(workDir string) *Executor {
    return &Executor{
        WorkDir:     workDir,
        ComposeFile: filepath.Join(workDir, "docker-compose.yml"),
    }
}

// Up runs docker-compose up -d
func (e *Executor) Up(ctx context.Context) error {
    return e.run(ctx, "up", "-d")
}

// Down runs docker-compose down
func (e *Executor) Down(ctx context.Context) error {
    return e.run(ctx, "down")
}

// Restart runs docker-compose restart
func (e *Executor) Restart(ctx context.Context) error {
    return e.run(ctx, "restart")
}

// Pull runs docker-compose pull
func (e *Executor) Pull(ctx context.Context) (string, error) {
    return e.runWithOutput(ctx, "pull")
}

// Ps runs docker-compose ps
func (e *Executor) Ps(ctx context.Context) (string, error) {
    return e.runWithOutput(ctx, "ps", "--format", "json")
}

// ForceRecreate runs docker-compose up -d --force-recreate
func (e *Executor) ForceRecreate(ctx context.Context) error {
    return e.run(ctx, "up", "-d", "--force-recreate")
}

func (e *Executor) run(ctx context.Context, args ...string) error {
    cmd := e.buildCmd(ctx, args...)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}

func (e *Executor) runWithOutput(ctx context.Context, args ...string) (string, error) {
    cmd := e.buildCmd(ctx, args...)
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    err := cmd.Run()
    if err != nil {
        return "", fmt.Errorf("%w: %s", err, stderr.String())
    }
    return stdout.String(), nil
}

func (e *Executor) buildCmd(ctx context.Context, args ...string) *exec.Cmd {
    // Try docker compose (v2) first, fallback to docker-compose (v1)
    cmdName := "docker"
    cmdArgs := append([]string{"compose", "-f", e.ComposeFile}, args...)

    // Check if docker compose v2 is available
    if _, err := exec.LookPath("docker"); err == nil {
        testCmd := exec.Command("docker", "compose", "version")
        if testCmd.Run() != nil {
            // Fallback to docker-compose v1
            cmdName = "docker-compose"
            cmdArgs = append([]string{"-f", e.ComposeFile}, args...)
        }
    }

    cmd := exec.CommandContext(ctx, cmdName, cmdArgs...)
    cmd.Dir = e.WorkDir
    return cmd
}

// DefaultTimeout for compose operations
const DefaultTimeout = 5 * time.Minute
```

### Step 2: Service Parser (2h)

**pkg/compose/parser.go:**
```go
package compose

import (
    "os"
    "path/filepath"

    "gopkg.in/yaml.v3"
)

type ComposeFile struct {
    Services map[string]Service `yaml:"services"`
}

type Service struct {
    Image       string            `yaml:"image"`
    Ports       []string          `yaml:"ports"`
    HealthCheck *HealthCheck      `yaml:"healthcheck"`
    DependsOn   interface{}       `yaml:"depends_on"`
}

type HealthCheck struct {
    Test     []string `yaml:"test"`
    Interval string   `yaml:"interval"`
    Timeout  string   `yaml:"timeout"`
    Retries  int      `yaml:"retries"`
}

// ParseComposeFile reads and parses docker-compose.yml
func ParseComposeFile(dir string) (*ComposeFile, error) {
    composePath := filepath.Join(dir, "docker-compose.yml")
    content, err := os.ReadFile(composePath)
    if err != nil {
        return nil, err
    }

    var compose ComposeFile
    if err := yaml.Unmarshal(content, &compose); err != nil {
        return nil, err
    }

    return &compose, nil
}

// GetServiceNames returns list of service names
func (c *ComposeFile) GetServiceNames() []string {
    var names []string
    for name := range c.Services {
        names = append(names, name)
    }
    return names
}

// HasHealthCheck returns true if service has healthcheck defined
func (c *ComposeFile) HasHealthCheck(serviceName string) bool {
    if svc, ok := c.Services[serviceName]; ok {
        return svc.HealthCheck != nil
    }
    return false
}

// GetServicePorts extracts exposed ports for a service
func (c *ComposeFile) GetServicePorts(serviceName string) []string {
    if svc, ok := c.Services[serviceName]; ok {
        return svc.Ports
    }
    return nil
}
```

### Step 3: Health Check Monitor (4h)

**pkg/monitor/health.go:**
```go
package monitor

import (
    "context"
    "fmt"
    "strings"
    "time"

    "github.com/docker/docker/api/types/container"
    "github.com/docker/docker/client"
)

const (
    MaxRetries     = 3
    InitialDelay   = 2 * time.Second
    MaxDelay       = 30 * time.Second
    CheckInterval  = 3 * time.Second
)

type HealthStatus struct {
    ServiceName string
    Container   string
    Status      string // healthy, unhealthy, starting, none
    Healthy     bool
    Message     string
}

// HealthMonitor checks container health status
type HealthMonitor struct {
    client *client.Client
}

func NewHealthMonitor() (*HealthMonitor, error) {
    cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
    if err != nil {
        return nil, fmt.Errorf("tao Docker client that bai: %w", err)
    }
    return &HealthMonitor{client: cli}, nil
}

func (m *HealthMonitor) Close() {
    m.client.Close()
}

// WaitForHealthy waits for container to become healthy with retry
func (m *HealthMonitor) WaitForHealthy(ctx context.Context, containerName string, hasHealthCheck bool) HealthStatus {
    status := HealthStatus{
        Container: containerName,
    }

    // Extract service name from container name (e.g., kkengine_db -> db)
    parts := strings.Split(containerName, "_")
    if len(parts) > 1 {
        status.ServiceName = parts[len(parts)-1]
    } else {
        status.ServiceName = containerName
    }

    // If no health check defined, just check if running
    if !hasHealthCheck {
        return m.checkRunning(ctx, containerName, status)
    }

    // Wait for health check with retries
    delay := InitialDelay
    for retry := 0; retry < MaxRetries; retry++ {
        result := m.checkHealth(ctx, containerName)
        if result.Healthy {
            return result
        }

        // Wait before retry
        select {
        case <-ctx.Done():
            status.Status = "timeout"
            status.Message = "Da het thoi gian cho"
            return status
        case <-time.After(delay):
            // Exponential backoff
            delay = min(delay*2, MaxDelay)
        }
    }

    // Final check after all retries
    return m.checkHealth(ctx, containerName)
}

func (m *HealthMonitor) checkHealth(ctx context.Context, containerName string) HealthStatus {
    status := HealthStatus{Container: containerName}

    info, err := m.client.ContainerInspect(ctx, containerName)
    if err != nil {
        status.Status = "error"
        status.Message = fmt.Sprintf("Khong kiem tra duoc: %v", err)
        return status
    }

    // Extract service name
    parts := strings.Split(containerName, "_")
    if len(parts) > 1 {
        status.ServiceName = parts[len(parts)-1]
    } else {
        status.ServiceName = containerName
    }

    // Check if health check exists
    if info.State.Health == nil {
        // No health check, just check running status
        if info.State.Running {
            status.Status = "running"
            status.Healthy = true
        } else {
            status.Status = "stopped"
            status.Message = fmt.Sprintf("Exit code: %d", info.State.ExitCode)
        }
        return status
    }

    // Check health status
    status.Status = info.State.Health.Status
    switch info.State.Health.Status {
    case "healthy":
        status.Healthy = true
    case "starting":
        status.Message = "Dang khoi dong..."
    case "unhealthy":
        // Get last health check log
        if len(info.State.Health.Log) > 0 {
            lastLog := info.State.Health.Log[len(info.State.Health.Log)-1]
            status.Message = lastLog.Output
        }
    }

    return status
}

func (m *HealthMonitor) checkRunning(ctx context.Context, containerName string, status HealthStatus) HealthStatus {
    info, err := m.client.ContainerInspect(ctx, containerName)
    if err != nil {
        status.Status = "error"
        status.Message = fmt.Sprintf("Khong kiem tra duoc: %v", err)
        return status
    }

    if info.State.Running {
        status.Status = "running"
        status.Healthy = true
    } else {
        status.Status = "stopped"
        status.Message = fmt.Sprintf("Exit code: %d", info.State.ExitCode)
    }

    return status
}

// MonitorAll waits for all containers to be healthy
func (m *HealthMonitor) MonitorAll(ctx context.Context, containers []ContainerInfo, onProgress func(HealthStatus)) []HealthStatus {
    var results []HealthStatus

    for _, c := range containers {
        // Report starting
        onProgress(HealthStatus{
            ServiceName: c.ServiceName,
            Container:   c.ContainerName,
            Status:      "starting",
            Message:     "Dang kiem tra...",
        })

        status := m.WaitForHealthy(ctx, c.ContainerName, c.HasHealthCheck)
        results = append(results, status)

        // Report result
        onProgress(status)
    }

    return results
}

type ContainerInfo struct {
    ServiceName    string
    ContainerName  string
    HasHealthCheck bool
}

func min(a, b time.Duration) time.Duration {
    if a < b {
        return a
    }
    return b
}
```

### Step 4: Status Checker (2h)

**pkg/monitor/status.go:**
```go
package monitor

import (
    "context"
    "encoding/json"
    "strings"

    "github.com/kkengine/kkcli/pkg/compose"
)

type ServiceStatus struct {
    Name    string
    Status  string
    Health  string
    Ports   string
    Running bool
}

// GetStatus returns status of all services
func GetStatus(ctx context.Context, executor *compose.Executor) ([]ServiceStatus, error) {
    output, err := executor.Ps(ctx)
    if err != nil {
        return nil, err
    }

    return parseComposePs(output)
}

// Docker compose ps --format json output structure
type composePsJSON struct {
    Name    string `json:"Name"`
    State   string `json:"State"`
    Health  string `json:"Health"`
    Ports   string `json:"Ports"`
    Service string `json:"Service"`
}

func parseComposePs(output string) ([]ServiceStatus, error) {
    var statuses []ServiceStatus

    // Each line is a JSON object
    lines := strings.Split(strings.TrimSpace(output), "\n")
    for _, line := range lines {
        if line == "" {
            continue
        }

        var ps composePsJSON
        if err := json.Unmarshal([]byte(line), &ps); err != nil {
            continue // Skip malformed lines
        }

        status := ServiceStatus{
            Name:    ps.Service,
            Status:  ps.State,
            Health:  ps.Health,
            Ports:   ps.Ports,
            Running: strings.ToLower(ps.State) == "running",
        }

        statuses = append(statuses, status)
    }

    return statuses, nil
}

// IsAllHealthy checks if all services are running/healthy
func IsAllHealthy(statuses []ServiceStatus) bool {
    for _, s := range statuses {
        if !s.Running {
            return false
        }
        // If health check exists, must be healthy
        if s.Health != "" && s.Health != "healthy" {
            return false
        }
    }
    return true
}
```

### Step 5: Progress UI (2h)

**pkg/ui/progress.go:**
```go
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
        done:    make(chan bool),
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
```

### Step 6: Status Table (2h)

**pkg/ui/table.go:**
```go
package ui

import (
    "fmt"
    "strings"

    "github.com/kkengine/kkcli/pkg/monitor"
)

// PrintStatusTable displays service status as formatted table
func PrintStatusTable(statuses []monitor.ServiceStatus) {
    // Calculate column widths
    nameWidth := 10
    statusWidth := 10
    healthWidth := 10
    portsWidth := 25

    for _, s := range statuses {
        if len(s.Name) > nameWidth {
            nameWidth = len(s.Name)
        }
    }

    // Print header
    fmt.Println()
    fmt.Println("Trang thai dich vu:")
    fmt.Println(strings.Repeat("─", nameWidth+statusWidth+healthWidth+portsWidth+10))
    fmt.Printf("│ %-*s │ %-*s │ %-*s │ %-*s │\n",
        nameWidth, "Service",
        statusWidth, "Status",
        healthWidth, "Health",
        portsWidth, "Ports")
    fmt.Println(strings.Repeat("─", nameWidth+statusWidth+healthWidth+portsWidth+10))

    // Print rows
    for _, s := range statuses {
        health := s.Health
        if health == "" {
            health = "-"
        }

        ports := s.Ports
        if ports == "" {
            ports = "-"
        }
        // Truncate ports if too long
        if len(ports) > portsWidth {
            ports = ports[:portsWidth-3] + "..."
        }

        statusIcon := "[OK]"
        if !s.Running {
            statusIcon = "[X]"
        }

        fmt.Printf("│ %-*s │ %s %-*s │ %-*s │ %-*s │\n",
            nameWidth, s.Name,
            statusIcon, statusWidth-4, s.Status,
            healthWidth, health,
            portsWidth, ports)
    }

    fmt.Println(strings.Repeat("─", nameWidth+statusWidth+healthWidth+portsWidth+10))
    fmt.Println()
}

// PrintAccessInfo shows access URLs for services
func PrintAccessInfo(statuses []monitor.ServiceStatus) {
    fmt.Println("Truy cap:")
    for _, s := range statuses {
        if !s.Running || s.Ports == "" {
            continue
        }

        // Parse ports to show URLs
        switch s.Name {
        case "kkengine":
            fmt.Printf("  - kkengine: http://localhost:8019\n")
        case "db":
            fmt.Printf("  - MariaDB: localhost:3307\n")
        case "caddy":
            fmt.Printf("  - Web: http://localhost (HTTPS: https://localhost)\n")
        }
    }
    fmt.Println()
}
```

### Step 7: Start Command (3h)

**cmd/start.go:**
```go
package cmd

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "syscall"

    "github.com/spf13/cobra"

    "github.com/kkengine/kkcli/pkg/compose"
    "github.com/kkengine/kkcli/pkg/monitor"
    "github.com/kkengine/kkcli/pkg/ui"
    "github.com/kkengine/kkcli/pkg/validator"
)

var startCmd = &cobra.Command{
    Use:   "start",
    Short: "Khoi dong kkengine Docker stack",
    Long:  `Chay preflight checks, sau do khoi dong tat ca services.`,
    RunE:  runStart,
}

func init() {
    rootCmd.AddCommand(startCmd)
}

func runStart(cmd *cobra.Command, args []string) error {
    cwd, err := os.Getwd()
    if err != nil {
        return err
    }

    // Setup graceful shutdown
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    go func() {
        <-sigChan
        fmt.Println("\n\nDang dung lai...")
        cancel()
    }()

    // Step 1: Detect if Caddy is enabled
    composeFile, err := compose.ParseComposeFile(cwd)
    includeCaddy := false
    if err == nil {
        _, includeCaddy = composeFile.Services["caddy"]
    }

    // Step 2: Run preflight checks
    fmt.Println("\nKiem tra truoc khi chay...")
    results, err := validator.RunPreflight(cwd, includeCaddy)
    validator.PrintPreflightResults(results)

    if err != nil {
        return fmt.Errorf("preflight checks that bai. Vui long sua loi tren")
    }

    // Step 3: Start docker-compose
    fmt.Println("Khoi dong services...")
    executor := compose.NewExecutor(cwd)

    timeoutCtx, timeoutCancel := context.WithTimeout(ctx, compose.DefaultTimeout)
    defer timeoutCancel()

    if err := executor.Up(timeoutCtx); err != nil {
        return fmt.Errorf("khoi dong that bai: %w", err)
    }

    // Step 4: Monitor health
    fmt.Println("\nDang kiem tra suc khoe dich vu...")

    healthMonitor, err := monitor.NewHealthMonitor()
    if err != nil {
        // Can't monitor, but services may still be running
        fmt.Printf("  [!] Khong the theo doi health: %v\n", err)
    } else {
        defer healthMonitor.Close()

        // Build container list
        var containers []monitor.ContainerInfo
        for name := range composeFile.Services {
            containers = append(containers, monitor.ContainerInfo{
                ServiceName:    name,
                ContainerName:  fmt.Sprintf("kkengine_%s", name),
                HasHealthCheck: composeFile.HasHealthCheck(name),
            })
        }

        // Monitor with progress callback
        healthResults := healthMonitor.MonitorAll(timeoutCtx, containers, func(status monitor.HealthStatus) {
            ui.ShowServiceProgress(status.ServiceName, status.Status)
        })

        // Check if all healthy
        allHealthy := true
        for _, r := range healthResults {
            if !r.Healthy {
                allHealthy = false
                break
            }
        }

        if !allHealthy {
            fmt.Println("\n[!] Mot so dich vu chua san sang. Kiem tra: kk status")
        }
    }

    // Step 5: Show status
    fmt.Println("\n[OK] Khoi dong hoan tat!")

    statuses, err := monitor.GetStatus(timeoutCtx, executor)
    if err == nil {
        ui.PrintStatusTable(statuses)
        ui.PrintAccessInfo(statuses)
    }

    return nil
}
```

### Step 8: Status Command (1h)

**cmd/status.go:**
```go
package cmd

import (
    "context"
    "fmt"
    "os"
    "time"

    "github.com/spf13/cobra"

    "github.com/kkengine/kkcli/pkg/compose"
    "github.com/kkengine/kkcli/pkg/monitor"
    "github.com/kkengine/kkcli/pkg/ui"
)

var statusCmd = &cobra.Command{
    Use:   "status",
    Short: "Xem trang thai dich vu",
    Long:  `Hien thi trang thai tat ca containers trong stack.`,
    RunE:  runStatus,
}

func init() {
    rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
    cwd, err := os.Getwd()
    if err != nil {
        return err
    }

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    executor := compose.NewExecutor(cwd)
    statuses, err := monitor.GetStatus(ctx, executor)
    if err != nil {
        return fmt.Errorf("khong lay duoc trang thai: %w", err)
    }

    if len(statuses) == 0 {
        fmt.Println("Khong co dich vu nao dang chay.")
        fmt.Println("Chay: kk start")
        return nil
    }

    ui.PrintStatusTable(statuses)
    ui.PrintAccessInfo(statuses)

    // Summary
    running := 0
    for _, s := range statuses {
        if s.Running {
            running++
        }
    }

    if running == len(statuses) {
        fmt.Printf("[OK] Tat ca %d dich vu dang chay.\n", running)
    } else {
        fmt.Printf("[!] %d/%d dich vu dang chay.\n", running, len(statuses))
    }

    return nil
}
```

### Step 9: Restart Command (1h)

**cmd/restart.go:**
```go
package cmd

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "syscall"

    "github.com/spf13/cobra"

    "github.com/kkengine/kkcli/pkg/compose"
    "github.com/kkengine/kkcli/pkg/monitor"
    "github.com/kkengine/kkcli/pkg/ui"
)

var restartCmd = &cobra.Command{
    Use:   "restart",
    Short: "Khoi dong lai tat ca dich vu",
    Long:  `Restart tat ca containers trong stack.`,
    RunE:  runRestart,
}

func init() {
    rootCmd.AddCommand(restartCmd)
}

func runRestart(cmd *cobra.Command, args []string) error {
    cwd, err := os.Getwd()
    if err != nil {
        return err
    }

    // Setup graceful shutdown
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    go func() {
        <-sigChan
        fmt.Println("\n\nDang dung lai...")
        cancel()
    }()

    fmt.Println("Dang khoi dong lai dich vu...")

    executor := compose.NewExecutor(cwd)

    timeoutCtx, timeoutCancel := context.WithTimeout(ctx, compose.DefaultTimeout)
    defer timeoutCancel()

    if err := executor.Restart(timeoutCtx); err != nil {
        return fmt.Errorf("restart that bai: %w", err)
    }

    fmt.Println("[OK] Da khoi dong lai.")

    // Step 2: Monitor health
    composeFile, err := compose.ParseComposeFile(cwd)
    if err == nil {
        healthMonitor, err := monitor.NewHealthMonitor()
        if err == nil {
            defer healthMonitor.Close()

            fmt.Println("\nDang kiem tra suc khoe...")

            var containers []monitor.ContainerInfo
            for name := range composeFile.Services {
                containers = append(containers, monitor.ContainerInfo{
                    ServiceName:    name,
                    ContainerName:  fmt.Sprintf("kkengine_%s", name),
                    HasHealthCheck: composeFile.HasHealthCheck(name),
                })
            }

            healthMonitor.MonitorAll(timeoutCtx, containers, func(status monitor.HealthStatus) {
                ui.ShowServiceProgress(status.ServiceName, status.Status)
            })
        }
    }

    // Show final status
    statuses, err := monitor.GetStatus(timeoutCtx, executor)
    if err == nil {
        ui.PrintStatusTable(statuses)
    }

    return nil
}
```

## Todo List

- [ ] Implement compose/executor.go
- [ ] Implement compose/parser.go
- [ ] Add Docker SDK dependency: `go get github.com/docker/docker/client`
- [ ] Implement monitor/health.go with retry logic
- [ ] Implement monitor/status.go
- [ ] Implement ui/progress.go (spinner)
- [ ] Implement ui/table.go
- [ ] Implement cmd/start.go
- [ ] Implement cmd/status.go
- [ ] Implement cmd/restart.go
- [ ] Test start command flow
- [ ] Test health check retry (simulate unhealthy container)
- [ ] Test graceful shutdown (SIGINT handling)

## Success Criteria

1. `kk start` runs preflight, starts stack, monitors health
2. Health check retries 3x on failure
3. `kk status` shows formatted table
4. `kk restart` restarts all services safely
5. SIGINT (Ctrl+C) stops operations gracefully
6. Progress indicators show during operations

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| Docker SDK version mismatch | Medium | Use APIVersionNegotiation |
| Health check timeout | Low | Configurable timeout, reasonable defaults |
| JSON parse errors (docker compose ps) | Low | Handle gracefully, fallback to text output |

## Security Considerations

1. **No Secret Exposure:** Don't log container environment variables
2. **Context Timeout:** Always use timeouts for Docker operations
3. **Signal Handling:** Properly cleanup on SIGINT/SIGTERM

## Next Steps

After completing Phase 03:
1. Proceed to [Phase 04: Advanced Features](./phase-04-advanced-features.md)
2. Implement `kk update` command
3. Add testing and documentation

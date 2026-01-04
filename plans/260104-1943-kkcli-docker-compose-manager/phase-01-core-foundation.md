# Phase 01: Core Foundation

## Context

- **Parent Plan:** [plan.md](./plan.md)
- **Dependencies:** None (starting phase)
- **Related Research:** [Go CLI Ecosystem](./research/researcher-01-go-cli-ecosystem.md)

## Overview

| Field | Value |
|-------|-------|
| Date | 2026-01-04 |
| Description | Setup Go module, Cobra boilerplate, kk init command, template embedding |
| Priority | P1 |
| Status | DONE
completed: 2026-01-04 |
| Effort | 1 week |
| Review | [code-reviewer-260104-2246-phase-01-implementation.md](../reports/code-reviewer-260104-2246-phase-01-implementation.md) |

## Key Insights (from Research)

1. **Cobra** la industry standard cho Go CLI (kubectl, Hugo, GitHub CLI)
2. **go:embed** don gian hoa viec nhung templates vao binary
3. **survey** hoac **promptui** cho interactive prompts
4. **crypto/rand** cho secure password generation
5. Static binary: `CGO_ENABLED=0 go build -ldflags="-s -w"`

## Requirements

- [x] Go module initialization
- [x] Cobra CLI scaffolding with root command
- [x] `kk init` command with interactive service selection
- [x] Template embedding system (docker-compose.yml, .env, Caddyfile, kkfiler.toml, kkphp.conf)
- [x] Template rendering with conditional sections
- [x] Secure password generation (DB, Redis)
- [x] Basic Docker daemon check

## Architecture

```
kkcli/
├── main.go
├── go.mod
├── go.sum
├── cmd/
│   ├── root.go          # Root command, version, help
│   └── init.go          # kk init command
├── pkg/
│   ├── templates/
│   │   └── embed.go     # Template embedding + rendering
│   ├── validator/
│   │   └── docker.go    # Basic Docker checks
│   └── ui/
│       └── messages.go  # Vietnamese messages
└── templates/
    ├── docker-compose.yml.tmpl
    ├── env.tmpl
    ├── Caddyfile.tmpl
    ├── kkfiler.toml.tmpl
    └── kkphp.conf.tmpl
```

## Related Code Files

After implementation, these files will exist:
- `/home/kkdev/kkcli/main.go`
- `/home/kkdev/kkcli/cmd/root.go`
- `/home/kkdev/kkcli/cmd/init.go`
- `/home/kkdev/kkcli/pkg/templates/embed.go`
- `/home/kkdev/kkcli/pkg/validator/docker.go`
- `/home/kkdev/kkcli/pkg/ui/messages.go`
- `/home/kkdev/kkcli/templates/*.tmpl`

## Implementation Steps

### Step 1: Project Setup (2h)

```bash
# Initialize Go module
go mod init github.com/kkengine/kkcli

# Install dependencies
go get github.com/spf13/cobra@latest
go get github.com/AlecAivazis/survey/v2@latest
# OR go get github.com/manifoldco/promptui@latest
```

**main.go:**
```go
package main

import "github.com/kkengine/kkcli/cmd"

func main() {
    cmd.Execute()
}
```

### Step 2: Root Command (1h)

**cmd/root.go:**
```go
package cmd

import (
    "fmt"
    "os"
    "github.com/spf13/cobra"
)

var Version = "0.1.0"

var rootCmd = &cobra.Command{
    Use:   "kk",
    Short: "KK CLI - Docker Compose management for kkengine",
    Long:  `KK CLI giup ban quan ly kkengine Docker stack de dang.`,
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}

func init() {
    rootCmd.Version = Version
}
```

### Step 3: Template Embedding (3h)

**templates/docker-compose.yml.tmpl:**
```yaml
services:
  kkengine:
    image: kkengine:latest
    container_name: kkengine_app
    restart: unless-stopped
    stop_grace_period: 10s
    ports:
      - "8019:8019"
    env_file:
      - ./.env
    volumes:
      - ./kkphp.conf:/config/kkphp.conf
    networks:
      - kkengine_net
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_started
{{- if .EnableSeaweedFS }}
      seaweedfs:
        condition: service_healthy
{{- end }}

  db:
    image: mariadb:10.6
    container_name: kkengine_db
    restart: unless-stopped
    stop_grace_period: 10s
    environment:
      MYSQL_ROOT_PASSWORD: ${DB_ROOT_PASSWORD}
      MYSQL_DATABASE: ${DB_DATABASE}
      MYSQL_USER: ${DB_USERNAME}
      MYSQL_PASSWORD: ${DB_PASSWORD}
    volumes:
      - ${SYSTEM_DATABASE:-./data_database}:/var/lib/mysql
    ports:
      - "3307:3306"
    networks:
      - kkengine_net
    healthcheck:
      test: ["CMD", "healthcheck.sh", "--connect", "--innodb_initialized"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 30s

  redis:
    image: redis:alpine
    container_name: kkengine_redis
    restart: unless-stopped
    command: redis-server --requirepass ${REDIS_PASSWORD}
    volumes:
      - redis_data:/data
    networks:
      - kkengine_net
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 3s
      retries: 5

{{- if .EnableSeaweedFS }}
  seaweedfs:
    image: chrislusf/seaweedfs:latest
    container_name: kkengine_seaweedfs
    restart: unless-stopped
    stop_grace_period: 10s
    command: >
      server -dir=/data -master.port=9333 -volume.port=8080 -filer -filer.port=8888 -s3 -s3.port=8333 -master.defaultReplication=000 -volume.max=0
    env_file:
      - ./.env
    environment:
      WEED_MYSQL_ENABLED: "true"
      WEED_MYSQL_HOSTNAME: ${DB_HOSTNAME}
      WEED_MYSQL_PORT: ${DB_PORT}
      WEED_MYSQL_USERNAME: ${DB_USERNAME}
      WEED_MYSQL_PASSWORD: ${DB_PASSWORD}
      WEED_MYSQL_DATABASE: ${DB_SEAWEEDFS}
    volumes:
      - ${SYSTEM_FILESTORE:-./data_file}:/data
      - ./kkfiler.toml:/etc/seaweedfs/filer.toml:ro
    networks:
      - kkengine_net
    depends_on:
      db:
        condition: service_healthy
    healthcheck:
      test: ["CMD-SHELL", "pgrep -f 'weed.*server' > /dev/null && timeout 2 bash -c 'exec 3<>/dev/tcp/localhost/8888' 2>/dev/null || exit 1"]
      interval: 10s
      timeout: 5s
      retries: 12
      start_period: 50s
{{- end }}

{{- if .EnableCaddy }}
  caddy:
    image: caddy:alpine
    container_name: kkengine_caddy
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    env_file:
      - ./.env
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile
      - caddy_data:/data
      - caddy_config:/config
    networks:
      - kkengine_net
    depends_on:
      - kkengine
{{- end }}

networks:
  kkengine_net:
    name: kkengine_net
    driver: bridge

volumes:
  redis_data:
{{- if .EnableCaddy }}
  caddy_data:
  caddy_config:
{{- end }}
```

**templates/env.tmpl:**
```
# KKEngine Configuration
# Generated by kk init

# Database
DB_HOSTNAME=db
DB_PORT=3306
DB_DATABASE=kkengine
DB_USERNAME=kkengine
DB_PASSWORD={{.DBPassword}}
DB_ROOT_PASSWORD={{.DBRootPassword}}
{{- if .EnableSeaweedFS }}
DB_SEAWEEDFS=kkengine_seaweedfs
{{- end }}

# Redis
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD={{.RedisPassword}}

{{- if .EnableCaddy }}
# Caddy
SYSTEM_DOMAIN={{.Domain}}
{{- end }}

# System paths (optional, can customize)
# SYSTEM_DATABASE=./data_database
# SYSTEM_FILESTORE=./data_file
```

**pkg/templates/embed.go:**
```go
package templates

import (
    "embed"
    "os"
    "path/filepath"
    "text/template"
)

//go:embed ../../templates/*
var templateFS embed.FS

type Config struct {
    EnableSeaweedFS  bool
    EnableCaddy      bool
    DBPassword       string
    DBRootPassword   string
    RedisPassword    string
    Domain           string
}

func RenderTemplate(name string, cfg Config, outputPath string) error {
    tmplContent, err := templateFS.ReadFile("templates/" + name + ".tmpl")
    if err != nil {
        return err
    }

    tmpl, err := template.New(name).Parse(string(tmplContent))
    if err != nil {
        return err
    }

    // Ensure directory exists
    dir := filepath.Dir(outputPath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return err
    }

    file, err := os.Create(outputPath)
    if err != nil {
        return err
    }
    defer file.Close()

    return tmpl.Execute(file, cfg)
}

func RenderAll(cfg Config, targetDir string) error {
    files := map[string]string{
        "docker-compose.yml": "docker-compose.yml",
        "env":                ".env",
        "kkphp.conf":         "kkphp.conf",
    }

    if cfg.EnableCaddy {
        files["Caddyfile"] = "Caddyfile"
    }
    if cfg.EnableSeaweedFS {
        files["kkfiler.toml"] = "kkfiler.toml"
    }

    for tmplName, outputName := range files {
        outputPath := filepath.Join(targetDir, outputName)
        if err := RenderTemplate(tmplName, cfg, outputPath); err != nil {
            return err
        }
    }
    return nil
}
```

### Step 4: Password Generation (1h)

**pkg/ui/passwords.go:**
```go
package ui

import (
    "crypto/rand"
    "encoding/base64"
)

// GeneratePassword creates cryptographically secure random password
func GeneratePassword(length int) (string, error) {
    bytes := make([]byte, length)
    if _, err := rand.Read(bytes); err != nil {
        return "", err
    }
    // Use URL-safe base64, no special chars that might break shell
    return base64.RawURLEncoding.EncodeToString(bytes)[:length], nil
}
```

### Step 5: Basic Docker Check (2h)

**pkg/validator/docker.go:**
```go
package validator

import (
    "context"
    "os/exec"
    "time"
)

// CheckDockerInstalled verifies docker binary exists
func CheckDockerInstalled() error {
    _, err := exec.LookPath("docker")
    if err != nil {
        return &UserError{
            Key:        "docker_not_installed",
            Message:    "Docker chua cai dat",
            Suggestion: "Cai tai: https://docs.docker.com/get-docker/",
        }
    }
    return nil
}

// CheckDockerDaemon verifies docker daemon is running
func CheckDockerDaemon() error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    cmd := exec.CommandContext(ctx, "docker", "info")
    if err := cmd.Run(); err != nil {
        return &UserError{
            Key:        "docker_not_running",
            Message:    "Docker daemon khong chay",
            Suggestion: "Chay: sudo systemctl start docker",
        }
    }
    return nil
}

// UserError represents user-friendly error
type UserError struct {
    Key        string
    Message    string
    Suggestion string
}

func (e *UserError) Error() string {
    return e.Message
}
```

### Step 6: Vietnamese Messages (1h)

**pkg/ui/messages.go:**
```go
package ui

import "fmt"

// Success messages
func MsgCheckingDocker() string { return "Dang kiem tra Docker..." }
func MsgDockerOK() string       { return "Docker da san sang" }
func MsgCreated(file string) string { return fmt.Sprintf("Da tao: %s", file) }
func MsgInitComplete() string   { return "Khoi tao hoan tat!" }

// Error messages
func MsgDockerNotInstalled() string { return "Docker chua cai dat" }
func MsgDockerNotRunning() string   { return "Docker daemon khong chay" }

// Next steps
func MsgNextSteps() string {
    return `
Buoc tiep theo:
  1. Kiem tra va chinh sua .env neu can
  2. Chay: kk start
`
}

// Progress indicators
func ShowSuccess(msg string) { fmt.Printf("  [OK] %s\n", msg) }
func ShowError(msg string)   { fmt.Printf("  [X] %s\n", msg) }
func ShowInfo(msg string)    { fmt.Printf("  [>] %s\n", msg) }
```

### Step 7: Init Command (4h)

**cmd/init.go:**
```go
package cmd

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/AlecAivazis/survey/v2"
    "github.com/spf13/cobra"

    "github.com/kkengine/kkcli/pkg/templates"
    "github.com/kkengine/kkcli/pkg/ui"
    "github.com/kkengine/kkcli/pkg/validator"
)

var initCmd = &cobra.Command{
    Use:   "init",
    Short: "Khoi tao kkengine Docker stack",
    Long:  `Tao docker-compose.yml va cac file config can thiet.`,
    RunE:  runInit,
}

func init() {
    rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
    // Step 1: Check Docker
    ui.ShowInfo(ui.MsgCheckingDocker())
    if err := validator.CheckDockerInstalled(); err != nil {
        return err
    }
    if err := validator.CheckDockerDaemon(); err != nil {
        return err
    }
    ui.ShowSuccess(ui.MsgDockerOK())

    // Step 2: Get working directory
    cwd, err := os.Getwd()
    if err != nil {
        return err
    }
    fmt.Printf("\nKhoi tao trong: %s\n\n", cwd)

    // Step 3: Check if already initialized
    composePath := filepath.Join(cwd, "docker-compose.yml")
    if _, err := os.Stat(composePath); err == nil {
        var overwrite bool
        prompt := &survey.Confirm{
            Message: "docker-compose.yml da ton tai. Ghi de?",
            Default: false,
        }
        survey.AskOne(prompt, &overwrite)
        if !overwrite {
            return fmt.Errorf("huy khoi tao")
        }
    }

    // Step 4: Interactive prompts
    var enableSeaweedFS bool
    var enableCaddy bool
    var domain string

    survey.AskOne(&survey.Confirm{
        Message: "Bat SeaweedFS file storage?",
        Default: false,
    }, &enableSeaweedFS)

    survey.AskOne(&survey.Confirm{
        Message: "Bat Caddy web server?",
        Default: false,
    }, &enableCaddy)

    if enableCaddy {
        survey.AskOne(&survey.Input{
            Message: "Nhap domain (vd: example.com):",
            Default: "localhost",
        }, &domain)
    }

    // Step 5: Generate passwords
    dbPass, _ := ui.GeneratePassword(24)
    dbRootPass, _ := ui.GeneratePassword(24)
    redisPass, _ := ui.GeneratePassword(24)

    // Step 6: Render templates
    cfg := templates.Config{
        EnableSeaweedFS:  enableSeaweedFS,
        EnableCaddy:      enableCaddy,
        DBPassword:       dbPass,
        DBRootPassword:   dbRootPass,
        RedisPassword:    redisPass,
        Domain:           domain,
    }

    if err := templates.RenderAll(cfg, cwd); err != nil {
        return fmt.Errorf("loi khi tao file: %w", err)
    }

    // Step 7: Show success
    fmt.Println()
    ui.ShowSuccess(ui.MsgCreated("docker-compose.yml"))
    ui.ShowSuccess(ui.MsgCreated(".env"))
    ui.ShowSuccess(ui.MsgCreated("kkphp.conf"))
    if enableCaddy {
        ui.ShowSuccess(ui.MsgCreated("Caddyfile"))
    }
    if enableSeaweedFS {
        ui.ShowSuccess(ui.MsgCreated("kkfiler.toml"))
    }

    fmt.Println()
    fmt.Println(ui.MsgInitComplete())
    fmt.Println(ui.MsgNextSteps())

    return nil
}
```

## Todo List

- [x] Initialize Go module
- [x] Create directory structure
- [x] Implement root.go with version command
- [x] Create all template files (docker-compose.yml.tmpl, env.tmpl, etc)
- [x] Implement embed.go for template embedding
- [x] Implement password generation with crypto/rand
- [x] Implement basic Docker validation
- [x] Implement Vietnamese messages
- [x] Implement init command with interactive prompts
- [x] Test init command flow
- [x] Build static binary and verify size

**Review Findings (2026-01-04):**
- ⚠️ 2 fixes required before Phase 02:
  1. Remove unused "time" import in kk_integration_test.go:12
  2. Fix docker-compose template to use env vars instead of hardcoded passwords

## Success Criteria

1. `go build` produces working binary
2. `kk --version` shows version
3. `kk init` runs interactive prompts
4. Files generated correctly based on selections
5. Passwords are cryptographically random (not predictable)
6. Docker check blocks if not installed/running
7. Binary size < 15MB (before UPX compression)

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| survey package deprecated | Medium | Can switch to promptui or huh |
| Template syntax errors | Low | Unit test each template |
| embed path issues | Low | Test in different directories |

## Security Considerations

1. **Password Generation:** Use crypto/rand, not math/rand
2. **No Logging Secrets:** Never log passwords to stdout/stderr
3. **.env Permissions:** Set 0600 (owner read/write only)
4. **Template Injection:** Validate user input before templating

## Next Steps

After completing Phase 01:
1. Proceed to [Phase 02: Validation Layer](./phase-02-validation-layer.md)
2. Add port conflict detection
3. Add env validation
4. Add config syntax validation

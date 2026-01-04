# Phase 04: Advanced Features

## Context

- **Parent Plan:** [plan.md](./plan.md)
- **Dependencies:** [Phase 01](./phase-01-core-foundation.md), [Phase 02](./phase-02-validation-layer.md), [Phase 03](./phase-03-operations.md)
- **Related Research:** [Docker Integration](./research/researcher-02-docker-integration.md)

## Overview

| Field | Value |
|-------|-------|
| Date | 2026-01-04 |
| Description | kk update command, image pull tracking, testing, documentation, distribution |
| Priority | P2 |
| Status | pending |
| Effort | 1 week |

## Key Insights (from Research)

1. **Image Pull:** `docker-compose pull` + parse output for updates
2. **Distribution:** GitHub Releases with static binaries
3. **Testing:** Table-driven tests, golden file testing, testscript for E2E
4. **Shell Completions:** Cobra built-in support

## Requirements

- [x] `kk update` command with image pull + confirmation
- [x] Show which images have updates
- [x] Confirmation before recreating containers
- [x] Unit tests for validators
- [x] Integration tests for commands
- [x] Build automation (Makefile/goreleaser)
- [x] Install script for easy distribution
- [x] Shell completions (bash, zsh)

## Architecture

```
kkcli/
├── cmd/
│   ├── update.go     # kk update command
│   └── completion.go # Shell completions
├── pkg/
│   └── updater/
│       └── updater.go # Image update logic
├── Makefile
├── .goreleaser.yml
├── scripts/
│   └── install.sh    # Curl install script
└── tests/
    ├── validator_test.go
    ├── compose_test.go
    └── integration_test.go
```

## Related Code Files

After implementation:
- `/home/kkdev/kkcli/cmd/update.go`
- `/home/kkdev/kkcli/cmd/completion.go`
- `/home/kkdev/kkcli/pkg/updater/updater.go`
- `/home/kkdev/kkcli/Makefile`
- `/home/kkdev/kkcli/.goreleaser.yml`
- `/home/kkdev/kkcli/scripts/install.sh`
- `/home/kkdev/kkcli/pkg/validator/docker_test.go`
- `/home/kkdev/kkcli/pkg/validator/ports_test.go`

## Implementation Steps

### Step 1: Update Command (4h)

**cmd/update.go:**
```go
package cmd

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "syscall"

    "github.com/AlecAivazis/survey/v2"
    "github.com/spf13/cobra"

    "github.com/kkengine/kkcli/pkg/compose"
    "github.com/kkengine/kkcli/pkg/monitor"
    "github.com/kkengine/kkcli/pkg/ui"
    "github.com/kkengine/kkcli/pkg/updater"
)

var updateCmd = &cobra.Command{
    Use:   "update",
    Short: "Cap nhat images moi nhat",
    Long:  `Kiem tra va tai images moi tu Docker Hub, sau do restart services.`,
    RunE:  runUpdate,
}

var forceUpdate bool

func init() {
    updateCmd.Flags().BoolVarP(&forceUpdate, "force", "f", false, "Khong hoi xac nhan")
    rootCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
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

    executor := compose.NewExecutor(cwd)

    // Step 1: Pull new images
    fmt.Println("Dang kiem tra cap nhat...")
    spinner := ui.NewSpinner("Dang tai images...")
    spinner.Start()

    timeoutCtx, timeoutCancel := context.WithTimeout(ctx, compose.DefaultTimeout)
    defer timeoutCancel()

    output, err := executor.Pull(timeoutCtx)
    spinner.Stop(err == nil)

    if err != nil {
        return fmt.Errorf("khong tai duoc images: %w", err)
    }

    // Step 2: Parse pull output
    updates := updater.ParsePullOutput(output)

    if len(updates) == 0 {
        fmt.Println("\n[OK] Tat ca images da la phien ban moi nhat.")
        return nil
    }

    // Step 3: Show updates
    fmt.Println("\nCo cap nhat:")
    for _, u := range updates {
        fmt.Printf("  - %s\n", u.Image)
        if u.OldDigest != "" && u.NewDigest != "" {
            fmt.Printf("    %s -> %s\n", u.OldDigest[:12], u.NewDigest[:12])
        }
    }
    fmt.Println()

    // Step 4: Confirm restart
    if !forceUpdate {
        var confirm bool
        prompt := &survey.Confirm{
            Message: "Khoi dong lai services voi images moi?",
            Default: true,
        }
        survey.AskOne(prompt, &confirm)

        if !confirm {
            fmt.Println("Huy cap nhat. Images da duoc tai, chay 'kk restart' de ap dung.")
            return nil
        }
    }

    // Step 5: Recreate containers
    fmt.Println("Dang khoi dong lai voi images moi...")
    if err := executor.ForceRecreate(timeoutCtx); err != nil {
        return fmt.Errorf("recreate that bai: %w", err)
    }

    // Step 6: Monitor health
    composeFile, err := compose.ParseComposeFile(cwd)
    if err == nil {
        healthMonitor, err := monitor.NewHealthMonitor()
        if err == nil {
            defer healthMonitor.Close()

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

    fmt.Println("\n[OK] Cap nhat hoan tat!")

    // Show status
    statuses, err := monitor.GetStatus(timeoutCtx, executor)
    if err == nil {
        ui.PrintStatusTable(statuses)
    }

    return nil
}
```

**pkg/updater/updater.go:**
```go
package updater

import (
    "regexp"
    "strings"
)

type ImageUpdate struct {
    Image     string
    OldDigest string
    NewDigest string
    Updated   bool
}

// ParsePullOutput parses docker-compose pull output
// Example output lines:
//   Pulling db ... done
//   Pulling redis ... downloading
//   kkengine Pulled
//   Status: Downloaded newer image for mariadb:10.6
func ParsePullOutput(output string) []ImageUpdate {
    var updates []ImageUpdate

    // Pattern for "Downloaded newer image"
    newerPattern := regexp.MustCompile(`Downloaded newer image for (.+)`)

    // Pattern for digest changes
    digestPattern := regexp.MustCompile(`Digest: sha256:([a-f0-9]+)`)

    lines := strings.Split(output, "\n")
    currentImage := ""

    for _, line := range lines {
        line = strings.TrimSpace(line)

        // Check for "newer image" pattern
        if matches := newerPattern.FindStringSubmatch(line); len(matches) > 1 {
            updates = append(updates, ImageUpdate{
                Image:   matches[1],
                Updated: true,
            })
            continue
        }

        // Track current image being pulled
        if strings.HasPrefix(line, "Pulling ") {
            parts := strings.Fields(line)
            if len(parts) >= 2 {
                currentImage = parts[1]
            }
        }

        // Extract digest
        if matches := digestPattern.FindStringSubmatch(line); len(matches) > 1 {
            if currentImage != "" {
                // Look for existing update or create new
                found := false
                for i := range updates {
                    if updates[i].Image == currentImage {
                        updates[i].NewDigest = matches[1]
                        found = true
                        break
                    }
                }
                if !found && strings.Contains(output, "Downloaded") {
                    updates = append(updates, ImageUpdate{
                        Image:     currentImage,
                        NewDigest: matches[1],
                        Updated:   true,
                    })
                }
            }
        }
    }

    return updates
}
```

### Step 2: Shell Completions (1h)

**cmd/completion.go:**
```go
package cmd

import (
    "os"

    "github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
    Use:   "completion [bash|zsh|fish]",
    Short: "Tao shell completion script",
    Long: `Tao shell completion script cho bash, zsh, hoac fish.

Bash:
  $ source <(kk completion bash)
  # Hoac them vao ~/.bashrc:
  $ kk completion bash > /etc/bash_completion.d/kk

Zsh:
  $ source <(kk completion zsh)
  # Hoac them vao ~/.zshrc:
  $ kk completion zsh > "${fpath[1]}/_kk"

Fish:
  $ kk completion fish | source
  # Hoac luu vao:
  $ kk completion fish > ~/.config/fish/completions/kk.fish
`,
    ValidArgs:             []string{"bash", "zsh", "fish"},
    Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
    DisableFlagsInUseLine: true,
    RunE: func(cmd *cobra.Command, args []string) error {
        switch args[0] {
        case "bash":
            return rootCmd.GenBashCompletion(os.Stdout)
        case "zsh":
            return rootCmd.GenZshCompletion(os.Stdout)
        case "fish":
            return rootCmd.GenFishCompletion(os.Stdout, true)
        }
        return nil
    },
}

func init() {
    rootCmd.AddCommand(completionCmd)
}
```

### Step 3: Unit Tests (4h)

**pkg/validator/docker_test.go:**
```go
package validator

import (
    "testing"
)

func TestCheckDockerInstalled(t *testing.T) {
    err := CheckDockerInstalled()
    // This test will pass if Docker is installed on the machine
    // In CI, Docker should be available
    if err != nil {
        t.Logf("Docker not installed: %v", err)
    }
}

func TestUserError(t *testing.T) {
    err := &UserError{
        Key:        "test_error",
        Message:    "Test message",
        Suggestion: "Test suggestion",
    }

    if err.Error() != "Test message" {
        t.Errorf("Expected 'Test message', got '%s'", err.Error())
    }
}
```

**pkg/validator/ports_test.go:**
```go
package validator

import (
    "net"
    "testing"
)

func TestCheckPort_Available(t *testing.T) {
    // Use a high port that's likely available
    status := CheckPort(59999)
    if status.InUse {
        t.Skip("Port 59999 is in use, skipping test")
    }

    if status.InUse {
        t.Error("Expected port to be available")
    }
}

func TestCheckPort_InUse(t *testing.T) {
    // Start a listener on a random port
    listener, err := net.Listen("tcp", ":0")
    if err != nil {
        t.Fatalf("Failed to start listener: %v", err)
    }
    defer listener.Close()

    // Get the port
    port := listener.Addr().(*net.TCPAddr).Port

    // Check the port
    status := CheckPort(port)
    if !status.InUse {
        t.Error("Expected port to be in use")
    }
}

func TestFormatPortConflict(t *testing.T) {
    tests := []struct {
        name     string
        status   PortStatus
        expected string
    }{
        {
            name: "with PID and process",
            status: PortStatus{
                Port:    3307,
                InUse:   true,
                PID:     1234,
                Process: "mysqld",
            },
            expected: "  - Port 3307 (MariaDB): dang dung boi mysqld (PID 1234). Stop: sudo kill 1234",
        },
        {
            name: "with PID only",
            status: PortStatus{
                Port:  8019,
                InUse: true,
                PID:   5678,
            },
            expected: "  - Port 8019 (kkengine): dang dung boi PID 5678. Stop: sudo kill 5678",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := formatPortConflict("MariaDB", tt.status)
            if tt.name == "with PID and process" && result != tt.expected {
                // Just check it contains key info
                if !containsAll(result, "3307", "1234", "mysqld") {
                    t.Errorf("Expected result to contain port, PID, and process")
                }
            }
        })
    }
}

func containsAll(s string, substrs ...string) bool {
    for _, sub := range substrs {
        if !contains(s, sub) {
            return false
        }
    }
    return true
}

func contains(s, sub string) bool {
    return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsHelper(s, sub))
}

func containsHelper(s, sub string) bool {
    for i := 0; i <= len(s)-len(sub); i++ {
        if s[i:i+len(sub)] == sub {
            return true
        }
    }
    return false
}
```

**pkg/validator/env_test.go:**
```go
package validator

import (
    "os"
    "path/filepath"
    "testing"
)

func TestValidateEnvFile_Missing(t *testing.T) {
    tmpDir := t.TempDir()

    err := ValidateEnvFile(tmpDir)
    if err == nil {
        t.Error("Expected error for missing .env file")
    }

    ue, ok := err.(*UserError)
    if !ok {
        t.Error("Expected UserError type")
    }
    if ue.Key != "env_missing" {
        t.Errorf("Expected key 'env_missing', got '%s'", ue.Key)
    }
}

func TestValidateEnvFile_MissingVars(t *testing.T) {
    tmpDir := t.TempDir()
    envPath := filepath.Join(tmpDir, ".env")

    // Create .env with missing required vars
    content := []byte("DB_HOSTNAME=localhost\n")
    if err := os.WriteFile(envPath, content, 0644); err != nil {
        t.Fatal(err)
    }

    err := ValidateEnvFile(tmpDir)
    if err == nil {
        t.Error("Expected error for missing required vars")
    }

    ue, ok := err.(*UserError)
    if !ok {
        t.Error("Expected UserError type")
    }
    if ue.Key != "env_missing_vars" {
        t.Errorf("Expected key 'env_missing_vars', got '%s'", ue.Key)
    }
}

func TestValidateEnvFile_Valid(t *testing.T) {
    tmpDir := t.TempDir()
    envPath := filepath.Join(tmpDir, ".env")

    content := []byte(`
DB_PASSWORD=supersecretpassword123
DB_ROOT_PASSWORD=rootpassword12345
REDIS_PASSWORD=redispassword1234
`)
    if err := os.WriteFile(envPath, content, 0644); err != nil {
        t.Fatal(err)
    }

    err := ValidateEnvFile(tmpDir)
    if err != nil {
        t.Errorf("Expected no error, got: %v", err)
    }
}

func TestParseEnvFile(t *testing.T) {
    tmpDir := t.TempDir()
    envPath := filepath.Join(tmpDir, ".env")

    content := []byte(`
# Comment
DB_HOST=localhost
DB_PORT=3306
DB_PASSWORD="quoted value"
EMPTY=

# Another comment
REDIS_HOST=redis
`)
    if err := os.WriteFile(envPath, content, 0644); err != nil {
        t.Fatal(err)
    }

    vars, err := parseEnvFile(envPath)
    if err != nil {
        t.Fatal(err)
    }

    expected := map[string]string{
        "DB_HOST":     "localhost",
        "DB_PORT":     "3306",
        "DB_PASSWORD": "quoted value",
        "EMPTY":       "",
        "REDIS_HOST":  "redis",
    }

    for key, expectedVal := range expected {
        if vars[key] != expectedVal {
            t.Errorf("Expected %s='%s', got '%s'", key, expectedVal, vars[key])
        }
    }
}
```

### Step 4: Makefile (1h)

**Makefile:**
```makefile
.PHONY: build test clean install release

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-s -w -X github.com/kkengine/kkcli/cmd.Version=$(VERSION)"
BINARY := kk

# Build for current platform
build:
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BINARY) .

# Build for all platforms
build-all: clean
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-amd64 .
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-linux-arm64 .
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-amd64 .
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY)-darwin-arm64 .

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -f $(BINARY)
	rm -rf dist/
	rm -f coverage.out coverage.html

# Install locally
install: build
	sudo cp $(BINARY) /usr/local/bin/

# Uninstall
uninstall:
	sudo rm -f /usr/local/bin/$(BINARY)

# Run linter
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...

# Download dependencies
deps:
	go mod download
	go mod tidy
```

### Step 5: GoReleaser Config (1h)

**.goreleaser.yml:**
```yaml
project_name: kkcli

before:
  hooks:
    - go mod tidy

builds:
  - id: kk
    main: .
    binary: kk
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X github.com/kkengine/kkcli/cmd.Version={{.Version}}

archives:
  - id: default
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    format: tar.gz
    format_overrides:
      - goos: windows
        format: zip
    files:
      - README.md
      - LICENSE

checksum:
  name_template: 'checksums.txt'

snapshot:
  name_template: "{{ .Tag }}-next"

changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'
      - '^ci:'

release:
  github:
    owner: kkengine
    name: kkcli
  draft: false
  prerelease: auto
```

### Step 6: Install Script (1h)

**scripts/install.sh:**
```bash
#!/bin/bash
set -e

# KK CLI Installer
# Usage: curl -sSL https://get.kkengine.com/cli | bash

REPO="kkengine/kkcli"
BINARY="kk"
INSTALL_DIR="/usr/local/bin"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        echo "Kien truc khong ho tro: $ARCH"
        exit 1
        ;;
esac

# Get latest release
echo "Dang kiem tra phien ban moi nhat..."
LATEST=$(curl -sL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST" ]; then
    echo "Khong tim thay phien ban. Vui long kiem tra ket noi mang."
    exit 1
fi

echo "Phien ban moi nhat: $LATEST"

# Download URL
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$LATEST/kkcli_${LATEST#v}_${OS}_${ARCH}.tar.gz"

# Create temp directory
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

# Download and extract
echo "Dang tai tu: $DOWNLOAD_URL"
curl -sL "$DOWNLOAD_URL" | tar -xz -C "$TMP_DIR"

# Install
echo "Dang cai dat..."
if [ -w "$INSTALL_DIR" ]; then
    mv "$TMP_DIR/$BINARY" "$INSTALL_DIR/"
else
    sudo mv "$TMP_DIR/$BINARY" "$INSTALL_DIR/"
fi

chmod +x "$INSTALL_DIR/$BINARY"

# Verify
if command -v $BINARY &> /dev/null; then
    echo ""
    echo "Cai dat thanh cong!"
    echo ""
    $BINARY --version
    echo ""
    echo "Bat dau su dung: kk init"
else
    echo "Cai dat that bai. Vui long thu lai."
    exit 1
fi
```

### Step 7: GitHub Actions CI (1h)

**.github/workflows/ci.yml:**
```yaml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Download dependencies
        run: go mod download

      - name: Run tests
        run: go test -v ./...

      - name: Build
        run: CGO_ENABLED=0 go build -o kk .

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v4
        with:
          version: latest

  release:
    needs: [test, lint]
    if: startsWith(github.ref, 'refs/tags/')
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v5
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## Todo List

- [ ] Implement cmd/update.go
- [ ] Implement pkg/updater/updater.go
- [ ] Implement cmd/completion.go
- [ ] Write unit tests for validators
- [ ] Write unit tests for updater
- [ ] Create Makefile
- [ ] Create .goreleaser.yml
- [ ] Create scripts/install.sh
- [ ] Create .github/workflows/ci.yml
- [ ] Test build on all platforms
- [ ] Test install script
- [ ] Tag first release (v0.1.0)

## Success Criteria

1. `kk update` pulls new images and shows changes
2. Confirmation prompt before recreating
3. `kk completion bash/zsh` generates valid scripts
4. All tests pass (`go test ./...`)
5. Binaries build for linux/darwin (amd64/arm64)
6. Install script works on fresh Ubuntu
7. GitHub Actions CI passes

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| GoReleaser config issues | Low | Test locally first |
| Install script security | Medium | Use HTTPS, verify checksums |
| Test flakiness | Low | Use table-driven tests, mock externals |

## Security Considerations

1. **Install Script:** Always use HTTPS, consider adding checksum verification
2. **Releases:** Sign releases with GPG (future enhancement)
3. **CI:** Use pinned action versions, minimal permissions

## Distribution Checklist

1. [ ] Create GitHub repository
2. [ ] Push code to main branch
3. [ ] Tag v0.1.0: `git tag v0.1.0 && git push origin v0.1.0`
4. [ ] Verify GitHub Actions creates release
5. [ ] Test install script: `curl -sSL https://... | bash`
6. [ ] Update README with installation instructions

## README Template

```markdown
# KK CLI

CLI tool de quan ly kkengine Docker stack.

## Cai dat

```bash
curl -sSL https://get.kkengine.com/cli | bash
```

Hoac tai truc tiep tu [Releases](https://github.com/kkengine/kkcli/releases).

## Su dung

```bash
# Khoi tao project moi
kk init

# Khoi dong stack
kk start

# Xem trang thai
kk status

# Khoi dong lai
kk restart

# Cap nhat images
kk update
```

## Yeu cau

- Docker >= 20.10
- Docker Compose >= 2.0
- Linux hoac macOS (amd64/arm64)

## License

MIT
```

## Next Steps

After completing Phase 04:
1. Release v0.1.0
2. Monitor user feedback
3. Plan v0.2.0 features:
   - `kk logs` command
   - `kk down` command
   - `kk self-update` command
   - Windows support (optional)

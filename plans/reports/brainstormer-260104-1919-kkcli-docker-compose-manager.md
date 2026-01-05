# KK CLI - Docker Compose Management Tool

**Type:** Brainstorm Report
**Date:** 2026-01-04 19:19
**Agent:** brainstormer
**Status:** Final Recommendation

---

## Problem Statement

Cần tool CLI để giúp non-technical users quản lý kkengine Docker stack dễ dàng.

**Requirements:**
- Global binary installation (Go)
- Copy template configs, user manual edit
- Pull images từ Docker Hub public
- Comprehensive validation
- Target: non-technical users
- Platform: Linux, Cloud VPS

**Core Commands:**
- `kk init` - Initialize configs
- `kk start` - Start stack with monitoring
- `kk restart` - Restart services
- `kk update` - Update images
- `kk status` - Show service status

**Stack Components:**
- kkengine (required)
- MariaDB (required)
- Redis (required)
- SeaweedFS (optional)
- Caddy (optional)

---

## Evaluated Approaches

### ❌ Approach 1: Pure Docker-Compose Wrapper

**Concept:** Thin wrapper gọi docker-compose trực tiếp

**Pros:**
- Simplest implementation
- Minimal code
- Fast development

**Cons:**
- Cannot meet comprehensive validation requirement
- Technical Docker errors exposed to users
- Poor UX for non-technical users

**Verdict:** Rejected - không đáp ứng validation requirement

---

### ✅ Approach 2: Intelligent CLI with Pre-flight Validation (RECOMMENDED)

**Concept:** Validate trước → docker-compose → Monitor → User-friendly output

**Workflow:**
```
Command → Pre-flight checks → Docker Compose → Health monitor → Friendly output
```

**Pros:**
- Comprehensive validation như requirement
- Transform technical errors → plain language
- Catch issues before Docker runs
- Excellent UX for non-technical users
- Still follows KISS - không reinvent docker-compose

**Cons:**
- Medium development effort (validation logic)
- Need maintain validation when docker-compose updates

**Verdict:** RECOMMENDED - best balance UX vs complexity

---

### ❌ Approach 3: Full Orchestration Layer

**Concept:** Manage containers qua Docker SDK, không dùng docker-compose

**Pros:**
- Full control
- Custom logic

**Cons:**
- Over-engineering (YAGNI violation)
- Must reimplement docker-compose features
- High maintenance cost
- More bugs risk

**Verdict:** Rejected - over-engineered

---

## Final Solution: Intelligent CLI (Approach 2)

### Tech Stack

- **Language:** Go (single binary, zero deps)
- **CLI Framework:** Cobra (industry standard)
- **Config:** Viper (env vars handling)
- **Docker:** `os/exec` for compose + Docker SDK for validation
- **Templates:** embed + text/template

### Architecture

```
kkcli/
├── cmd/                  # Commands
│   ├── root.go
│   ├── init.go
│   ├── start.go
│   ├── restart.go
│   ├── update.go
│   └── status.go
├── pkg/
│   ├── validator/        # Pre-flight validation
│   │   ├── docker.go     # Docker checks
│   │   ├── ports.go      # Port conflicts
│   │   ├── env.go        # Env validation
│   │   └── config.go     # Config syntax
│   ├── compose/          # Compose wrapper
│   │   ├── executor.go
│   │   └── parser.go
│   ├── monitor/          # Health monitoring
│   │   └── health.go
│   ├── ui/               # User-friendly output
│   │   ├── messages.go   # Error translation
│   │   └── progress.go
│   └── templates/        # Embedded templates
│       └── embed.go
└── templates/            # Template files
    ├── docker-compose.yml.tmpl
    ├── .env.tmpl
    ├── Caddyfile.tmpl
    ├── kkfiler.toml.tmpl
    └── kkphp.conf.tmpl
```

### Command Workflows

#### `kk init`

```
1. Check Docker installed & running
2. Detect working directory
3. Check if initialized (docker-compose.yml exists)
4. Interactive prompts:
   - Enable SeaweedFS? [y/N]
   - Enable Caddy web server? [y/N]
5. Copy template files based on selection
6. Generate random passwords (DB, Redis)
7. Write .env
8. Show success + next steps
```

**Interactive UX:**
```
🔍 Checking Docker... ✓
📁 Initializing in: /path/to/project

Select services to enable:
? Enable SeaweedFS file storage? [y/N]: n
? Enable Caddy web server? [y/N]: y

✓ Created docker-compose.yml
✓ Created .env (with generated passwords)
✓ Created Caddyfile
✓ Created kkphp.conf

Next steps:
  1. Review and edit .env if needed
  2. Run: kk start
```

#### `kk start`

```
1. Validate docker-compose.yml exists
2. Validate .env complete
3. Check port conflicts (3307, 8019, 80, 443)
4. Check disk space > 5GB
5. docker-compose up -d
6. Monitor health checks (auto retry 3 times if fail)
7. Display service status table + URLs
```

**Output:**
```
🔍 Pre-flight checks...
  ✓ Docker daemon running
  ✓ Ports available: 3307, 8019, 80, 443
  ✓ Environment variables complete
  ✓ Disk space: 24GB available

🚀 Starting services...
  ⏳ MariaDB starting... ✓ healthy
  ⏳ Redis starting... ✓ healthy
  ⏳ kkengine starting... ✓ healthy
  ⏳ Caddy starting... ✓ running

✅ All services running!

Services:
┌──────────┬──────────┬─────────────────────┐
│ Service  │ Status   │ Access              │
├──────────┼──────────┼─────────────────────┤
│ kkengine │ healthy  │ http://localhost:8019 │
│ MariaDB  │ healthy  │ localhost:3307      │
│ Redis    │ running  │ -                   │
│ Caddy    │ running  │ http://localhost    │
└──────────┴──────────┴─────────────────────┘
```

#### `kk status`

```
1. docker-compose ps
2. Parse output
3. Display formatted table
```

#### `kk restart`

```
1. docker-compose restart
2. Monitor health (auto retry 3 times)
3. Display status
```

#### `kk update`

```
1. docker-compose pull
2. Show updated images
3. Ask confirmation to restart
4. If yes: docker-compose up -d --force-recreate
5. Monitor health
```

**Output:**
```
🔄 Checking for updates...

Updates available:
  - kkengine:latest (current: abc123, new: def456)
  - mariadb:10.6 (current: xyz789, new: uvw012)

? Restart services with new images? [Y/n]: y

🚀 Recreating services...
  ✓ Services updated successfully
```

### Validation Matrix

| Check | Action if Fail | User Message |
|-------|----------------|--------------|
| Docker installed | Block | "Docker chưa cài. Cài tại: https://docs.docker.com/get-docker/" |
| Docker daemon running | Block | "Docker daemon không chạy. Chạy: sudo systemctl start docker" |
| Port 3307 conflict | Block | "Port 3307 đã dùng bởi PID X. Stop process: sudo kill X" |
| Port 8019 conflict | Block | "Port 8019 đã dùng..." |
| Port 80 conflict (Caddy) | Block | "Port 80 đã dùng..." |
| Port 443 conflict (Caddy) | Block | "Port 443 đã dùng..." |
| .env missing | Block | "File .env không tồn tại. Chạy: kk init" |
| DB_PASSWORD missing | Block | "Thiếu DB_PASSWORD trong .env" |
| REDIS_PASSWORD missing | Block | "Thiếu REDIS_PASSWORD trong .env" |
| Disk < 5GB | Warning | "⚠️  Disk space thấp (XGB). Recommend ít nhất 5GB" |
| Health check fail | Auto retry 3x | "Service X unhealthy, retrying (1/3)..." |

### Error Translation System

```go
type UserFriendlyError struct {
    TechnicalError error
    UserMessage    string
    Suggestion     string
    DocsURL        string
}

// Examples:
"port is already allocated" →
  Message: "Port đã được sử dụng"
  Suggestion: "Kiểm tra: sudo lsof -i :PORT"
  Docs: "https://docs.kkengine.com/troubleshooting/ports"

"connection refused" →
  Message: "Không thể kết nối tới Docker daemon"
  Suggestion: "Chạy: sudo systemctl start docker"
  Docs: "https://docs.kkengine.com/troubleshooting/docker"
```

### Implementation Phases

**Phase 1: Core Foundation (1 week)**
- Setup Cobra project structure
- Implement `kk init` with interactive prompts
- Template embedding system
- .env generation with random passwords
- Basic validation (Docker check)

**Phase 2: Validation Layer (1 week)**
- Port conflict detection
- Env vars validation
- Config syntax validation
- Error translation framework
- User-friendly messages

**Phase 3: Operations (1 week)**
- `kk start` with monitoring
- Health check system with auto-retry
- `kk status` with formatted output
- `kk restart`
- Progress indicators

**Phase 4: Advanced Features (1 week)**
- `kk update` command
- Image pull tracking
- Comprehensive error messages
- Documentation
- Testing

**Total:** 4 weeks

---

## Implementation Considerations

### Security
- Generate cryptographically random passwords
- Never log sensitive data (passwords, tokens)
- Validate .env permissions (warn if world-readable)

### Robustness
- Handle SIGINT/SIGTERM gracefully
- Timeout for long operations
- Disk space checks before pulling images
- Network connectivity check before pull

### Extensibility
- Plugin system for future commands?
- Config file for CLI settings (~/.kkcli.yaml)?
- Not needed now (YAGNI), but architecture allows

### Distribution
- Build: `CGO_ENABLED=0 go build -ldflags="-s -w"`
- Release: GitHub Releases with binaries
- Install: `curl -sSL https://get.kkengine.com/cli | bash`
- Update: `kk self-update` (future)

---

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Docker Compose version incompatibility | High | Detect version, warn if < 2.0 |
| Port conflicts không detect được | Medium | Check multiple ways: lsof, netstat, Docker API |
| Health checks false positives | Medium | Configurable retry count + timeout |
| Template rendering errors | Low | Extensive testing, validation before write |
| User modifies files incorrectly | Medium | Validate syntax before running |

---

## Success Metrics

**Primary:**
- User có thể init + start stack trong < 2 phút
- Zero Docker knowledge required
- 90% errors có friendly message + suggestion

**Secondary:**
- < 5s startup time cho CLI
- Binary size < 10MB
- Support Linux kernel >= 4.0

---

## Next Steps

1. **Technical Decisions Finalized:**
   - ✅ Go + Cobra framework
   - ✅ Interactive service selection
   - ✅ Single .env only
   - ✅ Auto-retry 3 times for health checks

2. **Create Implementation Plan:**
   - Detailed task breakdown
   - File structure
   - Code architecture diagrams
   - Testing strategy

3. **Setup Development:**
   - Initialize Go module
   - Setup Cobra boilerplate
   - Create template files
   - Setup CI/CD for releases

---

## Unresolved Questions

None - all critical decisions finalized.

---

## Sources

Research findings based on:
- Go CLI best practices and Cobra framework patterns
- Docker-compose wrapper implementation strategies
- Validation approaches for CLI tools targeting non-technical users
- Health check monitoring patterns

---

**RECOMMENDATION:** Proceed with Approach 2 (Intelligent CLI with Pre-flight Validation). Solution meets all requirements while following YAGNI, KISS, DRY principles.

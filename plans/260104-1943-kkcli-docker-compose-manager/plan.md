---
title: "KK CLI - Docker Compose Management Tool"
description: "Global binary CLI (Go + Cobra) for non-technical users to manage kkengine Docker stack"
status: completed
priority: P1
effort: 4w
branch: initial-2
tags: [go, cli, docker, cobra, devops]
created: 2026-01-04
---

# KK CLI - Docker Compose Management Tool

## Overview

CLI tool giup non-technical users quan ly kkengine Docker stack. Commands: init, start, status, restart, update. Target: Linux/VPS.

## Tech Stack

- **Language:** Go 1.21+ (static binary, CGO_ENABLED=0)
- **CLI Framework:** Cobra + survey/promptui (interactive)
- **Docker:** os/exec (compose) + Docker SDK (validation)
- **Templates:** go:embed + text/template
- **Progress:** pterm hoac spinner

## Architecture

```
kkcli/
├── cmd/           # Commands: root, init, start, restart, update, status
├── pkg/
│   ├── validator/ # Docker, ports, env, config validation
│   ├── compose/   # Executor, parser
│   ├── monitor/   # Health checks + auto-retry
│   ├── ui/        # Messages (Vietnamese), progress
│   └── templates/ # Embed logic
└── templates/     # *.tmpl files (docker-compose, .env, Caddyfile...)
```

## Phases

| Phase | Name | Effort | Status | File |
|-------|------|--------|--------|------|
| 01 | Core Foundation | 1w | done | [phase-01-core-foundation.md](./phase-01-core-foundation.md) |
| 02 | Validation Layer | 1w | done (2026-01-04) | [phase-02-validation-layer.md](./phase-02-validation-layer.md) |
| 03 | Operations | 1w | done (2026-01-05) | [phase-03-operations.md](./phase-03-operations.md) |
| 04 | Advanced Features | 1w | done (2026-01-05) | [phase-04-advanced-features.md](./phase-04-advanced-features.md) |

## Key Features

- **kk init:** Interactive service selection (SeaweedFS, Caddy optional), template generation, auto password generation
- **kk start:** Pre-flight validation, docker-compose up, health monitoring with 3x auto-retry
- **kk status:** Formatted service status table
- **kk restart:** Graceful restart with health monitoring
- **kk update:** Pull new images, confirmation, recreate containers

## Validation Matrix

| Check | Block/Warn | Vi Message |
|-------|------------|------------|
| Docker installed | Block | "Docker chua cai. Cai tai: https://docs.docker.com/get-docker/" |
| Docker daemon | Block | "Docker daemon khong chay. Chay: sudo systemctl start docker" |
| Port conflict | Block | "Port X da dung boi PID Y. Stop: sudo kill Y" |
| .env missing | Block | "File .env khong ton tai. Chay: kk init" |
| Disk < 5GB | Warn | "Disk space thap (XGB). Recommend it nhat 5GB" |

## Success Metrics

- User init + start trong < 2 phut
- Zero Docker knowledge required
- 90% errors co friendly message + suggestion
- Binary size < 10MB
- < 5s CLI startup time

## Distribution

- Build: `CGO_ENABLED=0 go build -ldflags="-s -w"`
- Release: GitHub Releases with binaries (linux-amd64, linux-arm64, darwin-amd64, darwin-arm64)
- Install: `curl -sSL https://get.kkengine.com/cli | bash`

## Validation Summary

**Validated:** 2026-01-04
**Questions asked:** 8

### Confirmed Decisions

1. **Interactive UI Library:** huh (bubbletea) - Modern TUI framework thay vi survey
2. **Progress Indicators:** pterm - Full-featured cho rich CLI UX
3. **Platform Support:** Linux only (amd64 + arm64) - Target users chu yeu VPS/Linux
4. **Health Check Retry:** Fixed intervals (1s, 2s, 4s) - Don gian, du doan
5. **.env Permissions:** Warn only, don't modify - Giu user control
6. **Compose Version:** Require v2.0+ - Modern standard
7. **Config Backup:** Auto backup to .bak - Safe cho user khi overwrite
8. **Compose Command:** Try v2 (docker compose), fallback v1 (docker-compose) - Best compatibility

### Action Items

- [x] Update Phase 01: Replace survey/promptui with huh (bubbletea) in code examples (2026-01-05)
- [x] Update Phase 01: Add .bak backup logic to init command (2026-01-05)
- [x] Update Phase 02: Add .env permission warning (not auto-fix) (2026-01-05)
- [x] Update Phase 02: Add Compose version check (require v2.0+) (2026-01-05)
- [x] Update Phase 03: Implement compose command detection (v2 fallback v1) (2026-01-05)
- [x] Update Distribution: Remove darwin-* targets, keep only linux-amd64 and linux-arm64 (2026-01-05)

## Related Docs

- [Brainstorm Report](../reports/brainstormer-260104-1919-kkcli-docker-compose-manager.md)
- [Research: Go CLI Ecosystem](./research/researcher-01-go-cli-ecosystem.md)
- [Research: Docker Integration](./research/researcher-02-docker-integration.md)

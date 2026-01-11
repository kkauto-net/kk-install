---
title: "Redesign kk init Command"
description: "Improve UX for init flow - clearer service options, customizable credentials with grouped form"
status: in-progress
priority: P2
effort: 4h
branch: main
tags: [cli, ux, init, templates]
created: 2026-01-11
---

# Plan: Redesign `kk init` Command

## Overview

Cải thiện UX cho `kk init` flow:
- Rõ ràng hơn về services (MariaDB/Redis mặc định ON, chỉ hỏi SeaweedFS/Caddy)
- Cho phép user customize credentials thay vì auto-generate
- Thêm JWT_SECRET và S3 keys vào config

## Current State

| File | Issue |
|------|-------|
| `cmd/init.go` | Flow 5 steps, credentials auto-gen không hiển thị |
| `pkg/templates/embed.go` | Config thiếu JWTSecret, S3AccessKey, S3SecretKey |
| `pkg/templates/env.tmpl` | S3 keys hardcode, thiếu JWT_SECRET |
| `pkg/ui/lang_*.go` | Thiếu messages cho grouped form |

## Target State

**New Flow (6 steps):**
1. Docker Check (giữ nguyên)
2. Language Selection (giữ nguyên)
3. Service Selection - chỉ hỏi SeaweedFS, Caddy
4. **Domain Configuration** - hỏi SYSTEM_DOMAIN
5. **Environment Configuration** - confirm random? -> conditional grouped form
6. Generate Files + Complete Summary

## Implementation Phases

| Phase | Description | Effort |
|-------|-------------|--------|
| [Phase 01](./phase-01-update-templates.md) | Update templates (embed.go, env.tmpl) | 30m |
| [Phase 02](./phase-02-refactor-init-flow.md) | Refactor cmd/init.go với new flow | 2h |
| [Phase 03](./phase-03-add-ui-messages.md) | Add UI messages (lang_en, lang_vi) | 30m |
| [Phase 04](./phase-04-update-tests.md) | Update tests | 1h |

## Files to Modify

```
pkg/templates/embed.go      # Add JWTSecret, S3AccessKey, S3SecretKey to Config
pkg/templates/env.tmpl      # Add JWT_SECRET, replace hardcoded S3 keys
cmd/init.go                 # Refactor flow: domain step, confirm random, grouped form
pkg/ui/lang_en.go           # Add new messages
pkg/ui/lang_vi.go           # Add new messages (Vietnamese)
pkg/templates/embed_test.go # Update tests for new Config fields
cmd/init_test.go            # Update tests if exists
```

## Key Implementation Details

### 1. Config Struct Changes

```go
type Config struct {
    EnableSeaweedFS bool
    EnableCaddy     bool
    Domain          string
    JWTSecret       string      // NEW
    DBPassword      string
    DBRootPassword  string
    RedisPassword   string
    S3AccessKey     string      // NEW
    S3SecretKey     string      // NEW
}
```

### 2. New Flow Logic

```
askDomain() -> SYSTEM_DOMAIN
generateRandomSecrets() -> pre-fill all
confirmUseRandom?
  -> Yes: use pre-filled
  -> No:  show grouped form with pre-filled values
generateFiles(config)
```

### 3. Grouped Form Structure

```
[System Configuration]
  - JWT_SECRET

[Database Secrets]
  - DB_PASSWORD
  - DB_ROOT_PASSWORD
  - REDIS_PASSWORD

[S3 Storage Secrets] (only if EnableSeaweedFS)
  - S3_ACCESS_KEY
  - S3_SECRET_KEY
```

## Validation Rules

| Field | Format | Length |
|-------|--------|--------|
| JWT_SECRET | base64-safe alphanumeric | 32 chars |
| DB_PASSWORD | alphanumeric | 24 chars |
| S3_ACCESS_KEY | alphanumeric uppercase | 20 chars |
| S3_SECRET_KEY | alphanumeric | 40 chars |

## Success Criteria

- [ ] `kk init` runs with new 6-step flow
- [ ] .env contains JWT_SECRET and dynamic S3 keys
- [ ] User can edit credentials via grouped form
- [ ] docker-compose.yml reads correctly from .env
- [ ] All tests pass
- [ ] Manual testing OK

## Dependencies

- `charmbracelet/huh` - already in use for forms
- `ui.GeneratePassword()` - already exists, reuse

## Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| User leaves field empty | Validate non-empty, show warning |
| S3 keys invalid format | Generate with correct format, validate on input |
| Existing tests break | Update test cases for new Config struct |

## Unresolved Questions

None - all clarified in brainstorm.

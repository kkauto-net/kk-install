# Brainstorm: Professional CLI Output Format

**Date:** 2026-01-10
**Status:** Completed
**Scope:** Both help text and runtime output

---

## Problem Statement

Current `kk` CLI has unprofessional output:
- Vietnamese messages without diacritics ("Khoi tao" instead of "Khởi tạo")
- Flat command list without logical grouping
- Inconsistent messaging style
- No persistent language preference

## Requirements

1. GitHub CLI style help format (UPPERCASE headers, grouped commands)
2. Vietnamese with full diacritics when selected
3. English as default language
4. User language preference stored persistently
5. Keep emojis and full colors
6. Use Cobra built-in template customization

## Evaluated Approaches

### 1. Custom Help Templates (Chosen)
**Pros:** Full control, no new deps, Cobra native support
**Cons:** Manual template maintenance

### 2. Switch CLI Framework (Rejected)
**Pros:** Some frameworks have better help formatting
**Cons:** Major rewrite, breaking changes, unnecessary complexity

### 3. External Help Generator (Rejected)
**Pros:** Consistent with other tools
**Cons:** Added dependency, over-engineering

## Recommended Solution

### Help Template (GitHub CLI Style)

```
🚀 Manage your kkengine Docker stack effortlessly.

USAGE
  kk <command> [flags]

CORE COMMANDS
  init:       Initialize Docker stack with interactive setup
  start:      Start all services with preflight checks
  status:     View service status and health

MANAGEMENT COMMANDS
  restart:    Restart all services
  update:     Pull latest images and recreate containers

ADDITIONAL COMMANDS
  completion: Generate shell completion scripts

FLAGS
  -h, --help      Show help for command
  -v, --version   Show version

LEARN MORE
  Use 'kk <command> --help' for more information
```

### Command Grouping

| Group | Commands |
|-------|----------|
| core | init, start, status |
| management | restart, update |
| additional | completion, help |

### Language Files Update

**Vietnamese (lang_vi.go) - With Diacritics:**
```go
"checking_docker":      "Đang kiểm tra Docker...",
"docker_ok":            "Docker đã sẵn sàng",
"docker_not_installed": "Docker chưa được cài đặt",
"docker_not_running":   "Docker daemon không chạy",
"init_in_dir":          "Khởi tạo trong: %s",
"compose_exists":       "docker-compose.yml đã tồn tại. Ghi đè?",
"init_cancelled":       "Hủy khởi tạo",
"enable_seaweedfs":     "Bật SeaweedFS file storage?",
"seaweedfs_desc":       "SeaweedFS là hệ thống lưu trữ file phân tán",
"enable_caddy":         "Bật Caddy web server?",
"caddy_desc":           "Caddy là reverse proxy với tự động HTTPS",
"enter_domain":         "Nhập domain (vd: example.com):",
"generating_files":     "Đang tạo các file cấu hình...",
"files_generated":      "Các file cấu hình đã được tạo",
"created":              "Đã tạo: %s",
"init_complete":        "Khởi tạo hoàn tất!",
"next_steps_box":       "Bước tiếp theo:\n  1. Kiểm tra và chỉnh sửa .env nếu cần\n  2. Chạy: kk start",
```

### Config Storage

```yaml
# ~/.kk/config.yaml
language: vi  # or "en"
```

### Files to Modify

| File | Action | Description |
|------|--------|-------------|
| `pkg/ui/help.go` | CREATE | Custom help/usage templates |
| `pkg/ui/lang_vi.go` | MODIFY | Add Vietnamese diacritics |
| `pkg/ui/lang_en.go` | MODIFY | Polish English messages |
| `pkg/config/config.go` | CREATE | Language preference storage |
| `cmd/root.go` | MODIFY | Apply custom templates, load config |
| `cmd/init.go` | MODIFY | Add group annotation |
| `cmd/start.go` | MODIFY | Add group annotation |
| `cmd/status.go` | MODIFY | Add group annotation |
| `cmd/restart.go` | MODIFY | Add group annotation |
| `cmd/update.go` | MODIFY | Add group annotation |

## Implementation Considerations

1. **Backward Compatibility:** No breaking changes
2. **Config Migration:** Auto-create ~/.kk/ if not exists
3. **Locale Detection:** Optional - can detect system locale as fallback
4. **Testing:** Add tests for help output format

## Success Metrics

- [ ] Help output matches GitHub CLI style
- [ ] Vietnamese displays with full diacritics
- [ ] Language preference persists between sessions
- [ ] All commands properly grouped

## Risks

| Risk | Mitigation |
|------|------------|
| Template syntax errors | Unit tests for help output |
| Unicode display issues | Test on various terminals |
| Config file permissions | Handle errors gracefully |

## Next Steps

1. Create `pkg/ui/help.go` with custom templates
2. Update lang_vi.go with diacritics
3. Create `pkg/config/config.go` for preferences
4. Add annotations to all commands
5. Apply templates in root.go
6. Test on multiple terminals

---

## Unresolved Questions

None - all requirements clarified.

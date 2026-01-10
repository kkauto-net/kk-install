---
phase: 03
title: Update Language Files
status: pending
effort: 30m
---

# Phase 03: Language Files Update

## Context

- Parent plan: [plan.md](plan.md)
- Dependencies: None (can run in parallel with Phase 01-02)

## Overview

Update Vietnamese messages to include full diacritics. Polish English messages for consistency.

## Requirements

1. Vietnamese with proper Unicode diacritics
2. Consistent messaging style across both languages
3. Add new keys for help template messages

## Implementation Steps

### 1. Update pkg/ui/lang_vi.go

```go
package ui

var messagesVI = map[string]string{
    // Docker validation
    "checking_docker":      "Đang kiểm tra Docker...",
    "docker_ok":            "Docker đã sẵn sàng",
    "docker_not_installed": "Docker chưa được cài đặt",
    "docker_not_running":   "Docker daemon không chạy",

    // Init flow
    "init_in_dir":    "Khởi tạo trong: %s",
    "compose_exists": "docker-compose.yml đã tồn tại. Ghi đè?",
    "init_cancelled": "Hủy khởi tạo",

    // Prompts
    "enable_seaweedfs": "Bật SeaweedFS file storage?",
    "seaweedfs_desc":   "SeaweedFS là hệ thống lưu trữ file phân tán",
    "enable_caddy":     "Bật Caddy web server?",
    "caddy_desc":       "Caddy là reverse proxy với tự động HTTPS",
    "enter_domain":     "Nhập domain (vd: example.com):",
    "yes_recommended":  "Có (khuyến nghị)",
    "no":               "Không",

    // Errors
    "error_db_password":  "Không thể tạo mật khẩu DB",
    "error_db_root_pass": "Không thể tạo mật khẩu DB root",
    "error_redis_pass":   "Không thể tạo mật khẩu Redis",
    "error_create_file":  "Lỗi khi tạo file",

    // File generation
    "generating_files": "Đang tạo các file cấu hình...",
    "files_generated":  "Các file cấu hình đã được tạo",

    // Success
    "created":       "Đã tạo: %s",
    "init_complete": "Khởi tạo hoàn tất!",

    // Next steps
    "next_steps": `
Bước tiếp theo:
  1. Kiểm tra và chỉnh sửa .env nếu cần
  2. Chạy: kk start
`,
    "next_steps_box": `Bước tiếp theo:
  1. Kiểm tra và chỉnh sửa .env nếu cần
  2. Chạy: kk start`,

    // Language selection
    "select_language": "Chọn ngôn ngữ / Select language",
    "lang_english":    "English",
    "lang_vietnamese": "Tiếng Việt",

    // Runtime messages (start, restart, update, status)
    "stopping":           "Đang dừng lại...",
    "preflight_checking": "Kiểm tra trước khi chạy...",
    "preflight_failed":   "Kiểm tra thất bại. Vui lòng sửa lỗi trên",
    "starting_services":  "Khởi động services...",
    "start_failed":       "Khởi động thất bại",
    "health_checking":    "Đang kiểm tra sức khỏe dịch vụ...",
    "health_failed":      "Không thể theo dõi health",
    "some_not_ready":     "Một số dịch vụ chưa sẵn sàng. Kiểm tra: kk status",
    "start_complete":     "Khởi động hoàn tất!",
    "restarting":         "Đang khởi động lại dịch vụ...",
    "restart_failed":     "Khởi động lại thất bại",
    "restart_complete":   "Đã khởi động lại",
    "checking_updates":   "Đang kiểm tra cập nhật...",
    "pulling_images":     "Đang tải images...",
    "pull_failed":        "Không tải được images",
    "images_up_to_date":  "Tất cả images đã là phiên bản mới nhất",
    "updates_available":  "Có cập nhật:",
    "confirm_restart":    "Khởi động lại services với images mới?",
    "update_cancelled":   "Hủy cập nhật. Images đã được tải, chạy 'kk restart' để áp dụng",
    "recreating":         "Đang khởi động lại với images mới...",
    "recreate_failed":    "Recreate thất bại",
    "update_complete":    "Cập nhật hoàn tất!",
    "no_services":        "Không có dịch vụ nào đang chạy",
    "run_start":          "Chạy: kk start",
    "all_running":        "Tất cả %d dịch vụ đang chạy",
    "some_running":       "%d/%d dịch vụ đang chạy",
}
```

### 2. Update pkg/ui/lang_en.go

```go
package ui

var messagesEN = map[string]string{
    // Docker validation
    "checking_docker":      "Checking Docker...",
    "docker_ok":            "Docker is ready",
    "docker_not_installed": "Docker is not installed",
    "docker_not_running":   "Docker daemon is not running",

    // Init flow
    "init_in_dir":    "Initializing in: %s",
    "compose_exists": "docker-compose.yml already exists. Overwrite?",
    "init_cancelled": "Initialization cancelled",

    // Prompts
    "enable_seaweedfs": "Enable SeaweedFS file storage?",
    "seaweedfs_desc":   "SeaweedFS is a distributed file storage system",
    "enable_caddy":     "Enable Caddy web server?",
    "caddy_desc":       "Caddy is a reverse proxy with automatic HTTPS",
    "enter_domain":     "Enter domain (e.g., example.com):",
    "yes_recommended":  "Yes (recommended)",
    "no":               "No",

    // Errors
    "error_db_password":  "Failed to generate DB password",
    "error_db_root_pass": "Failed to generate DB root password",
    "error_redis_pass":   "Failed to generate Redis password",
    "error_create_file":  "Failed to create file",

    // File generation
    "generating_files": "Generating configuration files...",
    "files_generated":  "Configuration files generated",

    // Success
    "created":       "Created: %s",
    "init_complete": "Initialization complete!",

    // Next steps
    "next_steps": `
Next steps:
  1. Review and edit .env if needed
  2. Run: kk start
`,
    "next_steps_box": `Next steps:
  1. Review and edit .env if needed
  2. Run: kk start`,

    // Language selection
    "select_language": "Select language / Chọn ngôn ngữ",
    "lang_english":    "English",
    "lang_vietnamese": "Tiếng Việt",

    // Runtime messages
    "stopping":           "Stopping...",
    "preflight_checking": "Running preflight checks...",
    "preflight_failed":   "Preflight checks failed. Please fix the errors above",
    "starting_services":  "Starting services...",
    "start_failed":       "Start failed",
    "health_checking":    "Checking service health...",
    "health_failed":      "Cannot monitor health",
    "some_not_ready":     "Some services not ready. Check: kk status",
    "start_complete":     "Start complete!",
    "restarting":         "Restarting services...",
    "restart_failed":     "Restart failed",
    "restart_complete":   "Restart complete",
    "checking_updates":   "Checking for updates...",
    "pulling_images":     "Pulling images...",
    "pull_failed":        "Failed to pull images",
    "images_up_to_date":  "All images are up to date",
    "updates_available":  "Updates available:",
    "confirm_restart":    "Restart services with new images?",
    "update_cancelled":   "Update cancelled. Images downloaded, run 'kk restart' to apply",
    "recreating":         "Recreating with new images...",
    "recreate_failed":    "Recreate failed",
    "update_complete":    "Update complete!",
    "no_services":        "No services running",
    "run_start":          "Run: kk start",
    "all_running":        "All %d services running",
    "some_running":       "%d/%d services running",
}
```

### 3. Update cmd/*.go files to use i18n

Replace hardcoded strings in cmd files with `ui.Msg()` calls.

Example in `cmd/start.go`:
```go
// Before
fmt.Println("Dang dung lai...")

// After
fmt.Println(ui.Msg("stopping"))
```

## Todo List

- [ ] Update lang_vi.go with diacritics
- [ ] Update lang_en.go with new keys
- [ ] Update cmd/start.go to use ui.Msg()
- [ ] Update cmd/restart.go to use ui.Msg()
- [ ] Update cmd/update.go to use ui.Msg()
- [ ] Update cmd/status.go to use ui.Msg()

## Success Criteria

- [ ] All Vietnamese messages display with diacritics
- [ ] All runtime messages use i18n system
- [ ] No hardcoded Vietnamese/English strings in cmd/

## Files Changed

| File | Action |
|------|--------|
| `pkg/ui/lang_vi.go` | MODIFY |
| `pkg/ui/lang_en.go` | MODIFY |
| `cmd/start.go` | MODIFY |
| `cmd/restart.go` | MODIFY |
| `cmd/update.go` | MODIFY |
| `cmd/status.go` | MODIFY |

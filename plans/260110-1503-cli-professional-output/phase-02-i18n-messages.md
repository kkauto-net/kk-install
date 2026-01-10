# Phase 02: i18n Messages

## Context Links
- Parent: [plan.md](./plan.md)
- Depends on: [Phase 01](./phase-01-ui-components.md)

## Overview
- **Priority**: High
- **Status**: Pending
- **Description**: Add i18n message keys for new UI components

## Key Insights
- Existing i18n uses `map[string]string` in `lang_en.go` and `lang_vi.go`
- `Msg(key)` function retrieves messages
- Vietnamese should have proper diacritics
- Default language is English

## Requirements

### New Keys Needed

| Key | EN | VI | Used By |
|-----|----|----|---------|
| `service_status` | Service Status | Trạng thái dịch vụ | PrintStatusTable |
| `access_info` | Access Information | Thông tin truy cập | PrintAccessInfo |
| `col_service` | Service | Dịch vụ | Tables |
| `col_status` | Status | Trạng thái | Tables |
| `col_health` | Health | Sức khỏe | Tables |
| `col_ports` | Ports | Cổng | Tables |
| `col_url` | URL | URL | Tables |
| `col_setting` | Setting | Cài đặt | Summary |
| `col_value` | Value | Giá trị | Summary |
| `config_summary` | Configuration Summary | Tóm tắt cấu hình | PrintInitSummary |
| `created_files` | Created Files | Các file đã tạo | PrintInitSummary |
| `enabled` | Enabled | Bật | boolToStatus |
| `disabled` | Disabled | Tắt | boolToStatus |
| `domain` | Domain | Tên miền | Summary |
| `step_docker_check` | Docker Check | Kiểm tra Docker | Init wizard |
| `step_language` | Language Selection | Chọn ngôn ngữ | Init wizard |
| `step_options` | Configuration Options | Tùy chọn cấu hình | Init wizard |
| `step_generate` | Generate Files | Tạo file | Init wizard |
| `step_complete` | Complete | Hoàn tất | Init wizard |
| `check` | Check | Kiểm tra | Preflight |
| `result` | Result | Kết quả | Preflight |

## Related Code Files

| File | Action | Description |
|------|--------|-------------|
| `pkg/ui/lang_en.go` | Modify | Add English messages |
| `pkg/ui/lang_vi.go` | Modify | Add Vietnamese messages |

## Implementation Steps

### 1. Update `pkg/ui/lang_en.go`

Add to `messagesEN` map:

```go
// Table columns
"service_status":   "Service Status",
"access_info":      "Access Information",
"col_service":      "Service",
"col_status":       "Status",
"col_health":       "Health",
"col_ports":        "Ports",
"col_url":          "URL",
"col_setting":      "Setting",
"col_value":        "Value",

// Init summary
"config_summary":   "Configuration Summary",
"created_files":    "Created Files",
"enabled":          "Enabled",
"disabled":         "Disabled",
"domain":           "Domain",

// Init wizard steps
"step_docker_check": "Docker Check",
"step_language":     "Language Selection",
"step_options":      "Configuration Options",
"step_generate":     "Generate Files",
"step_complete":     "Complete",

// Preflight
"check":             "Check",
"result":            "Result",
```

### 2. Update `pkg/ui/lang_vi.go`

Add to `messagesVI` map:

```go
// Table columns
"service_status":   "Trạng thái dịch vụ",
"access_info":      "Thông tin truy cập",
"col_service":      "Dịch vụ",
"col_status":       "Trạng thái",
"col_health":       "Sức khỏe",
"col_ports":        "Cổng",
"col_url":          "URL",
"col_setting":      "Cài đặt",
"col_value":        "Giá trị",

// Init summary
"config_summary":   "Tóm tắt cấu hình",
"created_files":    "Các file đã tạo",
"enabled":          "Bật",
"disabled":         "Tắt",
"domain":           "Tên miền",

// Init wizard steps
"step_docker_check": "Kiểm tra Docker",
"step_language":     "Chọn ngôn ngữ",
"step_options":      "Tùy chọn cấu hình",
"step_generate":     "Tạo file",
"step_complete":     "Hoàn tất",

// Preflight
"check":             "Kiểm tra",
"result":            "Kết quả",
```

## Todo List

- [ ] Add English messages to `lang_en.go`
- [ ] Add Vietnamese messages to `lang_vi.go`
- [ ] Verify all keys are used in Phase 01 code
- [ ] Run build to check for missing keys

## Success Criteria

- [ ] All new i18n keys added to both language files
- [ ] Vietnamese uses proper diacritics
- [ ] No hardcoded strings in UI components
- [ ] Build passes

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| Missing keys | Low | Fallback to key name if not found |
| Translation errors | Low | Review by native speaker |

## Security Considerations

- No security concerns - static string data

## Next Steps

- Phase 03: Update command files to use new UI components

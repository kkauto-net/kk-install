# Phase 03: I18n & Polish

## Context
- **Parent Plan:** [plan.md](./plan.md)
- **Dependencies:** [Phase 01](./phase-01-core-ui-components.md), [Phase 02](./phase-02-command-updates.md)
- **Brainstorm:** [brainstorm report](../reports/brainstorm-260110-1620-cli-professional-output-v2.md)

## Overview

| Field | Value |
|-------|-------|
| Date | 2026-01-10 |
| Priority | P2 |
| Effort | 0.5h |
| Implementation Status | pending |
| Review Status | pending |

**Description:** Add all new i18n message keys for both English and Vietnamese (có dấu), run tests, and final polish.

## Key Insights

1. Vietnamese messages must use proper diacritics (có dấu)
2. New keys needed for banners, errors, table columns
3. Existing i18n system works well - just add keys

## Requirements

- R6: Default English, Vietnamese với dấu

## Related Code Files

| File | Action | Description |
|------|--------|-------------|
| `pkg/ui/lang_en.go` | MODIFY | Add new English messages |
| `pkg/ui/lang_vi.go` | MODIFY | Add new Vietnamese messages |

## Implementation Steps

### 1. Add keys to lang_en.go

```go
// Command banners
"status_desc":      "Service Status",
"init_desc":        "Docker Stack Initialization",
"start_desc":       "Start All Services",
"restart_desc":     "Restart All Services",
"update_desc":      "Pull & Recreate",

// Error box
"to_fix":           "To fix",
"then_run":         "Then run",

// Table columns
"col_image":        "Image",
"col_current":      "Current",
"col_new":          "New",
"col_file":         "File",

// Progress
"starting":         "starting...",
"ready":            "ready",
"unhealthy":        "unhealthy",
"services_started": "Services started",
```

### 2. Add keys to lang_vi.go

```go
// Command banners
"status_desc":      "Trạng thái dịch vụ",
"init_desc":        "Khởi tạo Docker Stack",
"start_desc":       "Khởi động tất cả dịch vụ",
"restart_desc":     "Khởi động lại tất cả dịch vụ",
"update_desc":      "Cập nhật & Khởi tạo lại",

// Error box
"to_fix":           "Để khắc phục",
"then_run":         "Sau đó chạy",

// Table columns
"col_image":        "Image",
"col_current":      "Hiện tại",
"col_new":          "Mới",
"col_file":         "Tệp",

// Progress
"starting":         "đang khởi động...",
"ready":            "sẵn sàng",
"unhealthy":        "không khỏe mạnh",
"services_started": "Đã khởi động dịch vụ",
```

### 3. Run Tests

```bash
# Run all ui tests
go test ./pkg/ui/... -v

# Build to verify no errors
go build ./...

# Run full test suite
make test
```

### 4. Manual Testing Checklist

Test each command and verify output:

```bash
# Test init (in empty directory)
mkdir /tmp/test-kk && cd /tmp/test-kk
kk init

# Test start
kk start

# Test status
kk status

# Test restart
kk restart

# Test update
kk update

# Clean up
cd - && rm -rf /tmp/test-kk
```

### 5. Vietnamese Testing

```bash
# Set Vietnamese in config
kk init  # Select Vietnamese

# Verify diacritics display correctly
kk status
```

## Todo List

- [ ] Add new keys to `pkg/ui/lang_en.go`
- [ ] Add new keys to `pkg/ui/lang_vi.go` (với dấu)
- [ ] Run `go test ./pkg/ui/... -v`
- [ ] Run `go build ./...`
- [ ] Run `make test`
- [ ] Manual test: `kk init`
- [ ] Manual test: `kk start`
- [ ] Manual test: `kk status`
- [ ] Manual test: `kk update`
- [ ] Test Vietnamese output with diacritics

## Success Criteria

1. All i18n keys present in both language files
2. Vietnamese messages display với dấu correctly
3. All unit tests pass
4. All commands work as expected
5. Output looks professional and consistent

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Missing i18n key | Low | Medium | Test all commands |
| Diacritics encoding issue | Low | Medium | Use UTF-8 properly |

## Security Considerations

- No security impact - i18n-only changes

## Final Checklist

Before marking plan complete:

- [ ] All tests pass
- [ ] All commands tested manually
- [ ] Both languages verified
- [ ] Code reviewed
- [ ] Ready for commit

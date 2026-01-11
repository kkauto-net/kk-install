---
title: Phase 03 Add UI Messages
description: Thêm các messages mới cho grouped form và credential configuration vào cả 2 language files.
status: completed
priority: medium
effort: 30 minutes
branch: main
tags: [ui, i18n, init]
created: 2026-01-11
---

# Phase 03: Add UI Messages

**Effort:** 30 minutes

## Objective

Thêm các messages mới cho grouped form và credential configuration vào cả 2 language files.

---

## Tasks

### 3.1 Update `pkg/ui/lang_en.go`

**Add after line 141 (after step messages):**

```go
// Credential configuration
"step_domain":       "Domain Configuration",
"step_credentials":  "Environment Configuration",
"ask_use_random":    "Use randomly generated secrets?",
"no_edit":           "No, I want to edit",
"group_system":      "System Configuration",
"group_db_secrets":  "Database Secrets",
"group_s3_secrets":  "S3 Storage Secrets",

// Validation
"error_empty_field": "This field cannot be empty",
"error_jwt_secret":  "Failed to generate JWT secret",
"error_s3_keys":     "Failed to generate S3 keys",
```

---

### 3.2 Update `pkg/ui/lang_vi.go`

**Add after line 141 (after step messages):**

```go
// Credential configuration
"step_domain":       "Cấu hình Domain",
"step_credentials":  "Cấu hình môi trường",
"ask_use_random":    "Sử dụng thông tin bảo mật ngẫu nhiên?",
"no_edit":           "Không, tôi muốn chỉnh sửa",
"group_system":      "Cấu hình hệ thống",
"group_db_secrets":  "Thông tin bảo mật Database",
"group_s3_secrets":  "Thông tin bảo mật S3",

// Validation
"error_empty_field": "Trường này không được để trống",
"error_jwt_secret":  "Không thể tạo JWT secret",
"error_s3_keys":     "Không thể tạo S3 keys",
```

---

## Message Keys Summary

| Key | EN | VI |
|-----|----|----|
| `step_domain` | Domain Configuration | Cấu hình Domain |
| `step_credentials` | Environment Configuration | Cấu hình môi trường |
| `ask_use_random` | Use randomly generated secrets? | Sử dụng thông tin bảo mật ngẫu nhiên? |
| `no_edit` | No, I want to edit | Không, tôi muốn chỉnh sửa |
| `group_system` | System Configuration | Cấu hình hệ thống |
| `group_db_secrets` | Database Secrets | Thông tin bảo mật Database |
| `group_s3_secrets` | S3 Storage Secrets | Thông tin bảo mật S3 |
| `error_empty_field` | This field cannot be empty | Trường này không được để trống |
| `error_jwt_secret` | Failed to generate JWT secret | Không thể tạo JWT secret |
| `error_s3_keys` | Failed to generate S3 keys | Không thể tạo S3 keys |

---

## Verification

```bash
# Verify Go compiles
go build ./pkg/ui/...

# Check no missing keys between EN/VI
# Both files should have same keys
```

## Output

- ✅ Updated `pkg/ui/lang_en.go` with 10 new messages
- ✅ Updated `pkg/ui/lang_vi.go` with 10 new Vietnamese messages

## Status

**COMPLETED:** 2026-01-11

**Notes:** Messages were added during Phase 02 refactoring as part of the init flow implementation.

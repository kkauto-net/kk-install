# Brainstorm: Redesign `kk init` Command

**Date:** 2026-01-11
**Status:** Approved

---

## Problem Statement

Current `kk init`:
- Không rõ ràng về services bắt buộc vs optional
- SeaweedFS, Caddy là optional nhưng MariaDB, Redis luôn được cài mà không hỏi
- Credentials tự generate mà không hiển thị cho user edit
- Template .env còn thiếu JWT_SECRET, S3 keys còn hardcode

**Goal:** Cải thiện UX cho init flow - rõ ràng hơn về services, cho phép customize credentials.

---

## Final Design

### Flow tổng quan

```
Step 1: Docker Check (giữ nguyên)
Step 2: Language Selection (giữ nguyên)
Step 3: Service Selection
  - kkengine: ✅ bắt buộc (không hỏi)
  - MariaDB:  ✅ default ON (không hỏi - kkengine depends)
  - Redis:    ✅ default ON (không hỏi - kkengine depends)
  - SeaweedFS: ☐ hỏi user
  - Caddy:     ☐ hỏi user

Step 4: Domain Configuration
  → Input: SYSTEM_DOMAIN (bắt buộc, pre-fill = "localhost")

Step 5: Environment Configuration
  → Confirm: "Dùng thông tin ngẫu nhiên?" [Yes/No]

  Nếu Yes → skip to Step 6
  Nếu No  → Show edit form với pre-filled random values:

    [Thông tin hệ thống]
    • JWT_SECRET: <random-32-chars>

    [Secrets - Database]
    • DB_PASSWORD:      <random-24-chars>
    • DB_ROOT_PASSWORD: <random-24-chars>
    • REDIS_PASSWORD:   <random-24-chars>

    [Secrets - S3] (chỉ hiện nếu EnableSeaweedFS)
    • S3_ACCESS_KEY: <random-20-chars>
    • S3_SECRET_KEY: <random-40-chars>

Step 6: Generate Files
Step 7: Complete Summary
```

### Changes Required

#### 1. `cmd/init.go`

| Section | Change |
|---------|--------|
| Step 3 | Chỉ hỏi SeaweedFS và Caddy (MariaDB+Redis mặc định ON) |
| Step 4 | Hỏi SYSTEM_DOMAIN trước Step 5 |
| Step 5 | Thêm confirm "random?" → conditional edit form |
| Form edit | Chia nhóm: System / DB Secrets / S3 Secrets |
| Generate | Thêm JWT_SECRET, S3_ACCESS_KEY, S3_SECRET_KEY vào tmplCfg |

#### 2. `pkg/templates/embed.go`

```go
type Config struct {
    // Services
    EnableSeaweedFS bool
    EnableCaddy     bool

    // System
    Domain    string
    JWTSecret string

    // Database
    DBPassword     string
    DBRootPassword string
    RedisPassword  string

    // S3 (only when EnableSeaweedFS)
    S3AccessKey string
    S3SecretKey string
}
```

#### 3. `pkg/templates/env.tmpl`

**Thêm:**
```
JWT_SECRET={{.JWTSecret}}
```

**Thay thế:**
```
S3_ACCESS_KEY={{.S3AccessKey}}
S3_SECRET_KEY={{.S3SecretKey}}
```

(Hiện đang hardcode `your_access_key` và `secret_key`)

#### 4. `pkg/ui/` - Thêm messages

```go
// lang_en.go
"ask_use_random":    "Use randomly generated secrets?",
"group_system":      "System Configuration",
"group_db_secrets":  "Database Secrets",
"group_s3_secrets":  "S3 Storage Secrets",

// lang_vi.go
"ask_use_random":    "Sử dụng thông tin bảo mật ngẫu nhiên?",
"group_system":      "Cấu hình hệ thống",
"group_db_secrets":  "Thông tin bảo mật Database",
"group_s3_secrets":  "Thông tin bảo mật S3",
```

### Files Unchanged

- `Caddyfile.tmpl` - giữ nguyên
- `kkphp.conf.tmpl` - giữ nguyên
- `kkfiler.toml.tmpl` - giữ nguyên
- `docker-compose.yml.tmpl` - giữ nguyên (đã dùng ${VAR} từ .env)

---

## Implementation Considerations

### UX Notes

1. **Pre-fill vs Placeholder:** Dùng pre-fill (`.Value(&var)` với giá trị khởi tạo) - user có thể xóa và nhập lại
2. **Form grouping:** charmbracelet/huh hỗ trợ `huh.NewGroup()` với title
3. **Conditional fields:** S3 fields chỉ hiện khi `enableSeaweedFS == true`

### Security Notes

1. **Password generation:** Đã có `ui.GeneratePassword()` - reuse
2. **S3 key format:**
   - ACCESS_KEY: 20 chars alphanumeric
   - SECRET_KEY: 40 chars alphanumeric
3. **JWT_SECRET:** 32 chars, base64-safe

### Risks

| Risk | Mitigation |
|------|------------|
| User bỏ trống field | Validate non-empty, show warning |
| S3 keys không valid format | Generate với format chuẩn, không cho nhập sai format |
| Existing tests break | Update test cases cho new Config struct |

---

## Success Metrics

- [ ] `kk init` chạy thành công với flow mới
- [ ] .env chứa JWT_SECRET và S3 keys từ user input hoặc random
- [ ] docker-compose.yml vẫn đọc đúng từ .env
- [ ] Tests pass

---

## Next Steps

1. Update `templates.Config` struct
2. Update `env.tmpl` template
3. Refactor `cmd/init.go` với new flow
4. Add new UI messages (lang_en, lang_vi)
5. Update tests
6. Manual testing

---

## Unresolved Questions

None - tất cả đã được clarify.

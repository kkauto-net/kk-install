# Brainstorm: kk init Enhancement - Template Sync & UX Improvements

**Date:** 2026-01-05 08:43
**Context:** Cải thiện `kk init` command để sử dụng config từ example, mặc định SeaweedFS/Caddy=yes, hỗ trợ đa ngôn ngữ, và UI đẹp hơn

---

## Problem Statement

### Current Issues (Phát hiện từ `/home/tieutinh/kktest1`)

1. **Template Placeholder Content**
   - `Caddyfile.tmpl`: chỉ có placeholder "caddy config for {{.Domain}}"
   - `kkfiler.toml.tmpl`: chỉ có placeholder "seaweedfs config for {{.Domain}}"
   - `kkphp.conf.tmpl`: chỉ có placeholder "kkphp config"
   - **Vấn đề**: Files tạo ra không sử dụng được, thiếu config thực tế từ `/example`

2. **Default Options Not Set**
   - SeaweedFS và Caddy hiện tại default=false
   - User phải manually chọn Yes/No, không có default value

3. **No Language Support**
   - Tất cả messages hardcoded Vietnamese
   - Không có cơ chế chọn ngôn ngữ (English/Vietnamese)

4. **Basic UI/UX**
   - Messages đơn giản, thiếu formatting
   - Không có icons, colors rõ ràng
   - Thiếu progress indicators

---

## Requirements Analysis

### User Requirements
✅ **Templates từ example**: Copy full content từ `example/*.{conf,toml,Caddyfile}` vào templates
✅ **Default Options**: SeaweedFS & Caddy default=yes, có thể toggle
✅ **Language Selection**: Interactive menu chọn English/Vietnamese đầu tiên
✅ **Enhanced UI**: Icons, colors, better formatting, progress indicators

### Technical Constraints
- Maintain backward compatibility với current config structure
- Templates phải support Go text/template syntax
- UI library: `pterm` (đã có) + `huh` (đã có)
- Không thay đổi core command interface

---

## Solution Design

### 1️⃣ Template Sync Strategy

#### Approach: Copy Example Content to Template Files

**Files cần update:**
```
pkg/templates/Caddyfile.tmpl       → Copy từ example/Caddyfile + template vars
pkg/templates/kkfiler.toml.tmpl    → Copy từ example/kkfiler.toml + template vars
pkg/templates/kkphp.conf.tmpl      → Copy từ example/kkphp.conf (static)
```

**Implementation:**
1. Đọc content từ `example/Caddyfile`:
   ```
   {$SYSTEM_DOMAIN} {
       reverse_proxy kkengine:8019
   }
   ```
   → Convert sang template syntax:
   ```
   {{.Domain}} {
       reverse_proxy kkengine:8019
   }
   ```

2. `kkfiler.toml` - thay environment vars bằng template vars:
   ```
   # FROM (example):
   # hostname, port, etc set via WEED_MYSQL_* env vars

   # TO (template):
   hostname = "{{.DBHostname}}"
   port = {{.DBPort}}
   ...
   ```

3. `kkphp.conf` - static file, không cần template vars (copy nguyên bản)

**Pros:**
- ✅ Sử dụng được ngay khi tạo ra
- ✅ Sync với example configs đã test
- ✅ Dễ maintain: update example → update template

**Cons:**
- ⚠️ Cần sync manual khi example thay đổi (có thể automate bằng test)
- ⚠️ Template phức tạp hơn

---

### 2️⃣ Default Options Implementation

**Current behavior:**
```go
huh.NewConfirm().
    Title("Bat SeaweedFS file storage?").
    Value(&enableSeaweedFS),  // default = false
```

**Enhanced behavior:**
```go
enableSeaweedFS := true  // Set default = true
enableCaddy := true      // Set default = true

huh.NewConfirm().
    Title("Bat SeaweedFS file storage?").
    Description("SeaweedFS la he thong luu tru file phan tan").
    Affirmative("Yes (default)").  // Indicate default
    Negative("No").
    Value(&enableSeaweedFS),
```

**User experience:**
- Press Enter → Accept default (Yes)
- Arrow keys → Change to No

**Pros:**
- ✅ Faster workflow cho common use case
- ✅ Clear indication của default value
- ✅ Vẫn flexible cho user

---

### 3️⃣ Multi-Language Support

**Architecture:**

```
pkg/
  ui/
    messages.go      → Message functions (renamed from current)
    i18n.go          → NEW: Language manager
    lang_en.go       → NEW: English messages
    lang_vi.go       → NEW: Vietnamese messages
```

**Implementation:**

```go
// i18n.go
type Language string
const (
    LangEN Language = "en"
    LangVI Language = "vi"
)

var currentLang = LangEN  // default

func SetLanguage(lang Language) { currentLang = lang }
func GetLanguage() Language { return currentLang }

// Message dispatcher
func Msg(key string) string {
    switch currentLang {
    case LangVI:
        return msgVI[key]
    default:
        return msgEN[key]
    }
}
```

**User Flow (cmd/init.go):**
```go
func runInit(cmd *cobra.Command, args []string) error {
    // STEP 0: Language selection
    var langChoice string
    langForm := huh.NewForm(
        huh.NewGroup(
            huh.NewSelect[string]().
                Title("Select language / Chọn ngôn ngữ").
                Options(
                    huh.NewOption("English", "en"),
                    huh.NewOption("Tiếng Việt", "vi").Selected(),  // Default VI
                ).
                Value(&langChoice),
        ),
    )
    langForm.Run()
    ui.SetLanguage(ui.Language(langChoice))

    // STEP 1: Check Docker (now uses translated messages)
    ui.ShowInfo(ui.Msg("checking_docker"))
    ...
}
```

**Message Keys:**
```
checking_docker, docker_ok, docker_not_installed,
init_complete, next_steps, created_file, ...
```

**Pros:**
- ✅ Clean separation giữa logic và presentation
- ✅ Easy để add thêm languages (Japanese, Chinese...)
- ✅ Consistent message management

**Cons:**
- ⚠️ Tăng code complexity
- ⚠️ Phải maintain 2 language files

---

### 4️⃣ Enhanced UI/UX

**Components to enhance:**

#### A. Language Selection Screen
```
┌──────────────────────────────────────┐
│  🌍 Language Selection / Chọn ngôn ngữ │
│                                      │
│  ○ English                           │
│  ● Tiếng Việt (default)             │
│                                      │
│  ↑↓: Navigate  Enter: Confirm       │
└──────────────────────────────────────┘
```

#### B. Docker Validation
```
⚙️  Checking Docker installation...
✅ Docker is ready

📁 Initializing in: /path/to/project
```

#### C. Configuration Prompts (with icons)
```
🗄️  SeaweedFS File Storage
   Distributed file storage system
   ● Yes (recommended)  ○ No

🌐 Caddy Web Server
   Reverse proxy with automatic HTTPS
   ● Yes (recommended)  ○ No

🔗 Domain Configuration
   Enter your domain: example.com
```

#### D. Progress Indicators
```
📝 Generating configuration files...
  ✅ docker-compose.yml
  ✅ .env
  ✅ kkphp.conf
  ✅ Caddyfile
  ✅ kkfiler.toml

🎉 Initialization complete!

Next steps:
  1. Review and edit .env if needed
  2. Run: kk start
```

**Implementation using pterm:**
```go
// Spinner during file generation
spinner, _ := pterm.DefaultSpinner.Start("Generating configuration files...")
// ... generate files ...
spinner.Success("Configuration files generated")

// Box for completion
pterm.DefaultBox.WithTitle("🎉 Success").Println(
    "Initialization complete!\n\nNext steps:\n  1. Review .env\n  2. Run: kk start",
)
```

**Pros:**
- ✅ Professional appearance
- ✅ Clear visual hierarchy
- ✅ Better user engagement
- ✅ Accessibility improvements

---

## Implementation Plan Structure

### Phase 1: Template Sync (Critical)
**Priority: P0**
- [ ] Copy example configs to templates
- [ ] Add template variables for dynamic values
- [ ] Test generated files are valid
- [ ] Add validation tests

**Files:**
- `pkg/templates/*.tmpl`
- `pkg/templates/embed_test.go` (new tests)

### Phase 2: Default Options (Quick Win)
**Priority: P0**
- [ ] Set default values for SeaweedFS & Caddy
- [ ] Update confirm prompts with default indicators
- [ ] Update tests

**Files:**
- `cmd/init.go`

### Phase 3: Multi-Language (Feature)
**Priority: P1**
- [ ] Create i18n infrastructure
- [ ] Extract all messages to lang files
- [ ] Add language selection screen
- [ ] Update all user-facing messages

**Files:**
- `pkg/ui/i18n.go` (new)
- `pkg/ui/lang_en.go` (new)
- `pkg/ui/lang_vi.go` (new)
- `pkg/ui/messages.go` (refactor)
- `cmd/init.go`

### Phase 4: UI/UX Enhancement (Polish)
**Priority: P2**
- [ ] Add icons to all messages
- [ ] Implement progress indicators
- [ ] Add formatted boxes
- [ ] Color coding consistency

**Files:**
- `pkg/ui/messages.go`
- `cmd/init.go`

---

## Risk Assessment

### Technical Risks

1. **Template Complexity**
   - Risk: Go templates có thể render sai nếu vars thiếu
   - Mitigation: Comprehensive tests với all config combinations

2. **Backward Compatibility**
   - Risk: Existing users có thể bị surprise với new defaults
   - Mitigation: Clear changelog, versioning

3. **Language Files Drift**
   - Risk: EN và VI messages không sync
   - Mitigation: Add CI check để verify message keys match

### UX Risks

1. **Language Selection Overhead**
   - Risk: Thêm 1 step có thể chậm workflow
   - Mitigation: Remember last choice trong config file `~/.kk/config.yaml`

2. **Information Overload**
   - Risk: Quá nhiều icons/colors có thể overwhelming
   - Mitigation: A/B testing với users, tunable verbosity

---

## Success Metrics

### Functional Requirements
- [ ] Generated files từ templates work without modification
- [ ] Default yes cho SeaweedFS/Caddy reduce steps for 80% users
- [ ] Language selection works cho both EN và VI
- [ ] UI enhancement không làm chậm performance

### Quality Metrics
- [ ] Test coverage ≥ 80% cho template rendering
- [ ] Zero config errors sau kk init
- [ ] User feedback score ≥ 4/5 for UX improvements

---

## Alternative Approaches (Rejected)

### Alt 1: Embed Example Files Directly
**Approach:** Copy example files as-is, replace strings post-generation
**Rejected because:**
- ❌ Không flexible cho dynamic values
- ❌ String replacement error-prone
- ❌ Khó maintain khi example thay đổi

### Alt 2: Flag-based Language Selection
**Approach:** `kk init --lang=vi` instead of interactive
**Rejected because:**
- ❌ Không user-friendly cho first-time users
- ❌ Requires documentation lookup
- ✅ **Could add as optional enhancement later**

### Alt 3: Always Enable SeaweedFS/Caddy
**Approach:** Không hỏi, always include trong compose
**Rejected because:**
- ❌ Quá opinionated, giảm flexibility
- ❌ Tăng resource usage không cần thiết cho minimal setups

---

## Open Questions

1. **Config Persistence**
   - Q: Should language choice be saved to `~/.kk/config.yaml`?
   - A: [Pending user feedback] - Có thể là future enhancement

2. **Template Variable Expansion**
   - Q: Có cần thêm vars như `{{.ProjectName}}`, `{{.Ports.XXX}}`?
   - A: [Defer to Phase 1 testing] - Chỉ implement khi thực sự cần

3. **Migration Path**
   - Q: Users đã có existing configs react như thế nào?
   - A: Backup mechanism đã có (`.bak` files) - sufficient

---

## Next Steps

### Immediate Actions
1. ✅ User confirmed requirements
2. 🔄 Create detailed implementation plan using `/plan`
3. ⏳ Execute Phase 1: Template Sync (critical path)

### Follow-up
- Gather user feedback sau Phase 2
- Iterate on UI/UX based on real usage
- Document breaking changes in CHANGELOG

---

## Conclusion

Solution phân tích trên address tất cả requirements:
- ✅ Sync templates với example configs
- ✅ Default SeaweedFS/Caddy = yes
- ✅ Multi-language support với interactive selection
- ✅ Enhanced UI với icons, colors, formatting

Implementation chia 4 phases, ưu tiên P0 (template sync + defaults) để ship value nhanh, sau đó iterate với language và UI enhancements.

**Recommended approach: Proceed với implementation plan.**

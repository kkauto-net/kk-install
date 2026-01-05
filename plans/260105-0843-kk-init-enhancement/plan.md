---
title: "kk init Enhancement - Template Sync & UX Improvements"
description: "Cải thiện kk init với templates hoạt động, defaults tốt hơn, đa ngôn ngữ và UI đẹp hơn"
status: completed
priority: P0
effort: 8h
branch: main
tags: [init, templates, i18n, ux, cli]
created: 2026-01-05
completion_timestamp: 2026-01-05
---

# kk init Enhancement - Implementation Plan

## Overview

Cải thiện lệnh `kk init` qua 4 phases: sync templates với example configs, set defaults tốt hơn, thêm multi-language support, và nâng cấp UI/UX.

## Context

- **Brainstorm Report**: [brainstormer-260105-0843-kk-init-improvement.md](../reports/brainstormer-260105-0843-kk-init-improvement.md)
- **i18n Research**: [researcher-01-i18n-libraries.md](./research/researcher-01-i18n-libraries.md)
- **Template Testing Research**: [researcher-02-template-testing.md](./research/researcher-02-template-testing.md)
- **Codebase Summary**: [codebase-summary.md](/home/kkdev/kkcli/docs/codebase-summary.md)
- **Code Standards**: [code-standards.md](/home/kkdev/kkcli/docs/code-standards.md)

## Current State Analysis

### Issues Identified

| Issue | Location | Impact |
|-------|----------|--------|
| Templates chỉ có placeholder text | `pkg/templates/*.tmpl` | Files tạo ra không dùng được |
| SeaweedFS/Caddy default=false | `cmd/init.go` | Phải manually chọn |
| Hardcoded Vietnamese | `pkg/ui/messages.go` | Không hỗ trợ EN |
| Basic UI | `cmd/init.go` | Thiếu icons/progress |

### Current Template Content

```
Caddyfile.tmpl:      "caddy config for {{.Domain}}"         <- placeholder
kkfiler.toml.tmpl:   "seaweedfs config for {{.Domain}}"     <- placeholder
kkphp.conf.tmpl:     "kkphp config"                         <- placeholder
docker-compose.yml.tmpl: OK (full content)
env.tmpl:            OK (full content)
```

### Example Files (Source of Truth)

```
example/Caddyfile     <- {$SYSTEM_DOMAIN} { reverse_proxy kkengine:8019 }
example/kkfiler.toml  <- Full SeaweedFS config with MySQL backend
example/kkphp.conf    <- Full PHP-FPM config (static)
```

## Phase Overview

| Phase | Priority | Effort | Description | Dependencies |
|-------|----------|--------|-------------|--------------|
| [Phase 1](./phase-01-template-sync.md) | P0 | 3h | Template Sync - Critical Path | None | DONE
| [Phase 2](./phase-02-default-options.md) | P0 | 1h | Default Options - Quick Win | None | DONE
| [Phase 3](./phase-03-multi-language.md) | P1 | 2.5h | Multi-Language Support | Phase 1, 2 | DONE
| [Phase 4](./phase-04-ui-ux-enhancement.md) | P2 | 1.5h | UI/UX Enhancement | Phase 1, 2, 3 | DONE

**Note**: P0 phases (1 & 2) có thể implement parallel.

## Success Criteria

### Functional
- [x] Generated files từ templates work without modification
- [x] Default yes cho SeaweedFS/Caddy reduces setup steps
- [x] Language selection works cho both EN và VI
- [x] UI icons và progress indicators hoạt động

### Quality
- [x] Test coverage >= 80% cho template rendering
- [x] Zero config errors sau `kk init`
- [x] All YAML/TOML syntax valid

## Technical Constraints

1. Maintain backward compatibility
2. Use existing libraries: pterm (UI), huh (forms)
3. Go text/template syntax
4. No breaking CLI changes
5. Follow Go code standards từ docs/code-standards.md

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Template render errors | Medium | High | Comprehensive tests với all Config combinations |
| Language files drift | Low | Medium | CI check để verify message keys match |
| Breaking existing workflows | Low | High | Test with existing projects |

## Execution Plan

```
Week 1, Day 1:
├── Phase 1: Template Sync (3h) [PARALLEL]
└── Phase 2: Default Options (1h) [PARALLEL]

Week 1, Day 2:
└── Phase 3: Multi-Language (2.5h)

Week 1, Day 3:
└── Phase 4: UI/UX Enhancement (1.5h)
```

## Validation Summary

**Validated:** 2026-01-05 09:15
**Questions asked:** 4

### Confirmed Decisions

1. **Template Approach (Phase 1)**: ✅ Full copy với minimal template vars
   - Copy kkfiler.toml nguyên bản (config qua env vars)
   - Chỉ template {{.Domain}} cho Caddyfile
   - Rationale: Keep templates simple, avoid over-engineering

2. **i18n Implementation (Phase 3)**: ✅ Map-based approach
   - Simple map-based (lang_en.go, lang_vi.go)
   - Không cần external dependencies
   - Can migrate to go-i18n library later if needed

3. **Test Coverage (Phase 1)**: ✅ ≥80% coverage với full validation
   - All templates, all combinations
   - TOML/YAML syntax validation
   - Golden files comparison
   - Rationale: Comprehensive tests prevent config errors in production

4. **Language Default (Phase 3)**: ⚠️ **CHANGED FROM PLAN**
   - Original plan: Vietnamese default
   - User decision: **English default**
   - Action required: Update Phase 3 implementation
     - Change `var currentLang = LangEN` (was LangVI)
     - Update language select default: `.Selected()` on English option
     - Update tests to reflect EN default

### Action Items

- [x] Update Phase 3 plan: Change default language from VI to EN ✅
- [x] Update lang selection in cmd/init.go: English `.Selected()` ✅
- [x] Update i18n.go: `var currentLang = LangEN` ✅
- [x] Update i18n tests to reflect EN default ✅

**Code Review**: [code-reviewer-260105-1613-action-items.md](../reports/code-reviewer-260105-1613-action-items.md)
**Status**: ✅ APPROVED (1 minor bug fix recommended)
**Date**: 2026-01-05 16:13

## Completion Summary

Kế hoạch "kk init Enhancement" đã hoàn thành thành công. Tất cả 4 phase đã được triển khai đầy đủ, bao gồm đồng bộ hóa template, tùy chọn mặc định, hỗ trợ đa ngôn ngữ và cải tiến UI/UX. Tất cả các mục hành động đã được hoàn thành, các bài kiểm tra đã pass (6/6 gói) và code đã được review, chấp thuận.

## Unresolved Questions

1. **Config Persistence**: Should language choice be saved to `~/.kk/config.yaml`?
   - Decision: Defer to future enhancement - keep simple for now

2. **Template Variable Expansion**: Need more vars like `{{.ProjectName}}`?
   - Decision: Only implement what's needed - YAGNI

---

**Next Step**: Start with Phase 1 (Template Sync) and Phase 2 (Default Options) in parallel.

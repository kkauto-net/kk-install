---
title: Phase 01 Update Templates
description: Cập nhật template config struct và env.tmpl để hỗ trợ JWT_SECRET và dynamic S3 keys.
status: completed
priority: high
effort: 30 minutes
branch: main
tags: [templates, backend, init]
created: 2026-01-11
---

# Phase 01: Update Templates

**Effort:** 30 minutes

## Objective

Cập nhật template config struct và env.tmpl để hỗ trợ JWT_SECRET và dynamic S3 keys.

## Tasks

### 1.1 Update `pkg/templates/embed.go`

**File:** `/home/kkdev/kkcli/pkg/templates/embed.go`

**Current Config (line 13-20):**
```go
type Config struct {
    EnableSeaweedFS bool
    EnableCaddy     bool
    DBPassword      string
    DBRootPassword  string
    RedisPassword   string
    Domain          string
}
```

**Target Config:**
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

    // S3 (only used when EnableSeaweedFS)
    S3AccessKey string
    S3SecretKey string
}
```

**Steps:**
1. Add `JWTSecret string` after `Domain`
2. Add `S3AccessKey string` and `S3SecretKey string` at end
3. Add comments for grouping fields

---

### 1.2 Update `pkg/templates/env.tmpl`

**File:** `/home/kkdev/kkcli/pkg/templates/env.tmpl`

**Change 1 - Add JWT_SECRET (after line 17):**

```diff
 #--------------------------------------------------------------------
 # SYSTEM CONFIG
 #--------------------------------------------------------------------
 RATE_LIMIT_HTTP_PER_SECOND=100
 RATE_LIMIT_WS_EVENTS_PER_SECOND=50
+
+# JWT Authentication
+JWT_SECRET={{.JWTSecret}}
```

**Change 2 - Replace hardcoded S3 keys (line 48-49):**

```diff
 # Seaweedfs
 S3_DRIVER=s3
 S3_ENDPOINT=http://seaweedfs:8333
 S3_REGION=us-east-1
-S3_ACCESS_KEY=your_access_key
-S3_SECRET_KEY=secret_key
+S3_ACCESS_KEY={{.S3AccessKey}}
+S3_SECRET_KEY={{.S3SecretKey}}
```

---

## Verification

```bash
# Verify struct compiles
go build ./pkg/templates/...

# Verify template parses
go test ./pkg/templates/... -v
```

## Output

- ✅ Updated `pkg/templates/embed.go` with new Config fields
- ✅ Updated `pkg/templates/env.tmpl` with dynamic JWT_SECRET and S3 keys
- ✅ Updated `pkg/templates/embed_test.go` with test data for new fields
- ✅ Regenerated `pkg/templates/testdata/golden/env.golden`

## Status

**COMPLETED:** 2026-01-11 08:33

**Test Results:** ✅ 8/8 passing
**Build Status:** ✅ Pass
**Code Review:** See `/home/kkdev/kkcli/plans/reports/code-reviewer-260111-0833-phase01-template-update.md`

**Recommendations Before Phase 02:**
1. Add secret validation (JWT min 32 chars, S3 keys min 16/32 chars)
2. Document JWT_SECRET minimum length in code comments

**Next Phase:** Phase 02 - Refactor Init Flow (populate new fields in cmd/init.go)

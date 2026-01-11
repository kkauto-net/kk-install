---
title: Phase 04 Update Tests
description: Cập nhật test cases cho new Config struct và verify template rendering.
status: completed
priority: medium
effort: 1 hour
branch: main
tags: [testing, templates, init]
created: 2026-01-11
---

# Phase 04: Update Tests

**Effort:** 1 hour

## Objective

Cập nhật test cases cho new Config struct và verify template rendering.

---

## Tasks

### 4.1 Find Existing Tests

```bash
# Search for template tests
find . -name "*_test.go" | xargs grep -l "templates.Config\|RenderAll\|RenderTemplate"

# Expected files:
# - pkg/templates/embed_test.go
# - cmd/init_test.go (if exists)
```

### 4.2 Update Template Tests

**File:** `pkg/templates/embed_test.go` (if exists)

**Update test Config with new fields:**

```go
func TestRenderAll(t *testing.T) {
    cfg := templates.Config{
        EnableSeaweedFS: true,
        EnableCaddy:     true,
        Domain:          "test.example.com",
        JWTSecret:       "test-jwt-secret-32-chars-long!!", // NEW
        DBPassword:      "testdbpass",
        DBRootPassword:  "testrootpass",
        RedisPassword:   "testredispass",
        S3AccessKey:     "TESTACCESS1234567890",            // NEW
        S3SecretKey:     "testsecretkey1234567890123456789012345678", // NEW
    }

    // ... rest of test
}
```

### 4.3 Add Test for New Config Fields in env.tmpl

```go
func TestEnvTemplateContainsNewFields(t *testing.T) {
    cfg := templates.Config{
        EnableSeaweedFS: true,
        EnableCaddy:     false,
        Domain:          "localhost",
        JWTSecret:       "my-jwt-secret",
        DBPassword:      "dbpass",
        DBRootPassword:  "rootpass",
        RedisPassword:   "redispass",
        S3AccessKey:     "MYACCESSKEY12345",
        S3SecretKey:     "mysecretkey123456789012345678901234567890",
    }

    tmpDir := t.TempDir()
    err := templates.RenderAll(cfg, tmpDir)
    require.NoError(t, err)

    // Read .env
    envContent, err := os.ReadFile(filepath.Join(tmpDir, ".env"))
    require.NoError(t, err)

    // Verify new fields present
    assert.Contains(t, string(envContent), "JWT_SECRET=my-jwt-secret")
    assert.Contains(t, string(envContent), "S3_ACCESS_KEY=MYACCESSKEY12345")
    assert.Contains(t, string(envContent), "S3_SECRET_KEY=mysecretkey123456789012345678901234567890")

    // Verify old hardcoded values NOT present
    assert.NotContains(t, string(envContent), "your_access_key")
    assert.NotContains(t, string(envContent), "secret_key")
}
```

### 4.4 Test generateS3AccessKey Helper

**File:** `cmd/init_test.go` (create if needed)

```go
func TestGenerateS3AccessKey(t *testing.T) {
    key, err := generateS3AccessKey(20)
    require.NoError(t, err)

    // Check length
    assert.Len(t, key, 20)

    // Check format: only uppercase letters and digits
    for _, c := range key {
        assert.True(t, (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9'),
            "Invalid character in S3 access key: %c", c)
    }
}

func TestGenerateS3AccessKeyUniqueness(t *testing.T) {
    keys := make(map[string]bool)
    for i := 0; i < 100; i++ {
        key, _ := generateS3AccessKey(20)
        assert.False(t, keys[key], "Duplicate key generated")
        keys[key] = true
    }
}
```

### 4.5 Update Golden File Tests (if exists)

Nếu project dùng golden file tests cho templates:

1. Regenerate golden files với new Config structure
2. Update expected output trong golden files

```bash
# Regenerate golden files
go test ./pkg/templates/... -update
```

---

## Verification Commands

```bash
# Run all tests
make test

# Run template tests specifically
go test ./pkg/templates/... -v

# Run init tests
go test ./cmd/... -v

# Check test coverage
make test-coverage
```

---

## Expected Test Files

| File | Purpose |
|------|---------|
| `pkg/templates/embed_test.go` | Test RenderAll, RenderTemplate với new Config |
| `cmd/init_test.go` | Test generateS3AccessKey helper |

## Output

- ✅ Updated template tests for new Config fields
- ✅ New test for JWT_SECRET and S3 keys in .env output
- ✅ New test for generateS3AccessKey helper function
- ✅ All tests passing (8/8 packages)

## Status

**COMPLETED:** 2026-01-11

**Test Results:**
- `pkg/templates`: 8 tests passing
- `kk_integration_test.go`: All init tests passing with JWT_SECRET and S3 keys validation
- Golden files: Updated and passing
- ValidateSecrets: New tests for secret length validation

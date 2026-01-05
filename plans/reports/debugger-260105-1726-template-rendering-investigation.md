# Template Rendering Issues Investigation Report

**Date**: 2026-01-05
**Investigator**: Debugger Agent
**Mission**: Root cause analysis for template rendering failures in kkcli

---

## Executive Summary

**Root Cause**: Template files `docker-compose.yml.tmpl` and `env.tmpl` contain **literal escape sequences** (`\n`) instead of actual newline characters. This causes rendered output to display raw `\n` strings rather than line breaks.

**Business Impact**:
- Generated docker-compose.yml is **invalid YAML** (unparseable)
- Generated .env contains malformed entries
- Missing SeaweedFS & Caddy services from docker-compose.yml
- Users cannot run `docker-compose up` with generated configs

**Severity**: CRITICAL - Blocks core `kk init` functionality

---

## Technical Analysis

### Issue #1: Literal `\n` Characters in Templates

**Evidence from hex dump**:
```bash
# docker-compose.yml.tmpl contains literal \n (hex: 5c6e) not newlines (hex: 0a)
$ xxd /home/kkdev/kkcli/pkg/templates/docker-compose.yml.tmpl | head -5
00000000: 7665 7273 696f 6e3a 2027 332e 3827 5c6e  version: '3.8'\n
00000010: 7365 7276 6963 6573 3a5c 6e20 206b 6b65  services:\n  kke
                                    ^^^^
```

**File characteristics**:
```bash
$ wc -l pkg/templates/*.tmpl
    3 Caddyfile.tmpl      # ✓ CORRECT (3 actual lines)
    1 docker-compose.yml.tmpl  # ✗ BROKEN (should be ~60 lines)
    1 env.tmpl            # ✗ BROKEN (should be ~17 lines)
   20 kkfiler.toml.tmpl  # ✓ CORRECT
   20 kkphp.conf.tmpl    # ✓ CORRECT

$ file pkg/templates/*.tmpl
Caddyfile.tmpl: ASCII text
docker-compose.yml.tmpl: ASCII text, with very long lines (1447)  # ← RED FLAG
env.tmpl: ASCII text, with very long lines (415)                  # ← RED FLAG
kkfiler.toml.tmpl: ASCII text
kkphp.conf.tmpl: ASCII text
```

**Rendered output comparison**:
```yaml
# test/docker-compose.yml (ACTUAL OUTPUT - BROKEN)
version: '3.8'\nservices:\n  kkengine:\n    image: kkengine:latest\n...
                ^^  Literal \n shown instead of line break

# Should be:
version: '3.8'
services:
  kkengine:
    image: kkengine:latest
```

**Affected template files**:
- `/home/kkdev/kkcli/pkg/templates/docker-compose.yml.tmpl` - Lines 1 (entire file is single line)
- `/home/kkdev/kkcli/pkg/templates/env.tmpl` - Line 1 (entire file is single line)

---

### Issue #2: Hardcoded Passwords Instead of Template Variables

**Evidence from test/docker-compose.yml** (lines 25-34):
```yaml
  db:
    environment:
      MYSQL_ROOT_PASSWORD: mFVf7M_KPZh5GdXKquBpeMdM  # ✗ HARDCODED
      MYSQL_PASSWORD: fUJmWkye_pro67_lS0ysaGin      # ✗ HARDCODED
```

**Correct template syntax** (from pkg/templates/docker-compose.yml.tmpl):
```yaml
  db:
    environment:
      MYSQL_ROOT_PASSWORD: {{.DBRootPassword}}  # ✓ CORRECT
      MYSQL_PASSWORD: {{.DBPassword}}           # ✓ CORRECT
```

**HOWEVER**: Template is correct but output shows hardcoded values because template contains literal `\n`, causing parsing issues.

---

### Issue #3: Missing SeaweedFS & Caddy Services

**Evidence**:
```bash
$ grep -n "seaweedfs\|caddy" test/docker-compose.yml
# (no output - services missing)

$ wc -l test/docker-compose.yml example/docker-compose.yml
    1 test/docker-compose.yml       # ✗ Only 1 line (should be ~60)
  126 example/docker-compose.yml    # ✓ Contains all services
```

**Expected services** (from example/docker-compose.yml):
- kkengine ✓ (present)
- db (mariadb) ✓ (present)
- redis ✓ (present)
- seaweedfs ✗ (MISSING)
- caddy ✗ (MISSING)

**Root cause**: Template file is single-line string, rendering logic cannot conditionally include services properly.

---

### Issue #4: No Template Conditional Logic for Services

**Analysis of template rendering**:

Template rendering code (`pkg/templates/embed.go:58-86`) only controls which **separate files** get generated:
```go
if cfg.EnableCaddy {
    files["Caddyfile"] = "Caddyfile"  // ✓ Separate Caddyfile
}
if cfg.EnableSeaweedFS {
    files["kkfiler.toml"] = "kkfiler.toml"  // ✓ Separate kkfiler.toml
}
```

**BUT**: docker-compose.yml template has NO conditional blocks to include/exclude services:
```yaml
# docker-compose.yml.tmpl should have:
{{if .EnableSeaweedFS}}
  seaweedfs:
    image: chrislusf/seaweedfs:latest
    ...
{{end}}

{{if .EnableCaddy}}
  caddy:
    image: caddy:alpine
    ...
{{end}}
```

**Current template**: Static content copied from example/, no {{if}} conditionals.

---

### Issue #5: Test Coverage Gaps

**Evidence from pkg/templates/embed_test.go:232-233**:
```go
func TestValidateYAML(t *testing.T) {
    t.Skip("Skipping YAML validation - docker-compose.yml.tmpl needs proper newlines (out of scope for Phase 1)")
    // ^^^ Test explicitly skipped acknowledging the issue
```

**Test blind spots**:
1. No validation of actual newline characters (vs literal `\n`)
2. TestRenderTemplate validates content using `strings.Contains()` which matches even broken output
3. No hex dump validation
4. No YAML parser validation (skipped)
5. No line count validation

---

## Root Cause Chain

1. **Template sync process** copied example/ files to pkg/templates/*.tmpl
2. **Escape sequence conversion** occurred during copy (possibly via `echo` or JSON processing)
3. **Newlines became literal** `\n` strings instead of 0x0A bytes
4. **Template.Execute()** renders literal strings faithfully → output contains `\n`
5. **YAML parser fails** because file is invalid YAML
6. **Missing services** because template is unparseable single-line blob
7. **Tests pass** because validation checks were too permissive

---

## Affected Files

### Critical (Must Fix):
1. `/home/kkdev/kkcli/pkg/templates/docker-compose.yml.tmpl`
   - Issue: Entire file is 1 line with literal `\n` (should be ~60 lines)
   - Fix: Replace with properly formatted multi-line template

2. `/home/kkdev/kkcli/pkg/templates/env.tmpl`
   - Issue: Entire file is 1 line with literal `\n` (should be ~17 lines)
   - Fix: Replace with properly formatted multi-line template

### Enhancements (Should Fix):
3. `/home/kkdev/kkcli/pkg/templates/docker-compose.yml.tmpl`
   - Issue: Missing {{if .EnableSeaweedFS}} / {{if .EnableCaddy}} conditionals
   - Fix: Add template conditionals to include/exclude services

4. `/home/kkdev/kkcli/pkg/templates/embed_test.go`
   - Issue: YAML validation test skipped
   - Fix: Re-enable TestValidateYAML after fixing templates

---

## Proposed Fixes

### Fix #1: Replace Broken Templates (URGENT)

**Action**: Manually recreate templates from example/ files with proper newlines

**Method**:
```bash
# Copy example files and add template variables
cp example/docker-compose.yml pkg/templates/docker-compose.yml.tmpl
cp example/.env pkg/templates/env.tmpl

# Replace hardcoded values with template variables (manual editing required)
# In docker-compose.yml.tmpl:
#   - Replace actual passwords → {{.DBPassword}}, {{.DBRootPassword}}, {{.RedisPassword}}
#   - Replace domain → {{.Domain}}
# In env.tmpl:
#   - Replace actual passwords → {{.DBPassword}}, {{.DBRootPassword}}, {{.RedisPassword}}
#   - Replace domain → {{.Domain}}
```

**Verification**:
```bash
wc -l pkg/templates/docker-compose.yml.tmpl  # Should be ~60, not 1
file pkg/templates/docker-compose.yml.tmpl   # Should NOT say "very long lines"
xxd pkg/templates/docker-compose.yml.tmpl | head -3  # Should show 0a (newlines), not 5c6e (\n)
```

---

### Fix #2: Add Template Conditionals

**In docker-compose.yml.tmpl**, wrap optional services:

```yaml
services:
  kkengine:
    # ... (always included)

  db:
    # ... (always included)

  redis:
    # ... (always included)

{{if .EnableSeaweedFS}}
  seaweedfs:
    image: chrislusf/seaweedfs:latest
    container_name: kkengine_seaweedfs
    # ... (full seaweedfs config from example/)
{{end}}

{{if .EnableCaddy}}
  caddy:
    image: caddy:alpine
    container_name: kkengine_caddy
    # ... (full caddy config from example/)
{{end}}

networks:
  kkengine_net:
    # ...

volumes:
  redis_data:
{{if .EnableCaddy}}
  caddy_data:
  caddy_config:
{{end}}
```

---

### Fix #3: Strengthen Test Coverage

**Update pkg/templates/embed_test.go**:

```go
// Re-enable YAML validation
func TestValidateYAML(t *testing.T) {
    // Remove t.Skip() line

    // Add line count validation
    lines := strings.Count(rendered, "\n")
    if lines < 50 {
        t.Errorf("docker-compose.yml has too few lines: %d (expected ~60+)", lines)
    }

    // Add literal \n detection
    if strings.Contains(rendered, "\\n") {
        t.Error("docker-compose.yml contains literal \\n escape sequences")
    }
}

// Add hex dump test
func TestNoLiteralEscapes(t *testing.T) {
    for _, tmpl := range []string{"docker-compose.yml", "env"} {
        content, _ := templateFS.ReadFile(tmpl + ".tmpl")
        if bytes.Contains(content, []byte("\\n")) {
            t.Errorf("%s.tmpl contains literal \\n escape sequences", tmpl)
        }
    }
}
```

---

### Fix #4: Environment Variable Substitution (Future Enhancement)

**Current approach**: Templates render passwords directly into docker-compose.yml
```yaml
MYSQL_PASSWORD: testpassword123  # ← Hardcoded in file
```

**Better approach**: Use environment variable placeholders
```yaml
MYSQL_PASSWORD: ${DB_PASSWORD}  # ← References .env file
```

**Scope**: OUT OF SCOPE for this fix (architectural change). Current template variable approach is acceptable for Phase 1.

---

## Test Validation Plan

After applying fixes, verify:

### 1. Template Integrity
```bash
# Line counts should be reasonable
wc -l pkg/templates/*.tmpl

# No "very long lines" warnings
file pkg/templates/*.tmpl

# No literal \n sequences in critical templates
grep -n '\\n' pkg/templates/docker-compose.yml.tmpl
grep -n '\\n' pkg/templates/env.tmpl
```

### 2. Rendered Output
```bash
# Run kk init in test directory
cd test/ && ../kk init

# Validate YAML syntax
docker-compose config --quiet || echo "YAML INVALID"

# Count services
grep -c "^  [a-z]" docker-compose.yml  # Should be 3-5 depending on flags

# Verify no literal \n
cat -A docker-compose.yml | grep '\\n' && echo "FAIL: Literal escapes found"
```

### 3. Conditional Logic
```bash
# Test with EnableSeaweedFS=false, EnableCaddy=false
# Should generate 3 services: kkengine, db, redis

# Test with EnableSeaweedFS=true, EnableCaddy=true
# Should generate 5 services: kkengine, db, redis, seaweedfs, caddy
```

### 4. Unit Tests
```bash
go test ./pkg/templates -v -run TestValidateYAML      # Should PASS (currently skipped)
go test ./pkg/templates -v -run TestGoldenFiles       # Should PASS
go test ./pkg/templates -v -run TestAllConfigCombinations  # Should PASS
```

---

## Preventive Measures

### 1. Pre-commit Hook
Add validation to `.git/hooks/pre-commit`:
```bash
#!/bin/bash
# Prevent committing templates with literal \n
if git diff --cached --name-only | grep -q '\.tmpl$'; then
    for file in $(git diff --cached --name-only | grep '\.tmpl$'); do
        if grep -q '\\n' "$file"; then
            echo "ERROR: $file contains literal \\n escape sequences"
            exit 1
        fi
    done
fi
```

### 2. CI Validation
Add to GitHub Actions workflow:
```yaml
- name: Validate Templates
  run: |
    # Check for literal escapes
    ! grep -r '\\n' pkg/templates/*.tmpl

    # Validate line counts
    test $(wc -l < pkg/templates/docker-compose.yml.tmpl) -gt 50
    test $(wc -l < pkg/templates/env.tmpl) -gt 10
```

### 3. Documentation
Update `docs/template-sync-process.md`:
```markdown
## Template Sync Process

**CRITICAL**: When copying example/ files to pkg/templates/*.tmpl:

❌ NEVER use: echo, JSON encoding, or tools that escape newlines
✅ ALWAYS use: cp, cat, or direct file editing

**Verification after sync**:
```bash
wc -l pkg/templates/*.tmpl  # Should show realistic line counts
file pkg/templates/*.tmpl   # Should NOT say "very long lines"
```

---

## Summary of Deliverables

### Immediate Actions Required:
1. ✓ Root cause identified: Literal `\n` in templates
2. ⚠ Fix template files (docker-compose.yml.tmpl, env.tmpl)
3. ⚠ Add conditional logic for optional services
4. ⚠ Re-enable YAML validation tests
5. ⚠ Add preventive CI checks

### Timeline Estimate:
- Fix #1 (Replace templates): 15 min
- Fix #2 (Add conditionals): 20 min
- Fix #3 (Test updates): 15 min
- Fix #4 (CI/hooks): 10 min
- **Total**: ~60 minutes

---

## Resolved Questions

### Q1: How did literal `\n` get introduced?

**Answer**: Issue exists since initial commit `14cffdf feat: implement Phase 01 - Core Foundation`

**Evidence**:
```bash
$ git log --follow -- pkg/templates/docker-compose.yml.tmpl
14cffdf feat: implement Phase 01 - Core Foundation

$ git show 14cffdf:pkg/templates/docker-compose.yml.tmpl | xxd | head -3
00000000: 7665 7273 696f 6e3a 2027 332e 3827 5c6e  version: '3.8'\n
00000010: 7365 7276 6963 6573 3a5c 6e20 206b 6b65  services:\n  kke
```

**Root cause**: Files were committed with literal `\n` from the start. Likely caused by:
- Using `echo` with `-e` flag during file creation
- JSON/string processing that escaped newlines
- Copy-paste from editor that interpreted escapes

**Impact**: All generated files since project start are broken.

---

### Q2: Are golden test files also affected?

**Answer**: YES - Golden files have identical corruption

**Evidence**:
```bash
$ wc -l pkg/templates/testdata/golden/*.golden
    3 Caddyfile.golden          # ✓ CORRECT
    1 docker-compose.yml.golden # ✗ BROKEN (should be ~60)
    1 env.golden                # ✗ BROKEN (should be ~17)
   20 kkfiler.toml.golden       # ✓ CORRECT
   20 kkphp.conf.golden         # ✓ CORRECT

$ file pkg/templates/testdata/golden/docker-compose.yml.golden
ASCII text, with very long lines (1439)  # ← Same issue as template
```

**Why tests pass**: TestGoldenFiles compares broken template output against broken golden files → perfect match → test passes incorrectly.

---

### Q3: Why did tests pass despite broken output?

**Multiple contributing factors**:

1. **Golden file corruption**: Expected output also contains `\n`, so comparison passes
2. **Permissive assertions**: `strings.Contains()` matches literal `\n` strings
3. **Skipped validation**: YAML parser test explicitly disabled
4. **No structural checks**: No line count, no hex validation, no YAML parsing

**Example from embed_test.go:36**:
```go
if !strings.Contains(string(content), "MYSQL_PASSWORD: testdbpassword") {
    t.Errorf("Rendered content mismatch")
}
```
This passes even if content is: `...MYSQL_PASSWORD: testdbpassword\n...` (single line)

---

## Unresolved Questions

1. **Should docker-compose.yml use env var substitution?**
   - Current: `MYSQL_PASSWORD: {{.DBPassword}}` renders to `MYSQL_PASSWORD: actual_password`
   - Alternative: `MYSQL_PASSWORD: ${DB_PASSWORD}` references .env file
   - Trade-offs: Current approach is simpler but less flexible
   - Recommendation: Keep current approach for Phase 1 simplicity

2. **Do users have broken configs in production?**
   - If `kk init` was run before this fix, users have invalid docker-compose.yml
   - Need migration strategy or warning in release notes
   - Consider adding `kk validate` command to check existing configs

---

**End of Report**

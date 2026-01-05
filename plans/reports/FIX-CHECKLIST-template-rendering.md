# Template Rendering Issues - Fix Checklist

**Priority**: P0 CRITICAL  
**Estimated time**: 60 minutes  
**Date**: 2026-01-05

---

## Pre-Fix Verification

Run these commands to confirm the issues:

```bash
# Verify broken templates (should show 1 line each)
wc -l pkg/templates/docker-compose.yml.tmpl pkg/templates/env.tmpl

# Verify literal \n escape sequences (should show matches)
xxd pkg/templates/docker-compose.yml.tmpl | head -1 | grep '5c6e'

# Verify missing services (should show empty output)
grep -c "seaweedfs\|caddy" test/docker-compose.yml
```

**Expected**: All checks confirm issues exist

---

## Fix #1: Replace Broken Templates ⚠️ CRITICAL

### Step 1.1: Backup current templates
```bash
cd /home/kkdev/kkcli
cp pkg/templates/docker-compose.yml.tmpl pkg/templates/docker-compose.yml.tmpl.broken
cp pkg/templates/env.tmpl pkg/templates/env.tmpl.broken
```

### Step 1.2: Copy example files
```bash
# Copy example files as new templates
cp example/docker-compose.yml pkg/templates/docker-compose.yml.tmpl
cp example/.env pkg/templates/env.tmpl
```

### Step 1.3: Replace hardcoded values with template variables

**In docker-compose.yml.tmpl**, replace:
- Line 30: `MYSQL_ROOT_PASSWORD: <actual_password>` → `MYSQL_ROOT_PASSWORD: {{.DBRootPassword}}`
- Line 33: `MYSQL_PASSWORD: <actual_password>` → `MYSQL_PASSWORD: {{.DBPassword}}`
- Line 51: `--requirepass <actual_password>` → `--requirepass {{.RedisPassword}}`
- Line 78: `WEED_MYSQL_HOSTNAME: <value>` → `WEED_MYSQL_HOSTNAME: ${DB_HOSTNAME}`
- Line 79: `WEED_MYSQL_PORT: <value>` → `WEED_MYSQL_PORT: ${DB_PORT}`
- Line 80: `WEED_MYSQL_USERNAME: <value>` → `WEED_MYSQL_USERNAME: ${DB_USERNAME}`
- Line 81: `WEED_MYSQL_PASSWORD: <value>` → `WEED_MYSQL_PASSWORD: ${DB_PASSWORD}`
- Line 82: `WEED_MYSQL_DATABASE: <value>` → `WEED_MYSQL_DATABASE: ${DB_SEAWEEDFS}`

**In env.tmpl**, replace:
- `DB_PASSWORD=<value>` → `DB_PASSWORD={{.DBPassword}}`
- `DB_ROOT_PASSWORD=<value>` → `DB_ROOT_PASSWORD={{.DBRootPassword}}`
- `REDIS_PASSWORD=<value>` → `REDIS_PASSWORD={{.RedisPassword}}`
- `DOMAIN=<value>` → `DOMAIN={{.Domain}}`

### Step 1.4: Verify fix
```bash
# Should show ~126 lines for docker-compose, ~17 for env
wc -l pkg/templates/docker-compose.yml.tmpl pkg/templates/env.tmpl

# Should be empty (no literal \n)
xxd pkg/templates/docker-compose.yml.tmpl | grep '5c6e'
xxd pkg/templates/env.tmpl | grep '5c6e'

# Should show template variables
grep '{{\.DBPassword}}' pkg/templates/docker-compose.yml.tmpl
grep '{{\.DBPassword}}' pkg/templates/env.tmpl
```

**Expected**: Line counts correct, no literal escapes, template vars present

---

## Fix #2: Add Template Conditionals ✨ ENHANCEMENT

### Step 2.1: Add conditionals for SeaweedFS service

Edit `pkg/templates/docker-compose.yml.tmpl`:

**Find line 62** (seaweedfs service start):
```yaml
  seaweedfs:
```

**Add before it**:
```yaml
{{if .EnableSeaweedFS}}
```

**Find line 96** (seaweedfs service end, before caddy):
```yaml
      start_period: 50s

  caddy:
```

**Add after seaweedfs section**:
```yaml
      start_period: 50s
{{end}}

  caddy:
```

### Step 2.2: Add conditionals for Caddy service

**Find line 98** (caddy service start):
```yaml
  caddy:
```

**Add before it**:
```yaml
{{if .EnableCaddy}}
```

**Find line 114** (caddy service end):
```yaml
    depends_on:
      - kkengine

networks:
```

**Add after caddy section**:
```yaml
    depends_on:
      - kkengine
{{end}}

networks:
```

### Step 2.3: Add conditionals for Caddy volumes

**Find volumes section** (around line 121):
```yaml
volumes:
  redis_data:
  caddy_data:
  caddy_config:
```

**Replace with**:
```yaml
volumes:
  redis_data:
{{if .EnableCaddy}}
  caddy_data:
  caddy_config:
{{end}}
```

### Step 2.4: Verify conditionals
```bash
# Should show 3 matches for EnableSeaweedFS (open + close + volumes check if needed)
grep -c 'EnableSeaweedFS' pkg/templates/docker-compose.yml.tmpl

# Should show 2 matches for EnableCaddy (open + close)
grep -c 'EnableCaddy' pkg/templates/docker-compose.yml.tmpl
```

---

## Fix #3: Regenerate Golden Test Files 🧪 QUALITY

### Step 3.1: Run golden file generator
```bash
cd pkg/templates/testdata
go run generate_golden.go
```

### Step 3.2: Verify golden files
```bash
# Should show realistic line counts
wc -l pkg/templates/testdata/golden/*.golden

# Should be empty (no literal \n)
xxd pkg/templates/testdata/golden/docker-compose.yml.golden | grep '5c6e'
xxd pkg/templates/testdata/golden/env.golden | grep '5c6e'

# Should include seaweedfs and caddy services
grep -c "seaweedfs\|caddy" pkg/templates/testdata/golden/docker-compose.yml.golden
```

**Expected**: 
- docker-compose.yml.golden: ~126 lines
- env.golden: ~17 lines
- No literal escapes
- Contains seaweedfs and caddy

### Step 3.3: Enable YAML validation test

Edit `pkg/templates/embed_test.go`, find line 233:
```go
func TestValidateYAML(t *testing.T) {
    t.Skip("Skipping YAML validation - docker-compose.yml.tmpl needs proper newlines (out of scope for Phase 1)")
```

**Replace with**:
```go
func TestValidateYAML(t *testing.T) {
    // YAML validation now enabled after fixing template newlines
```

### Step 3.4: Add hex dump validation test

Edit `pkg/templates/embed_test.go`, add new test after TestValidateYAML:

```go
// TestNoLiteralEscapes ensures templates don't contain literal \n escape sequences
func TestNoLiteralEscapes(t *testing.T) {
    criticalTemplates := []string{"docker-compose.yml", "env"}
    
    for _, tmpl := range criticalTemplates {
        content, err := templateFS.ReadFile(tmpl + ".tmpl")
        if err != nil {
            t.Fatalf("Failed to read %s.tmpl: %v", tmpl, err)
        }
        
        // Check for literal \n (hex: 5c6e)
        if bytes.Contains(content, []byte("\\n")) {
            t.Errorf("%s.tmpl contains literal \\n escape sequences", tmpl)
        }
        
        // Check line count is reasonable
        lines := bytes.Count(content, []byte("\n"))
        minLines := map[string]int{
            "docker-compose.yml": 50,
            "env": 10,
        }
        
        if lines < minLines[tmpl] {
            t.Errorf("%s.tmpl has only %d lines (expected at least %d)", tmpl, lines, minLines[tmpl])
        }
    }
}
```

---

## Fix #4: Add CI Validation 🔒 PREVENTION

### Step 4.1: Create GitHub Actions workflow

Create `.github/workflows/validate-templates.yml`:

```yaml
name: Validate Templates

on:
  pull_request:
    paths:
      - 'pkg/templates/*.tmpl'
      - 'pkg/templates/testdata/golden/*'
  push:
    paths:
      - 'pkg/templates/*.tmpl'
      - 'pkg/templates/testdata/golden/*'

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Check for literal escape sequences
        run: |
          echo "Checking for literal \n escape sequences in templates..."
          if grep -r '\\n' pkg/templates/*.tmpl; then
            echo "❌ Found literal \\n escape sequences in templates"
            exit 1
          fi
          echo "✅ No literal escape sequences found"
      
      - name: Validate line counts
        run: |
          echo "Validating template line counts..."
          LINES=$(wc -l < pkg/templates/docker-compose.yml.tmpl)
          if [ "$LINES" -lt 50 ]; then
            echo "❌ docker-compose.yml.tmpl has only $LINES lines (expected 50+)"
            exit 1
          fi
          
          LINES=$(wc -l < pkg/templates/env.tmpl)
          if [ "$LINES" -lt 10 ]; then
            echo "❌ env.tmpl has only $LINES lines (expected 10+)"
            exit 1
          fi
          echo "✅ Line counts valid"
      
      - name: Validate YAML syntax
        run: |
          echo "Installing yq for YAML validation..."
          sudo wget -qO /usr/local/bin/yq https://github.com/mikefarah/yq/releases/latest/download/yq_linux_amd64
          sudo chmod +x /usr/local/bin/yq
          
          echo "Validating golden docker-compose.yml syntax..."
          yq eval pkg/templates/testdata/golden/docker-compose.yml.golden > /dev/null
          echo "✅ YAML syntax valid"
```

### Step 4.2: Add pre-commit hook (optional)

Create `.git/hooks/pre-commit`:

```bash
#!/bin/bash
# Prevent committing templates with literal \n

TEMPLATES=$(git diff --cached --name-only | grep '\.tmpl$')

if [ -n "$TEMPLATES" ]; then
    echo "Validating templates..."
    for file in $TEMPLATES; do
        if grep -q '\\n' "$file"; then
            echo "❌ ERROR: $file contains literal \\n escape sequences"
            echo "Fix: Replace literal \\n with actual newlines"
            exit 1
        fi
    done
    echo "✅ Templates valid"
fi
```

```bash
chmod +x .git/hooks/pre-commit
```

---

## Post-Fix Testing

### Test 1: Run unit tests
```bash
cd /home/kkdev/kkcli
go test ./pkg/templates -v
```

**Expected**: All tests PASS (including TestValidateYAML)

---

### Test 2: Test with minimal config (no optional services)
```bash
cd test/
rm -rf .env docker-compose.yml Caddyfile kkfiler.toml kkphp.conf

# Run with EnableSeaweedFS=false, EnableCaddy=false
# (Modify cmd/init.go temporarily or use test binary)
../kk init --db-password=testpass123 --redis-password=redispass456 --domain=test.local

# Verify output
docker-compose config --quiet || echo "❌ YAML INVALID"
grep -c "^  [a-z]" docker-compose.yml  # Should be 3 (kkengine, db, redis)
cat -A docker-compose.yml | grep '\\n' && echo "❌ Literal escapes found"
```

**Expected**:
- YAML valid
- 3 services (kkengine, db, redis)
- No literal escapes
- No seaweedfs or caddy

---

### Test 3: Test with all services enabled
```bash
cd test/
rm -rf .env docker-compose.yml Caddyfile kkfiler.toml kkphp.conf

# Run with EnableSeaweedFS=true, EnableCaddy=true (default)
../kk init --db-password=testpass123 --redis-password=redispass456 --domain=test.local

# Verify output
docker-compose config --quiet || echo "❌ YAML INVALID"
grep -c "^  [a-z]" docker-compose.yml  # Should be 5 (all services)
grep -q "seaweedfs:" docker-compose.yml && echo "✅ SeaweedFS present"
grep -q "caddy:" docker-compose.yml && echo "✅ Caddy present"
```

**Expected**:
- YAML valid
- 5 services (kkengine, db, redis, seaweedfs, caddy)
- Both optional services present

---

### Test 4: Integration test
```bash
cd test/
docker-compose up -d
docker-compose ps  # Should show all services running
docker-compose down
```

**Expected**: All services start successfully

---

## Completion Checklist

- [ ] Fix #1: Replaced docker-compose.yml.tmpl and env.tmpl
- [ ] Fix #1: Verified line counts and no literal escapes
- [ ] Fix #1: Added template variables ({{.DBPassword}}, etc.)
- [ ] Fix #2: Added {{if .EnableSeaweedFS}} conditional
- [ ] Fix #2: Added {{if .EnableCaddy}} conditional
- [ ] Fix #2: Added conditional for caddy volumes
- [ ] Fix #3: Regenerated golden test files
- [ ] Fix #3: Enabled TestValidateYAML
- [ ] Fix #3: Added TestNoLiteralEscapes
- [ ] Fix #4: Created GitHub Actions workflow
- [ ] Fix #4: Added pre-commit hook (optional)
- [ ] All unit tests pass
- [ ] Minimal config test passes (no optional services)
- [ ] Full config test passes (all services)
- [ ] Integration test passes (docker-compose up)

---

## Rollback Plan

If fixes cause issues:

```bash
# Restore broken templates
cd /home/kkdev/kkcli
cp pkg/templates/docker-compose.yml.tmpl.broken pkg/templates/docker-compose.yml.tmpl
cp pkg/templates/env.tmpl.broken pkg/templates/env.tmpl

# Restore original tests
git checkout pkg/templates/embed_test.go
```

---

**Report**: debugger-260105-1726-template-rendering-investigation.md  
**Summary**: SUMMARY-template-rendering-issues.md

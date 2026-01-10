# Phase 04: Testing

## Context Links
- Parent: [plan.md](./plan.md)
- Depends on: [Phase 03](./phase-03-command-updates.md)

## Overview
- **Priority**: Medium
- **Status**: Pending
- **Description**: Test all commands in both English and Vietnamese

## Key Insights
- Need to test both language modes
- Test with Docker running and not running
- Verify terminal width handling
- Check diacritics display correctly in Vietnamese

## Requirements

### Test Scenarios

| Command | Scenario | Expected Result |
|---------|----------|-----------------|
| `kk init` | Fresh directory | Step 1-5 wizard, summary table, next steps box |
| `kk init` | Existing compose | Overwrite prompt, then wizard flow |
| `kk status` | No services | "No services running" message |
| `kk status` | Services running | Boxed table with colored status |
| `kk start` | Preflight pass | Green checkmarks, then status table |
| `kk start` | Preflight fail | Red X marks in table |

### Language Tests

| Test | Command | Expected |
|------|---------|----------|
| English default | `kk status` | "Service Status", "Service", "Status", etc. |
| Vietnamese | Set lang=vi, `kk status` | "Trạng thái dịch vụ", "Dịch vụ", etc. |
| Vietnamese diacritics | `kk init` in Vietnamese | Proper diacritics: Kiểm tra, Trạng thái |

## Implementation Steps

### 1. Build and Lint

```bash
cd /home/kkdev/kkcli
go build ./...
golangci-lint run
```

### 2. Manual Test: kk init (English)

```bash
# Clean directory
mkdir -p /tmp/kk-test && cd /tmp/kk-test

# Run init
kk init

# Expected output:
# Step 1/5: Docker Check
# ✅ Docker is ready
# Step 2/5: Language Selection
# [Language selector]
# Step 3/5: Configuration Options
# [SeaweedFS/Caddy prompts]
# Step 4/5: Generate Files
# [Spinner: Generating...]
# Step 5/5: Complete
#
# Configuration Summary
# ┌─────────────┬───────────────┐
# │ Setting     │ Value         │
# ├─────────────┼───────────────┤
# │ SeaweedFS   │ ✓ Enabled     │
# │ Caddy       │ ✓ Enabled     │
# │ Domain      │ localhost     │
# └─────────────┴───────────────┘
#
# Created Files
# ✅ docker-compose.yml
# ✅ .env
# ...
```

### 3. Manual Test: kk init (Vietnamese)

```bash
# Select Vietnamese during init
# Expected: Vietnamese strings with proper diacritics
# "Kiểm tra Docker", "Tùy chọn cấu hình", "Trạng thái"
```

### 4. Manual Test: kk status

```bash
cd /tmp/kk-test
kk start
kk status

# Expected:
# Service Status
# ┌───────────┬────────────┬─────────┬────────────────────┐
# │ Service   │ Status     │ Health  │ Ports              │
# ├───────────┼────────────┼─────────┼────────────────────┤
# │ kkengine  │ ● Running  │ healthy │ 8019:80            │
# │ db        │ ● Running  │ healthy │ 3307:3306          │
# └───────────┴────────────┴─────────┴────────────────────┘
#
# Access Information
# ┌───────────┬─────────────────────────┐
# │ Service   │ URL                     │
# ├───────────┼─────────────────────────┤
# │ kkengine  │ http://localhost:8019   │
# └───────────┴─────────────────────────┘
```

### 5. Manual Test: kk start (Preflight)

```bash
# With Docker not running (to test failure display)
sudo systemctl stop docker
kk start

# Expected:
# Running preflight checks...
# ┌────────────────────┬────────────────────┐
# │ Check              │ Result             │
# ├────────────────────┼────────────────────┤
# │ Docker Installed   │ ✓ Pass             │
# │ Docker Daemon      │ ✗ Docker not running│
# └────────────────────┴────────────────────┘

# Restart Docker
sudo systemctl start docker
```

### 6. Terminal Width Test

```bash
# Narrow terminal
stty cols 60
kk status
# Verify table wraps correctly or truncates gracefully
```

## Todo List

- [ ] Run `go build ./...`
- [ ] Run `golangci-lint run`
- [ ] Test `kk init` in English
- [ ] Test `kk init` in Vietnamese
- [ ] Test `kk status` with services
- [ ] Test `kk status` without services
- [ ] Test `kk start` preflight pass
- [ ] Test `kk start` preflight fail
- [ ] Test narrow terminal

## Success Criteria

- [ ] Build passes with no errors
- [ ] Lint passes with no warnings
- [ ] All manual tests pass
- [ ] Vietnamese diacritics display correctly
- [ ] Tables render correctly in different terminal widths

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| Docker not available for testing | Medium | Use mock/stub for CI |
| Terminal encoding issues | Low | Test on multiple terminals |

## Security Considerations

- No security concerns - testing only

## Next Steps

- Commit changes
- Update documentation if needed

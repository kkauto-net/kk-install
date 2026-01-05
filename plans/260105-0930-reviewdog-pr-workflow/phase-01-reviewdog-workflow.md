# Phase 01: Create Reviewdog Workflow

## Overview

| Field | Value |
|-------|-------|
| Date | 2026-01-05 |
| Priority | P2 |
| Status | pending |
| Effort | 1h |
| Dependencies | None |

## Context

- Plan: [plan.md](./plan.md)
- Existing CI: `.github/workflows/ci.yml`

## Key Insights

### Reviewdog Benefits
- Inline PR comments instead of console-only output
- `github-pr-review` reporter creates review comments on exact lines
- `filter_mode: added` - only flags new/changed lines (reduces noise)
- Official actions handle tool installation automatically

### Action Versions
- `reviewdog/action-golangci-lint@v1` - Go linting with reviewdog
- `reviewdog/action-shellcheck@v1` - Shell script linting

### Reporter Modes
| Mode | Description |
|------|-------------|
| `github-pr-review` | PR review comments (recommended) |
| `github-pr-check` | Check annotations |
| `github-check` | Check run annotations |

## Requirements

1. New workflow file: `.github/workflows/reviewdog.yml`
2. Trigger: `pull_request` events only
3. Two jobs: go-lint, shell-lint
4. Reporter: `github-pr-review`
5. Filter mode: `added` (only changed lines)

## Architecture Decisions

### AD-01: Separate Workflow File
- **Decision**: Create new `reviewdog.yml` instead of modifying `ci.yml`
- **Rationale**:
  - Different purposes (PR feedback vs branch protection)
  - ci.yml runs on push+PR, reviewdog only on PR
  - Easier to maintain/disable independently

### AD-02: Two Parallel Jobs
- **Decision**: Separate `go-lint` and `shell-lint` jobs
- **Rationale**: Run in parallel, fail independently, clear logs

### AD-03: Go Version from go.mod
- **Decision**: Use `go-version-file: 'go.mod'` (consistent with ci.yml)
- **Rationale**: Single source of truth for Go version

## Related Code Files

### .github/workflows/ci.yml (reference)
```yaml
# Existing lint job uses golangci/golangci-lint-action@v4
# This outputs to console only, no PR comments
lint:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v5
      with:
        go-version-file: 'go.mod'
    - uses: golangci/golangci-lint-action@v4
```

### scripts/install.sh
- 143 lines bash script
- Target for shellcheck linting

## Implementation Steps

### Step 1: Create workflow file

Create `.github/workflows/reviewdog.yml`:

```yaml
name: reviewdog

on:
  pull_request:
    branches: [main]

permissions:
  contents: read
  pull-requests: write

jobs:
  go-lint:
    name: Go Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version-file: 'go.mod'

      - name: golangci-lint with reviewdog
        uses: reviewdog/action-golangci-lint@v1
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          reporter: github-pr-review
          filter_mode: added
          fail_level: warning
          level: warning

  shell-lint:
    name: Shell Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: shellcheck with reviewdog
        uses: reviewdog/action-shellcheck@v1
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          reporter: github-pr-review
          filter_mode: added
          fail_level: warning
          level: warning
          path: "scripts"
          pattern: "*.sh"
```

### Step 2: Verify workflow syntax

```bash
# If actionlint available
actionlint .github/workflows/reviewdog.yml

# Or use GitHub's workflow validation on push
```

### Step 3: Test with PR

1. Create test branch
2. Add intentional lint issue (e.g., unused var in Go, unquoted var in shell)
3. Create PR to main
4. Verify reviewdog comments appear

## Todo List

- [ ] Create `.github/workflows/reviewdog.yml`
- [ ] Verify YAML syntax
- [ ] Test with sample PR
- [ ] Confirm inline comments appear
- [ ] Update documentation if needed

## Success Criteria

| Criteria | Validation |
|----------|------------|
| Workflow triggers on PR | Check Actions tab |
| Go lint comments appear | Create PR with lint issue |
| Shell lint comments appear | Create PR with shell issue |
| Only changed lines flagged | Check filter_mode working |
| Jobs run in parallel | Check Actions timeline |

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| GITHUB_TOKEN permissions | Medium | Explicit `permissions` block |
| golangci-lint version mismatch | Low | Uses latest stable |
| False positives in existing code | Low | `filter_mode: added` |

## Security Considerations

1. **Token Scope**: Uses default `GITHUB_TOKEN` with minimal permissions
2. **Permissions Block**: Explicitly declares `contents: read`, `pull-requests: write`
3. **No Secrets Exposure**: Only built-in token used
4. **Fork PRs**: reviewdog handles fork PRs safely (may have limited permissions)

## Next Steps

After implementation:
1. Monitor first few PRs for noise level
2. Adjust `fail_level` if needed (error vs warning)
3. Consider adding `.golangci.yml` for custom rules
4. Consider adding `.shellcheckrc` for exclusions

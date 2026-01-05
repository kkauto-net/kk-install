---
title: "Reviewdog PR Workflow"
description: "Add reviewdog GitHub Actions for automated PR code review (Go + Shell)"
status: pending
priority: P2
effort: 1h
branch: main
tags: [ci, github-actions, reviewdog, golangci-lint, shellcheck]
created: 2026-01-05
---

# Reviewdog PR Workflow Implementation

## Overview

Add reviewdog-based GitHub Actions workflow for automated PR reviews. Provides inline comments on PRs for Go code (golangci-lint) and shell scripts (shellcheck).

## Objectives

1. Automated PR code review with inline comments
2. Go linting via reviewdog/action-golangci-lint
3. Shell linting via reviewdog/action-shellcheck
4. Only check changed lines (filter_mode: added)

## Phases

| # | Phase | Status | Effort | File |
|---|-------|--------|--------|------|
| 1 | Create reviewdog.yml workflow | pending | 1h | [phase-01](./phase-01-reviewdog-workflow.md) |

## Quick Reference

- **Trigger**: pull_request only
- **Reporter**: github-pr-review (inline comments)
- **Go Action**: reviewdog/action-golangci-lint@v1
- **Shell Action**: reviewdog/action-shellcheck@v1
- **Target Files**: `*.go`, `scripts/*.sh`

## Dependencies

- Existing CI workflow: `.github/workflows/ci.yml`
- Shell script: `scripts/install.sh`
- Go version: from `go.mod`

## Success Criteria

- [ ] Workflow triggers on PRs only
- [ ] Go lint issues appear as PR comments
- [ ] Shell lint issues appear as PR comments
- [ ] Only changed lines flagged (filter_mode: added)
- [ ] No conflicts with existing ci.yml

## Notes

- Existing ci.yml has golangci-lint job but uses `golangci/golangci-lint-action` (console output only)
- New reviewdog workflow provides PR review comments (better DX)
- Keep both workflows: ci.yml for branch protection, reviewdog.yml for PR feedback

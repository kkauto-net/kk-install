---
title: "Load .env on kk init"
description: "Pre-fill init form from existing .env and backup with timestamp"
status: implemented
priority: P2
effort: 2h
branch: main
tags: [init, ux, env]
created: 2026-01-12
reviewed: 2026-01-12
review_score: 8/10
---

# Load .env Values on kk init

## Overview

Enhance `kk init` to load existing `.env` values as default form inputs and backup files with timestamp format.

## Problem

When running `kk init` in a directory with existing `.env`:
1. User must re-enter all credentials manually → time-consuming, error-prone
2. Backup uses simple `.bak` suffix → overwrites previous backups

## Solution

1. Parse existing `.env` and pre-fill form fields
2. Backup with timestamp: `filename-Ymd-His.bak`

## Implementation Phases

| Phase | Description | Status | Est |
|-------|-------------|--------|-----|
| [Phase 01](phase-01-env-preload.md) | Implement env loading and backup with timestamp | completed | 2h |

## Files to Modify

- `cmd/init.go` - Add loadExistingEnv(), update backupExistingConfigs()
- `pkg/templates/embed.go` - Remove duplicate backup logic

## Success Criteria

- [x] Form pre-fills values from existing .env
- [x] Backup files have unique timestamps
- [x] Missing fields auto-generate random values
- [x] Invalid secrets (too short) regenerated

## Code Review Summary

**Date:** 2026-01-12 14:28
**Score:** 8/10
**Status:** Production ready with minor improvements recommended

**Key findings:**
- ✓ Security: No credential exposure, proper .env handling
- ✓ Architecture: DRY violation fixed, clean separation
- ⚠ HIGH: ENV parser lacks quote handling for quoted values
- ⚠ HIGH: No validation on loaded env values before use
- ⚠ MEDIUM: Silent backup failures, timestamp collision risk

**Full review:** [code-reviewer-260112-1428-env-preload-init.md](../reports/code-reviewer-260112-1428-env-preload-init.md)

## Related Documents

- [Brainstorm Report](../reports/brainstorm-260112-1357-env-preload-init.md)

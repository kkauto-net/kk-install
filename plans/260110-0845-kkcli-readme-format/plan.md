---
title: "kkcli README/Docs Format Improvements"
description: "Rewrite README in English with Modern OSS style and fix command descriptions in docs"
status: done
priority: P2
effort: 1.5h
branch: main
tags: [docs, readme, i18n]
created: 2026-01-10
---

# kkcli README/Docs Format Improvements

## Overview

Transform kkcli documentation to be more professional:
1. Rewrite README.md in English with Modern OSS style
2. Fix `kk` command descriptions in docs/ files (Vietnamese → English)

## Phases

| # | Phase | Status | Effort | Link |
|---|-------|--------|--------|------|
| 1 | Rewrite README.md | DONE | 1h | [phase-01](./phase-01-rewrite-readme.md) |
| 2 | Fix docs command descriptions | DONE | 0.5h | [phase-02](./phase-02-fix-docs-commands.md) |

## Dependencies

- GitHub repo: `kkauto-net/kk-install`
- CI workflow name: `CI`

## Files Affected

- `/home/kkdev/kkcli/README.md` (rewrite)
- `/home/kkdev/kkcli/docs/project-overview-pdr.md` (modify)
- `/home/kkdev/kkcli/docs/codebase-summary.md` (modify)

## Success Criteria

- [x] README renders professionally on GitHub
- [x] All badges display correctly
- [x] Quick install command works
- [x] Command descriptions consistent EN/VI

---
date: 2026-05-26
type: implementation
plan: ../../plans/260525-2327-kk-update-image-version/plan.md
---

# KK Update Image Version Implementation

## Context

Implemented `kk update` image version visibility and recreate behavior from the approved plan. Scope stayed `kk update` only; `kk n8n update` command flow remains on legacy pull-output parsing.

## What Changed

- Added Compose service image discovery and container name resolution.
- Added Docker image identity inspection with repo digest preference and image ID fallback.
- Added running-container image ID comparison so pulled-but-not-recreated images still require apply.
- Wired `kk update` to inspect before pull, pull, inspect after pull, compare identities, show current/new table, confirm, then `ForceRecreate`.
- Added kk-update-specific recreate prompt/cancel messages so n8n shared update messages stay unchanged.
- Updated architecture/code docs and README command summary.

## Validation

- `go test -v ./cmd ./pkg/updater ./pkg/compose ./pkg/ui`
- `go test ./...`
- `go build ./...`
- `make fmt`
- `make test-smoke`
- Code review and tester subagents reported no Critical/High findings after fixes.

## Decisions

- No separate `docker compose restart`; recreate remains `docker compose up -d --force-recreate`.
- If user cancels after pull, a later `kk update` still detects stale running containers by image ID.
- Missing running container is ignored during update detection; local image diff still handles new pulls.

## Next

- Optional: commit with `feat(update): add image version update support`.

## Unresolved Questions

None.

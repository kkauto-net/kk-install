# KK Update Image Version Planning

---
date: 2026-05-25
type: planning
plan: ../../plans/260525-2327-kk-update-image-version/plan.md
---

## Context

User reported `kk` lacks current/new Docker image version display and `kk update` should check next version then restart/recreate to apply updates.

## What Happened

- Confirmed scope through brainstorm: use Docker image digest/ID, not Docker Engine version or mutable tag string.
- Confirmed restart behavior: keep prompt before recreate; `-f` skips prompt.
- Confirmed apply scope: `kk update` only, not `kk n8n update`.
- Created brainstorm report at `docs/260525-2325-kk-update-image-version-brainstorm.md`.
- Created implementation plan at `plans/260525-2327-kk-update-image-version/plan.md` with 4 phase files.

## Decisions

- Compare image identities before and after `docker compose pull`.
- Prefer repo digest, fallback image ID.
- Keep `docker compose up -d --force-recreate`; do not add separate restart.
- Preserve `ParsePullOutput` for `kk n8n update` compatibility.

## Next

Run:

```bash
/ck:cook /home/kkdev/kkinstall/plans/260525-2327-kk-update-image-version/plan.md --auto
```

## Unresolved Questions

- None.

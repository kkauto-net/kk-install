# kk Update Image Version Brainstorm

---
date: 2026-05-25
status: approved
scope: kk-update-only
---

## Summary

Improve `kk update` UX by showing current and new Docker image identity before asking user to recreate services. Use image digest/ID, not tag string, because tags like `latest` are mutable and not real versions.

Decision: apply to `kk update` only. Do not change `kk n8n update` in this phase.

## Problem Statement

Current `kk update` flow:

1. Ensure project dir.
2. Run `docker compose pull`.
3. Parse pull output.
4. If update detected, show table.
5. Ask confirmation unless `-f`.
6. Run `docker compose up -d --force-recreate`.
7. Show health/status.

Problems:

- Update table has `Current` and `New` columns, but current parser does not populate `OldDigest` or `NewDigest`.
- User cannot see what changed before confirming restart/recreate.
- “Docker version” wording is ambiguous. For this use case, target identity is Docker image digest/ID, not Docker Engine or Compose version.
- Restart behavior exists as force recreate, but command messaging does not clearly state that recreate/restart applies new images.

## Requirements

- Show current image digest/ID and new image digest/ID for changed services/images.
- Keep prompt before applying restart/recreate.
- Keep `-f` behavior: skip confirmation.
- Do not add Docker Registry API calls.
- Do not apply to `kk n8n update` in this phase.
- Preserve no-Docker unit-testability for compare logic.

## Evaluated Approaches

| Approach | Pros | Cons | Verdict |
|---|---|---|---|
| Parse `docker compose pull` output deeper | Smallest command surface | Output varies by Docker/Compose version; cannot know pre-pull current identity reliably | Reject |
| Inspect local image identity before and after pull | Accurate for mutable tags; simple mental model; no registry auth/rate-limit | Requires extra inspect/images calls; image may be absent before pull | Recommend |
| Query Docker Registry API before pull | Can know remote latest before download | Over-engineered; registry-specific auth/rate limits; more failure modes | Reject for now |

## Recommended Design

Use local image identity comparison around pull.

Flow:

1. Parse `docker-compose.yml` to collect service images.
2. Inspect local identity for each image before pull.
3. Run `docker compose pull`.
4. Inspect local identity for each image after pull.
5. Compare before vs after.
6. Show changed rows as `Image | Current | New`.
7. If no changes, show images up to date.
8. If changes exist, ask confirmation unless `-f`.
9. On confirm, run existing `docker compose up -d --force-recreate`.
10. Show health/status.

Identity preference:

1. Prefer repo digest when available.
2. Fall back to image ID when digest unavailable.
3. Use `-` or `not present` when image does not exist before pull.

## Rationale

- Digest/ID is the real version boundary for Docker images using mutable tags.
- Before/after local inspect avoids brittle text parsing.
- No Registry API keeps implementation small and robust.
- Existing recreate command is enough to apply new images. Adding separate `docker compose restart` would be redundant and may increase downtime.

## Implementation Considerations

- Add small reusable logic in `pkg/updater` or `pkg/compose` for image identity collection and diffing.
- Keep Docker command execution behind testable functions, similar to existing `compose` package patterns.
- Update `cmd/update.go` only; leave `cmd/n8n_update.go` unchanged.
- Update UI text to explain recreate/restart before confirmation.
- Add unit tests for identity diff behavior:
  - unchanged image ID = no update
  - changed image ID = update row
  - missing before + present after = update row
  - inspect failure after pull = clear error

## Risks

- Some images may not expose repo digest locally. Image ID fallback mitigates.
- Compose files with services lacking `image` should be skipped or reported as unsupported.
- Image ID changes do not expose human semantic version. This is acceptable because current stack uses mutable tags.
- If Docker inspect behavior differs by platform, command output parser must be narrow and tested.

## Success Metrics

- `kk update` shows non-empty `Current` and `New` values when images change.
- `kk update` reports up-to-date when before/after identities match.
- User sees restart/recreate confirmation before services are changed.
- Existing update flow remains compatible with `-f`.
- Unit tests pass without Docker daemon.

## Next Steps

- Create implementation plan if needed.
- Then implement in a separate coding phase after plan approval.

## Unresolved Questions

- None.

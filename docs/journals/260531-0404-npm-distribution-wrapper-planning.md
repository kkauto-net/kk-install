# Npm Distribution Wrapper Planning

---
date: 2026-05-31
type: planning
plan: ../../plans/260531-0359-npm-distribution-wrapper/plan.md
---

## Context

The user asked to create an implementation plan for publishing `kkcli` on npm. This is planning-only work; no npm wrapper code or workflow implementation was created.

## What Happened

- Reviewed README, code standards, system architecture, deployment guide, codebase summary, release workflow, GoReleaser config, installer scripts, and existing plans.
- Scanned current plan history. Relevant release/checksum/test plans are completed, so no active blocking dependency was added.
- Checked npm registry names with `npm view @kkauto/kkcli` and `npm view kkcli`; both returned E404, which suggests no public package is visible but does not prove scope ownership.
- Created `plans/260531-0359-npm-distribution-wrapper/` with `plan.md` and four phase files.

## Decisions

- Recommended `@kkauto/kkcli` as the npm package name.
- Treat npm as a distribution wrapper, not a Go build system.
- Keep phase 1 Linux-only because current GoReleaser artifacts are Linux `amd64` and `arm64` only.
- Preserve fail-closed SHA256 verification using GitHub Release `checksums.txt`.
- Publish npm only after GitHub Release assets exist.

## Validation

- `git diff --check -- plans/260531-0359-npm-distribution-wrapper` passed.
- Active plan script could not persist because `CK_SESSION_ID` is not set; session todos were hydrated manually.

## Next

Use Cook to implement:

```bash
/ck:cook /home/kkdev/kkinstall/plans/260531-0359-npm-distribution-wrapper/plan.md
```

## Unresolved Questions

- Does the team control npm org/scope `@kkauto`?
- Should publishing use npm trusted publishing or `NPM_TOKEN`?
- Confirm final package name before implementation; `@kkauto/kkcli` is recommended.

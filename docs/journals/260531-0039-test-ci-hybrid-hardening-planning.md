# Test CI Hybrid Hardening Planning Journal

---
date: 2026-05-31
scope: planning
plan: ../../plans/260531-0000-test-ci-hybrid-hardening/plan.md
status: completed
---

## Context

Created implementation plan from approved hybrid staged Test/CI brainstorm.

## What Happened

- Confirmed current baseline: PR/main CI, reviewdog, template validation, nightly e2e, cleanup artifacts, release checksum integrity.
- Rejected strict PR e2e gate because it depends on Docker, network, image pulls, license secret, and external APIs.
- Chose hybrid staged design: deterministic PR, deeper main/schedule checks, real runtime nightly/manual, release integrity gate.
- Updated stale `plans/260105-0930-reviewdog-pr-workflow/plan.md` to completed because workflow already exists.
- Created `plans/260531-0000-test-ci-hybrid-hardening/` with one plan and five phase files.

## Decisions

- Add focused security tests before broader workflow changes.
- Prefer installer shell harness for checksum fail-closed branches.
- Trial security scanning outside required PR gates first.
- Keep e2e Compose nightly/manual only.
- Keep cleanup scoped to Actions artifacts, not release assets.

## Next

Run implementation through:

```bash
/ck:cook /home/kkdev/kkinstall/plans/260531-0000-test-ci-hybrid-hardening/plan.md
```

## Unresolved Questions

- Which security scanner should be trialed first: `govulncheck`, Trivy, or GitHub code scanning?
- Should repeated nightly e2e failures become explicit release blockers?

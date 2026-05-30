# Test CI Hybrid Hardening Completion

**Date**: 2026-05-31 02:56
**Severity**: Medium
**Component**: Test/CI, installer, selfupdate, e2e diagnostics
**Status**: Resolved

## What Happened

Completed `plans/260531-0000-test-ci-hybrid-hardening/plan.md` to 100. The implementation added selfupdate checksum regression tests, a source-safe and `curl | bash` compatible installer entrypoint guard, `scripts/install_test.sh` with 7 offline installer tests, a CI installer shell test step, scheduled-only `govulncheck`, and e2e diagnostics redaction for Compose `ps`/logs that fail-closes by deleting diagnostics when redaction fails. `docs/README` was updated too.

## The Brutal Truth

The first pass was not good enough. Review caught exactly the kind of CI theater we were trying to remove: a `curl | bash` guard that could break the documented install path, a no-checksum-tool case that could pass for the wrong reason, and e2e redaction that initially failed open. That is maddening because these are not exotic edge cases; they are the whole point of hardening.

## Technical Details

Validation now passes:

- `bash -n scripts/install.sh`
- `bash -n scripts/install_test.sh`
- `bash scripts/install_test.sh` → `PASS: 7 installer checksum tests`, stdin guard probe `RUN`
- `go test -v ./pkg/selfupdate`
- `go test -v ./...`
- `make test-smoke`
- `node $HOME/.claude/scripts/validate-docs.cjs docs/` → links OK 29 with existing warnings
- `git diff --check`

## What We Tried

- Rejected PR-gating real Docker e2e because it needs Docker, network pulls, secrets, and external service behavior.
- Chose deterministic offline installer/selfupdate checks in PR CI instead.
- Moved `govulncheck` to scheduled-only so it gives security signal without creating flaky PR blockage.
- Changed diagnostics redaction to fail closed: if redaction fails, delete diagnostics rather than upload possible secrets.

## Root Cause Analysis

We previously trusted shell installer and diagnostics paths without enough hostile-path testing. That was the fundamental mistake. The install script is both a sourced library and a streamed executable, so naïve entrypoint guards break one mode while fixing another. Diagnostics are also dangerous by default; logs should be treated as secret-bearing until proven otherwise.

## Lessons Learned

Test the documented invocation, not the happy local invocation. If README says `curl -sSL ... | bash`, CI needs a stdin-mode probe. Any checksum path must explicitly test missing tools and tampered artifacts. Any diagnostic artifact must be redacted before upload and fail closed.

## Next Steps

- Owner: implementer/release owner. Before commit, inspect `git status` and stage only intended workflow, docs, selfupdate, and installer files.
- Owner: CI maintainer. Watch the next scheduled `govulncheck` and nightly e2e run; fix real findings, do not turn them into ignored noise.
- Timeline: before merging this hardening branch.

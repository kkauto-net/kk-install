# npm Distribution Wrapper Completion

**Date**: 2026-05-31 04:54
**Severity**: Medium
**Component**: npm distribution / release automation
**Status**: Resolved with operational follow-up

## What Happened

Completed the `@kkauto/kkcli` npm distribution wrapper under `npm/kkcli`. The package is intentionally Linux-only (`x64`/`arm64`) because current GoReleaser output only publishes Linux artifacts. The wrapper exposes a `kk` bin shim, and `postinstall` downloads GitHub Release artifacts plus `checksums.txt`, then verifies the exact release filename SHA256 before installing the binary.

## The Brutal Truth

This was packaging work, but it had all the ways packaging can quietly become a supply-chain footgun. The exhausting part is that npm makes global install feel simple while the real risk is a shell script fetching tarballs from a moving release page at install time. We hardened the obvious traps, but no one should pretend this is equivalent to signed provenance yet.

## Technical Details

Validation passed: `npm test` reported `12/12`, `npm pack --dry-run` completed, installer shell tests passed, `go test -v ./...` passed, `make test-smoke` passed, workflow YAML parsed, docs validation passed with existing warnings, and `git diff --check` passed.

Security behavior now includes exact filename SHA256 verification against `checksums.txt`, unsafe archive rejection, timeout/size caps, and explicit rejection of archive traversal. There is still **no signature verification**; checksum verification only proves consistency with the release checksum file, not independent publisher identity.

CI now runs npm tests and npm pack. `release.yml` calls `publish-npm.yml` only when `NPM_PUBLISH_ENABLED=true`. The publish workflow supports manual dispatch, version sync, release asset polling, already-published skip, trusted publishing, and `NPM_TOKEN` fallback.

## What We Tried

Subagent review found and we fixed shell injection risk, tar-slip exposure, missing timeout/size caps, release asset race conditions, and a duplicate workflow trigger. Rejected pretending npm publish should happen locally; first publish needs real registry credentials and release assets, not a developer laptop cosplay.

## Root Cause Analysis

The hard part was not writing a shim. The real mistake would have been treating install-time download as harmless plumbing. It is code execution during package install, so every unchecked filename, archive path, and network fetch is an attack surface.

## Lessons Learned

Package managers are distribution infrastructure, not convenience wrappers. Always threat-model `postinstall`, verify the exact artifact name, reject unsafe archive paths, cap network behavior, and test the publish workflow separately from local pack tests.

## Next Steps

- Release owner: confirm ownership/control of the `@kkauto` npm scope before enabling publication.
- Release owner: configure trusted publishing or `NPM_TOKEN` fallback in GitHub.
- Maintainer: set `NPM_PUBLISH_ENABLED=true` only after credentials and scope are confirmed.
- Maintainer: run the first live npm publish from CI; it was not run locally.

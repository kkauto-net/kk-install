---
title: "Secret-safe license input for unattended kk init"
description: "Add non-argv license input for unattended kk init so VPS provisioning can avoid process-argument secret exposure."
status: completed
progress: 100
priority: P1
effort: 5h
issue: 4
branch: main
tags: [feature, cli, init, automation, provisioning, security]
blockedBy: []
blocks: []
created: 2026-05-21
---

# Secret-safe license input for unattended kk init

## Overview

Implement GitHub issue #4: add secret-safe license input for unattended `kk init` so backend VPS provisioning does not pass KKAuto license keys through process arguments. The minimum production contract is `kk init --yes --license-file /root/.kk/license.tmp --domain customer-subdomain.kkauto.net --language en`, with owner-only temp-file guidance and smoke coverage.

## Issue Context

Issue #3 added true unattended init:

```bash
kk init --yes --license LICENSE-ABCDEF0123456789 --domain example.com --language en
```

That is functional, but license values in argv can appear in `ps`, `/proc/$pid/cmdline`, shell history, audit logs, or VPS telemetry. Backend automation remains production-blocked until a non-argv input path exists and is tested.

## Scope Challenge

| Question | Answer |
|---|---|
| What already exists? | Issue #3 completed `--yes`, `initOptions`, deterministic exit codes, license API validation, prompt guards, secret masking, `.env`/backup `0600`, and README unattended docs. |
| Minimum change set | Add a license-source resolver, `--license-file`, targeted tests/smoke, and docs. Keep legacy `--license` for compatibility but discourage it in automation. |
| Complexity | Estimated 4-6 modified files, no new external dependency, no broad `cmd/init.go` rewrite. `cmd/init.go` is 796 lines, so new logic should live mostly in `cmd/init_options.go` or a small helper file. |

Selected scope: HOLD SCOPE. Focus on issue #4 acceptance criteria, security edge cases, and tests. Do not expand into secret-manager integrations or broad config-source redesign.

## Cross-Plan Dependencies

No active blocking plan detected.

| Relationship | Plan | Status | Reason |
|---|---|---|---|
| Prerequisite already complete | [`260521-1327-non-interactive-kk-init`](../260521-1327-non-interactive-kk-init/plan.md) | completed | Issue #4 builds directly on issue #3 unattended flags and exit-code model. Completed, so no `blockedBy` needed. |
| None | [`260105-0930-reviewdog-pr-workflow`](../260105-0930-reviewdog-pr-workflow/plan.md) | pending | CI/reviewdog workflow only; no direct init/license implementation dependency. |

## Current State Findings

| Area | Finding |
|---|---|
| CLI flags | `cmd/init.go` registers `--yes`, `--license`, `--domain`, and `--language`; no non-argv license source exists. |
| Option model | `cmd/init_options.go` stores one `License string` and requires `--license` when `--yes` is set. |
| License validation | `pkg/license.ValidateFormat` enforces `LICENSE-[A-F0-9]{16}`; `LicenseClient.Validate` calls the remote API. |
| Error handling | `cmd/exit_error.go` maps input validation to `2`, license API failure to `3`, Docker to `4`, render/write to `5`. |
| Secret masking | `sanitizeLicenseError` masks raw license values from API error strings. |
| Output permissions | `pkg/templates.RenderAll` chmods generated `.env` to `0600`; `backupExistingConfigs` writes `.env` backup as `0600`. |
| Docs | README unattended example currently uses unsafe argv `--license`. Docs mention issue #3 argv contract. |
| Tests | `cmd/init_test.go` covers current option validation, masking, backup `.env` mode, and exit-code mapping; no license-file or smoke tests exist. |

## Goals

- Add `kk init --yes --license-file <path> --domain <domain> --language <en|vi>`.
- Keep unattended/no-prompt behavior.
- Ensure license values do not appear in process arguments in documented backend automation flow.
- Preserve deterministic exit codes from issue #3.
- Avoid printing full license values in stdout, stderr, logs, test failure messages, or README examples.
- Document temp-file creation, `0600` permissions, and cleanup.
- Add tests/smoke coverage for missing file, unreadable file, invalid license, and successful no-prompt init.

## Non-Goals

- Do not remove legacy `--license`; keep backward compatibility for existing users.
- Do not add a full secret-manager integration.
- Do not implement broad `.env`/env-file parsing for all init options.
- Do not rewrite `cmd/init.go` or move interactive flow unless directly needed.
- Do not change the license API contract.
- Do not add service-disable flags or unrelated unattended options.

## Command Contract

Recommended automation path:

```bash
install -d -m 700 /root/.kk
umask 077
printf '%s\n' "$KKAUTO_LICENSE" > /root/.kk/license.tmp

kk init \
  --yes \
  --license-file /root/.kk/license.tmp \
  --domain customer-subdomain.kkauto.net \
  --language en

rm -f /root/.kk/license.tmp
```

Compatibility path, allowed but discouraged:

```bash
kk init --yes --license LICENSE-ABCDEF0123456789 --domain example.com --language en
```

Optional within scope if implementation remains small:

```bash
printf '%s\n' "$KKAUTO_LICENSE" | kk init --yes --license-stdin --domain example.com --language en
```

## License Source Rules

| Rule | Behavior |
|---|---|
| Exactly one source | In unattended mode, accept exactly one of `--license`, `--license-file`, or optional `--license-stdin`. |
| Missing source | Exit code `2`. |
| Multiple sources | Exit code `2`; no precedence to avoid ambiguous secret handling. |
| File source | Read file, trim surrounding whitespace/newline, reject empty content, validate license format before API call. |
| File permissions | Document `0600`; implementation may warn or reject group/world-readable files only if kept portable and tested. Minimum issue requirement is docs guidance and safe examples. |
| Stdin source | Read only when explicit `--license-stdin`; reject empty stdin; avoid interactive-mode reads. |
| Legacy argv source | Continue working, but help/docs should warn that it can expose secrets in process listings/history. |

## Exit Code Contract

| Code | Meaning | Issue #4 examples |
|---:|---|---|
| 0 | Success | Files rendered and `.env` chmodded to `0600`. |
| 1 | Legacy untyped error | Existing unrelated behavior. |
| 2 | Input/source validation failure | Missing license source, multiple sources, missing file, unreadable file, empty file, invalid file license, bad domain/language. |
| 3 | License API validation failure | Remote API rejects otherwise well-formed license. |
| 4 | Docker validation failure | Docker/Compose preflight failure unless `--force` bypasses. |
| 5 | Render/write failure | Template render, file write, chmod failure. |

## Security Requirements

- Do not log or print full license from argv, file, stdin, API errors, or tests.
- Error messages should name the source, not the value.
- Mask license as `LICENSE-************6789` only if display is unavoidable.
- Do not pass the license to subprocess arguments in implementation or smoke scripts.
- Generated `.env` and `.env` backups must remain `0600`.
- README examples must use fake license values and owner-only temp-file guidance.
- If a license file is temporary, docs must instruct cleanup.

## Architecture

### Current Issue #3 Flow

```text
Cobra flags
  -> collectInitOptions()
  -> validateInitOptions()
  -> opts.License copied to licenseKey
  -> pkg/license.Validate()
  -> Docker checks
  -> templates.Config
  -> pkg/templates.RenderAll()
```

### Target Flow

```text
Cobra flags / stdin / license file
  -> collectInitOptions()
  -> resolveInitLicenseSource(cmd.InOrStdin())
  -> validateInitOptions() on resolved license + other flags
  -> licenseKey copied from resolved opts.License
  -> existing pkg/license.Validate()
  -> existing Docker checks
  -> existing templates.Config
  -> existing pkg/templates.RenderAll()
```

Implementation note: make the resolver testable with `io.Reader`; `runInit` should use `cmd.InOrStdin()` only when `--license-stdin` is explicitly set.

## Affected Files

| File | Action | Purpose |
|---|---|---|
| `/home/kkdev/kkinstall/cmd/init.go` | Modify | Register `--license-file`; optionally `--license-stdin`; call resolver before validation; keep prompt-free unattended flow. |
| `/home/kkdev/kkinstall/cmd/init_options.go` | Modify | Extend `initOptions`, collect new flags, resolve license sources, validate exactly-one-source and resolved format. |
| `/home/kkdev/kkinstall/cmd/init_test.go` | Modify | Add source-resolution, no-leak, missing/unreadable/invalid/success tests. |
| `/home/kkdev/kkinstall/pkg/templates/embed_test.go` | Modify optional | Add explicit generated `.env` permission assertion if not already covered during smoke. |
| `/home/kkdev/kkinstall/README.md` | Modify | Replace automation example with `--license-file`, add temp-file permission/cleanup and discourage `--license` for automation. |
| `/home/kkdev/kkinstall/docs/code-standards.md` | Modify | Update security standard: automation must use non-argv license source. |
| `/home/kkdev/kkinstall/docs/system-architecture.md` | Modify | Update unattended flow and command examples. |
| `/home/kkdev/kkinstall/docs/codebase-summary.md` | Modify | Update command surface summary after implementation. |

## Phases

| Phase | Name | Status | Effort |
|---|---|---|---|
| 1 | [License Source Resolver](./phase-01-license-source-resolver.md) | Completed | 2h |
| 2 | [Tests And Smoke Coverage](./phase-02-tests-smoke-coverage.md) | Completed | 2h |
| 3 | [Docs And Handoff](./phase-03-docs-handoff.md) | Completed | 1h |

## Dependency Graph

```text
Phase 1 -> Phase 2 -> Phase 3
```

## Validation Strategy

Run after implementation:

```bash
go test ./cmd
go test ./pkg/license ./pkg/templates
go test ./...
```

Final verification completed 2026-05-22:

| Check | Command / Evidence | Result |
|---|---|---|
| Command package tests | `go test ./cmd` | Passed |
| Full Go suite | `go test ./...` | Passed |
| Tester review | tester passed | Passed |
| Debugger review | debugger passed | Passed |
| Code reviewer review | code-reviewer passed | Passed |

Manual smoke with local API skip:

```bash
tmpdir="$(mktemp -d)"
license_file="$tmpdir/license.tmp"
chmod 700 "$tmpdir"
printf 'LICENSE-ABCDEF0123456789\n' > "$license_file"
chmod 600 "$license_file"

workdir="$(mktemp -d)"
(
  cd "$workdir"
  KK_TEST_SKIP_LICENSE_VALIDATION=true kk init --yes --license-file "$license_file" --domain example.com --language en
  test -f docker-compose.yml
  test -f .env
  test -f kkphp.conf
  stat -c '%a' .env
)

rm -f "$license_file"
```

Manual exit-code checks:

```bash
kk init --yes --license-file /missing/license.tmp --domain example.com --language en
echo $? # expect 2

printf 'bad\n' > /tmp/kk-license-bad.tmp
chmod 600 /tmp/kk-license-bad.tmp
kk init --yes --license-file /tmp/kk-license-bad.tmp --domain example.com --language en
echo $? # expect 2
rm -f /tmp/kk-license-bad.tmp
```

## Success Criteria

- [x] `kk init --yes --license-file <path> --domain <domain> --language <lang>` completes without prompts.
- [x] Missing license file exits with code `2`.
- [x] Unreadable license file exits with code `2` or test skips only when root/OS permissions make unreadable simulation invalid.
- [x] Invalid license content exits with code `2` before remote API call.
- [x] Successful file-source init renders expected files and `.env` remains `0600`.
- [x] Legacy `--license` still works.
- [x] Multiple license sources fail with code `2`.
- [x] License API failures still return code `3`.
- [x] No full license values appear in command output, errors, docs examples, or test failure messages.
- [x] README documents owner-only temp files and cleanup.

## Completion Sync-Back

| Area | Evidence | Status |
|---|---|---|
| Phase 1 resolver | `--license-file`, `--license-stdin`, exactly-one-source resolver, source read/trim/empty/large handling, typed exit code `2` | Completed |
| Phase 2 tests | Unit tests cover missing/conflict/file/stdin/invalid/no-leak/large/unreadable-safe paths; `go test ./cmd`; `go test ./...` | Completed |
| Phase 3 docs | README and docs updated to recommend `--license-file`, owner-only temp file, cleanup, legacy argv warning | Completed |

## Scope Changes

| Change | Reason | Impact |
|---|---|---|
| Added optional `--license-stdin` | Stayed small and testable within planned optional scope | Positive; extra non-argv source, covered by tests |
| Added max license source size guard (`4096` bytes) | Prevent accidental large file/stdin reads | Low; tests added |
| Updated `docs/project-overview-pdr.md` | Optional PDR status sync after implementation | Low; docs consistency |

## Risks

| Risk | Impact | Mitigation |
|---|---|---|
| Secret leaks through error strings | High | Closed: source-specific errors without values; masking/no-leak tests for file/stdin paths. |
| `--license-stdin` blocks unexpectedly | Medium | Closed: only reads when explicit `--license-stdin`; TTY check added. |
| Permission tests fail under root | Medium | Closed: unreadable-file test skips when effective user can still read. |
| Breaking existing backend scripts | High | Closed: legacy `--license` preserved. |
| `cmd/init.go` grows harder to maintain | Medium | Closed: resolver lives in `cmd/init_options.go`; `runInit` change small. |
| Over-hardening file mode blocks valid Docker/Kubernetes secret mounts | Medium | Closed: docs guide `0600`; implementation does not reject readable secret mounts by mode. |

## Unresolved Questions

None. Plan completed and progress set to 100%.

# Phase 03: Docs And Handoff

## Context Links

- [Plan Overview](./plan.md)
- [Phase 01](./phase-01-license-source-resolver.md)
- [Phase 02](./phase-02-tests-smoke-coverage.md)
- `/home/kkdev/kkinstall/README.md`
- `/home/kkdev/kkinstall/docs/code-standards.md`
- `/home/kkdev/kkinstall/docs/system-architecture.md`
- `/home/kkdev/kkinstall/docs/codebase-summary.md`

## Overview

Priority: P1. Status: Completed. Updated public and internal docs so backend provisioning uses a non-argv license path by default, and handed off exact verification commands.

## Requirements

### Functional

- README unattended VPS example must use `--license-file` instead of `--license`.
- README must instruct owner-only temp-file creation and cleanup.
- README must keep deterministic exit-code table from issue #3.
- Docs must state `--license` remains available but is discouraged for automation.
- Architecture/standards docs must reflect the new resolver and source rules.

### Non-Functional

- Documentation examples use fake license values only.
- Avoid suggesting env vars directly in command argv.
- Keep docs concise and evidence-based.

## Related Files

| File | Action |
|---|---|
| `/home/kkdev/kkinstall/README.md` | Replace unattended example, add temp-file security notes. |
| `/home/kkdev/kkinstall/docs/code-standards.md` | Add non-argv license-source standard for automation. |
| `/home/kkdev/kkinstall/docs/system-architecture.md` | Update unattended init flow and license-source resolver description. |
| `/home/kkdev/kkinstall/docs/codebase-summary.md` | Update issue #4 behavior after implementation. |
| `/home/kkdev/kkinstall/docs/project-overview-pdr.md` | Optional update if PDR requirements need issue #4 status. |

## Documentation Content

Recommended README snippet:

```bash
install -d -m 700 /root/.kk
umask 077
printf '%s\n' "LICENSE-ABCDEF0123456789" > /root/.kk/license.tmp

kk init \
  --yes \
  --license-file /root/.kk/license.tmp \
  --domain example.com \
  --language en

rm -f /root/.kk/license.tmp
kk start
kk status
```

Security note:

```text
Use `--license-file` for automation. Avoid `--license <key>` in provisioning scripts because argv can be visible through process listings, shell history, audit tooling, or telemetry. Keep temp license files owner-only (`0600`) and delete them after init.
```

## Implementation Steps

1. Update README unattended install section.
2. Add a short warning near command flags or notes for legacy `--license`.
3. Update docs/code-standards.md security rules.
4. Update docs/system-architecture.md unattended flow diagram.
5. Update docs/codebase-summary.md command-surface summary.
6. Optionally update docs/project-overview-pdr.md issue #4 requirement/status after implementation.
7. Run docs spell/grep sanity by searching for unsafe `kk init --yes --license ` examples.

## Todo List

- [x] README uses `--license-file` in unattended example.
- [x] README includes `0600` and cleanup guidance.
- [x] README warns against argv license in automation.
- [x] Code standards updated.
- [x] Architecture docs updated.
- [x] Codebase summary updated.
- [x] Search confirms no recommended automation example uses raw `--license <key>`.

## Completion Evidence

| File | Update | Status |
|---|---|---|
| `README.md` | unattended VPS example uses owner-only temp file and `--license-file`; cleanup guidance; argv warning | Done |
| `docs/code-standards.md` | non-argv license-source standard; no raw license errors | Done |
| `docs/system-architecture.md` | unattended flow includes `resolveInitLicenseSource`; source rules documented | Done |
| `docs/codebase-summary.md` | command surface updated for `--license-file`, `--license-stdin`, legacy `--license` | Done |
| `docs/project-overview-pdr.md` | PDR status synced to implemented | Done |

## Final Handoff Verification

```bash
go test ./cmd
go test ./...
```

Result: both passed. Tester/debugger/code-reviewer passed.

## Success Criteria

- Backend-facing docs no longer instruct passing the license key through argv.
- Legacy compatibility is documented without making it the recommended path.
- Final handoff includes test commands and smoke commands.

## Risk Assessment

| Risk | Mitigation |
|---|---|
| Docs still contain unsafe example | Grep for `kk init --yes --license ` and update all recommendation contexts. |
| Docs imply license file is permanently safe | Explicitly require owner-only permissions and cleanup. |
| Backend needs exact command contract | Keep the `kk init --yes --license-file ... --domain ... --language ...` example copy-pasteable. |

## Security Considerations

- `printf '%s\n' "$KKAUTO_LICENSE" > file` is acceptable if the env var is managed by backend and not expanded into argv of `kk`; docs should still prefer secure secret storage upstream.
- Do not add real-looking customer domains or license keys beyond fake examples.

## Next Steps

Completed. Main agent should keep plan complete; no unfinished phase tasks remain.

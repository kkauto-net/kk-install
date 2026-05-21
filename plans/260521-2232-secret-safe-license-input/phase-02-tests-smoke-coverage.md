# Phase 02: Tests And Smoke Coverage

## Context Links

- [Plan Overview](./plan.md)
- [Phase 01](./phase-01-license-source-resolver.md)
- `/home/kkdev/kkinstall/cmd/init_test.go`
- `/home/kkdev/kkinstall/pkg/templates/embed_test.go`
- `/home/kkdev/kkinstall/pkg/license/license_test.go`

## Overview

Priority: P1. Status: Completed. Added focused tests for issue #4 acceptance criteria and documented smoke path proving no-prompt init works with `--license-file`.

## Requirements

### Functional

- Cover missing license file.
- Cover unreadable license file, with OS/root-safe handling.
- Cover invalid license content from file.
- Cover successful file-source resolution.
- Cover multiple license sources.
- Cover full-license non-exposure in errors.
- Cover successful no-prompt init with `KK_TEST_SKIP_LICENSE_VALIDATION=true` where feasible.

### Non-Functional

- Tests should avoid live API calls.
- Tests should not depend on terminal prompts.
- Tests should assert exit-code classification using `ExitCode(err)`.
- Keep tests deterministic on Linux CI.

## Related Code Files

| File | Action |
|---|---|
| `/home/kkdev/kkinstall/cmd/init_test.go` | Add unit tests for source resolution and option validation. |
| `/home/kkdev/kkinstall/pkg/templates/embed_test.go` | Add explicit `.env` `0600` permission test if not already asserted by command-level smoke. |
| `/home/kkdev/kkinstall/pkg/license/license_test.go` | Optional if adding license normalization helper. |

## Test Cases

| Test | Expected |
|---|---|
| `--yes` without any license source | Exit code `2`. |
| `--license-file` missing path | Exit code `2`, no raw license. |
| `--license-file` unreadable file | Exit code `2`; skip/conditional when running as root makes chmod unreadable ineffective. |
| `--license-file` empty file | Exit code `2`. |
| `--license-file` with `bad` | Exit code `2`; does not call API. |
| `--license-file` with valid key and newline | Validation passes after trim. |
| `--license` plus `--license-file` | Exit code `2`. |
| Optional `--license-stdin` valid key | Validation passes after trim. |
| Optional `--license-stdin` empty stdin | Exit code `2`. |
| Error string checks | No full `LICENSE-ABCDEF0123456789` in source/validation errors. |

## Smoke Scenario

Use local validation skip to avoid live API:

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
  test "$(stat -c '%a' .env)" = "600"
)

rm -f "$license_file"
```

If using `go run .` in local development, run from the repository root and set `workdir` through a shell subshell carefully.

## Implementation Steps

1. Add tests around helper functions first; avoid invoking full Cobra command unless needed.
2. Use `t.TempDir()` and `os.WriteFile(..., 0600)` for license fixtures.
3. For unreadable-file test, `chmod 0000` and restore permissions in cleanup; skip if file can still be read due to root/effective privileges.
4. Assert `ExitCode(err) == exitCodeInputValidation` for all source validation failures.
5. Assert no raw license key appears in returned errors.
6. Add `.env` permission assertion to template tests if missing.
7. Run full test suite.

## Todo List

- [x] Add unit tests for source absence and conflicts.
- [x] Add unit tests for missing/unreadable/empty/invalid file.
- [x] Add unit test for valid file read with trailing newline.
- [x] Add no-leak assertions for source errors.
- [x] Add optional stdin tests if Phase 01 implements `--license-stdin`.
- [x] Add `.env` permission test if not already covered.
- [x] Run `go test ./cmd` and `go test ./...`.

## Verification Results

| Check | Command / Evidence | Result |
|---|---|---|
| Command package tests | `go test ./cmd` | Passed |
| Full Go suite | `go test ./...` | Passed |
| Tester review | tester passed | Passed |
| Debugger review | debugger passed | Passed |
| Code reviewer review | code-reviewer passed | Passed |

Notes:
- Unreadable-file test uses skip path when current user can still read chmod `0000` files.
- `.env` `0600` behavior already covered by command backup test and template render suite.

## Success Criteria

- Issue #4 required error scenarios are covered by automated tests.
- Successful no-prompt file-source path is covered by a test or documented smoke command.
- Full suite passes.
- Tests avoid network dependency.

## Risk Assessment

| Risk | Mitigation |
|---|---|
| Full command smoke is brittle because Docker preflight runs | Use `--force` only if needed to bypass Docker; ensure this does not hide prompt behavior. Prefer helper tests plus manual smoke on disposable VPS. |
| Unreadable test fails under root | Detect and skip with explanation, or test a directory-as-file read error separately. |
| Tests leak fixture license in failure output | Use fake issue-standard license and assert no raw fixture in errors. |

## Security Considerations

- Test fixtures must use fake license values only.
- Do not commit generated `.env` or temp license files.
- Ensure smoke commands remove temp license file.

## Next Steps

Completed. Phase 03 also completed.

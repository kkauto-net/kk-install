# Phase 01: License Source Resolver

## Context Links

- [Plan Overview](./plan.md)
- [Issue #3 Plan](../260521-1327-non-interactive-kk-init/plan.md)
- `/home/kkdev/kkinstall/cmd/init.go`
- `/home/kkdev/kkinstall/cmd/init_options.go`
- `/home/kkdev/kkinstall/cmd/exit_error.go`
- `/home/kkdev/kkinstall/pkg/license/license.go`

## Overview

Priority: P1. Status: Completed. Added a small, testable license-source abstraction for unattended init. Kept existing issue #3 behavior, and allowed non-argv license input through `--license-file` and explicit `--license-stdin`.

## Requirements

### Functional

- Add `--license-file <path>` to `kk init`.
- Optionally add `--license-stdin` if it stays simple and fully tested.
- Keep `--license <key>` backward-compatible.
- In unattended mode, require exactly one license source.
- Read license from file/stdin, trim whitespace/newline, reject empty content.
- Validate license format before remote API call.
- Return typed input validation errors with exit code `2` for source problems.

### Non-Functional

- Do not print full license values.
- Avoid adding dependencies.
- Keep code local and small; do not rewrite `runInit`.
- Prefer helpers in `cmd/init_options.go` or a focused `cmd/init_license_source.go` if `init_options.go` becomes unclear.

## Architecture

```text
collectInitOptions()
  -> initOptions{License, LicenseFile, LicenseStdin, Domain, Language}
resolveInitLicenseSource(opts, stdin)
  -> returns updated opts with License set to resolved value
validateInitOptions(resolvedOpts)
  -> checks domain/language/license format
runInit()
  -> uses opts.License exactly as before
```

Suggested helper shape:

```go
type initOptions struct {
    NonInteractive bool
    Force          bool
    License        string
    LicenseFile    string
    LicenseStdin   bool
    Domain         string
    Language       string
}

func resolveInitLicenseSource(opts initOptions, stdin io.Reader) (initOptions, error) {
    // no-op for interactive mode
    // count sources: --license, --license-file, --license-stdin
    // read and trim selected source
    // return NewExitError(exitCodeInputValidation, err) for bad source
}
```

## Related Code Files

| File | Action |
|---|---|
| `/home/kkdev/kkinstall/cmd/init.go` | Add flag vars and flag registration; call resolver before validation. |
| `/home/kkdev/kkinstall/cmd/init_options.go` | Extend options and implement source resolution/validation. |
| `/home/kkdev/kkinstall/cmd/exit_error.go` | No change expected; reuse `exitCodeInputValidation`. |
| `/home/kkdev/kkinstall/pkg/license/license.go` | No change expected unless adding a tiny normalize helper is cleaner. |

## Implementation Steps

1. Add global flag vars in `cmd/init.go`: `initLicenseFile string` and optional `initLicenseStdin bool`.
2. Register flags in `init()`:
   - `--license-file`, help: "Read license key from file for unattended init".
   - Optional `--license-stdin`, help: "Read license key from stdin for unattended init".
3. Update `initOptions` fields and `collectInitOptions()` trim behavior.
4. Add source-count validation for unattended mode.
5. Add `readLicenseFile(path string) (string, error)` with `os.ReadFile`, `strings.TrimSpace`, empty-content rejection, and source-only errors.
6. Add optional stdin reader using the provided `io.Reader`; do not use global `os.Stdin` in helper tests.
7. Change `runInit` ordering:
   - collect options.
   - resolve license source using `cmd.InOrStdin()`.
   - validate resolved options.
8. Keep the rest of `runInit` using `opts.License` to minimize risk.
9. Update `--license` help to discourage automation use without breaking compatibility.

## Todo List

- [x] Add `--license-file` flag and option field.
- [x] Add optional `--license-stdin` only if implementation remains small.
- [x] Implement exactly-one-source validation.
- [x] Implement file read/trim/empty handling.
- [x] Ensure invalid source errors use exit code `2`.
- [x] Ensure errors do not include full license values.
- [x] Keep legacy `--license` behavior working.

## Completion Evidence

| Item | Evidence | Status |
|---|---|---|
| Flags registered | `cmd/init.go` adds `--license-file`, `--license-stdin`, legacy `--license` help warning | Done |
| Resolver implemented | `resolveInitLicenseSource` in `cmd/init_options.go` | Done |
| Source reads safe | file/stdin trim, empty reject, non-regular reject, max 4096 bytes | Done |
| Exit code contract | source errors wrapped with `exitCodeInputValidation` (`2`) | Done |
| Secret non-exposure | errors name source, not value; no-leak tests added | Done |

## Success Criteria

- `validateInitOptions` no longer requires argv `--license`; it requires a resolved license from exactly one source.
- File-source license reaches existing API validation path.
- Missing/multiple/empty/invalid source errors produce `ExitCode(err) == 2`.
- No prompt path is introduced before source validation.

## Risk Assessment

| Risk | Mitigation |
|---|---|
| Resolver reads stdin in interactive mode | Return early unless `opts.NonInteractive && opts.LicenseStdin`. |
| Multiple sources silently choose unsafe argv | Reject all multiple-source cases. |
| License appears in error message | Never include read content in errors; use source names only. |
| Large file read | License files are tiny; optional guard can reject content over a small limit, but avoid overengineering unless tests need it. |

## Security Considerations

- `--license-file` hides license content from argv; path is not secret.
- Do not enforce strict file mode unless portable behavior is designed. The plan requires documentation and safe examples at minimum.
- Avoid printing raw `opts.License` in new errors.

## Next Steps

Completed. Phase 02 also completed.

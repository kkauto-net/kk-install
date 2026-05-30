# AGENTS.md

## Project Shape

- Go CLI module `github.com/kkauto-net/kk-install`; installed binary is `kk`.
- Runtime entrypoint is `main.go -> cmd.Execute()`. Keep `main.go` thin; command wiring lives in `cmd/`, reusable behavior in `pkg/*`.
- Read `docs/code-standards.md` and `docs/system-architecture.md` before changing command behavior, templates, validation, release flow, or docs.

## Commands

- Setup deps: `make deps` (`go mod download` then `go mod tidy`).
- Format: `make fmt` (`go fmt ./...`).
- Lint: `make lint` (`golangci-lint run`; config enables tests and has a 5m timeout).
- Full tests: `make test` or `go test -v ./...`.
- Docker-free binary smoke: `make test-smoke`.
- Focused package test: `go test -v ./cmd -run TestName`, `go test -v ./pkg/templates -run TestName`, etc.
- Build current platform: `make build` (`CGO_ENABLED=0 go build ... -o build/kk .`).
- Linux release-style local builds only: `make build-all`.

## Package Boundaries

- `cmd/`: Cobra commands, flags, prompt flow, command orchestration, typed exit-code mapping.
- `pkg/license/`: license format and kk license API validation; do not duplicate this in commands.
- `pkg/templates/`: kkengine embedded templates and generated `.env` permissions.
- `pkg/validator/`: Docker, Compose, port, env, config, disk, and preflight checks.
- `pkg/compose/`: Docker Compose command execution and Compose YAML parsing.
- `pkg/ui/`: i18n, terminal output, help templates, progress, tables, password generation.
- `pkg/n8n/`: n8n stack paths, config validation, and templates.
- `pkg/selfupdate/`: GitHub release lookup/download and binary replacement.

## `kk init --yes` Contract

- Unattended init requires `--yes`, exactly one license source, `--domain`, and `--language en|vi`.
- Prefer `--license-file` or `--license-stdin` for automation; `--license` is compatibility-only because argv can leak.
- License format is `LICENSE-[A-F0-9]{16}`; file/stdin license input is capped at 4096 bytes.
- Stable unattended exit codes: `2` input, `3` license, `4` Docker preflight, `5` template/write; untyped legacy errors exit `1`.
- Keep unattended paths prompt-free. Interactive prompts may use `huh` only outside automation paths.

## Templates And Generated Files

- Template changes usually need `go test -v ./pkg/templates`.
- Golden generator must run from its own directory: `cd pkg/templates/testdata && go run generate_golden.go`; otherwise it writes `golden/` in the wrong place.
- Generated kkengine files are `docker-compose.yml`, `.env`, `kkphp.conf`, optional `Caddyfile`, optional `kkfiler.toml`; generated `.env` must stay `0600`.
- Do not read, print, or commit generated `.env` values or license/private secrets. Treat `example/.env` as sensitive unless a maintainer says otherwise.

## CI And Release Gotchas

- CI uses Go from `go.mod`, runs `go test -v ./...`, builds, runs `make test-smoke`, lints on push/PR, and runs race/shuffle outside PRs.
- `release.yml` and `draft-release.yml` run `go test -v ./...` before release steps.
- `validate-templates.yml` uses `go-version-file: go.mod`.
- Real Docker Compose e2e lives in `.github/workflows/e2e-compose.yml` and is nightly/manual only; it requires `KKAUTO_E2E_LICENSE`.
- GoReleaser publishes Linux `amd64`/`arm64` tarballs only; do not promise macOS artifacts without changing `.goreleaser.yml` and installer behavior.
- `scripts/install.sh` and `pkg/selfupdate` require `checksums.txt` SHA256 verification before installing or replacing the `kk` binary; no release signature verification is implemented yet.

## Docs

- Public command changes belong in `README.md`; architecture/standards/deployment/roadmap details belong in `docs/`.
- Keep `README.md` under 300 lines and evergreen docs under the project doc line target.
- Documentation and reports are written in English; user-facing chat in this workspace is Vietnamese unless explicitly requested otherwise.

# kkcli

[![Go Report Card](https://goreportcard.com/badge/github.com/kkauto-net/kk-install)](https://goreportcard.com/report/github.com/kkauto-net/kk-install)
[![Go Reference](https://pkg.go.dev/badge/github.com/kkauto-net/kk-install.svg)](https://pkg.go.dev/github.com/kkauto-net/kk-install)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![CI](https://github.com/kkauto-net/kk-install/actions/workflows/ci.yml/badge.svg)](https://github.com/kkauto-net/kk-install/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/kkauto-net/kk-install)](https://github.com/kkauto-net/kk-install/releases)
![Platform](https://img.shields.io/badge/release%20artifacts-linux-blue)

`kkcli` is a Go CLI for installing and operating kkengine Docker Compose stacks. The installed binary is `kk`.

## Quick Install

```bash
curl -sSL https://raw.githubusercontent.com/kkauto-net/kk-install/main/scripts/install.sh | bash
```

Current GoReleaser artifacts are Linux `amd64` and `arm64`. Other platforms may build from source, but are not published by the current release config.

Verify installation:

```bash
kk --version
```

## Features

- 🐳 **Docker Compose Management** - Initialize, start, restart, and monitor your stack
- ⚡ **Health Monitoring** - Real-time container health checks
- 🔄 **Auto Update** - Keep images up-to-date with one command
- 🌐 **Multi-language** - English and Vietnamese support
- 🔒 **Secure by Default** - Auto-generates strong passwords
- 🤖 **n8n Support** - Optional n8n stack install and lifecycle commands

## Quick Start

```bash
# Initialize your stack
kk init

# Start all services
kk start

# Check status
kk status
```

## Unattended VPS Install

Use this mode for backend provisioning scripts where no interactive prompts are available.

```bash
curl -sSL https://raw.githubusercontent.com/kkauto-net/kk-install/main/scripts/install.sh | bash

install -d -m 700 /root/.kk
license_file="$(mktemp /root/.kk/license.XXXXXX)"
cleanup_license_file() { rm -f "$license_file"; }
trap cleanup_license_file EXIT

printf '%s\n' "$KKAUTO_LICENSE" > "$license_file"
chmod 600 "$license_file"

kk init \
  --yes \
  --license-file "$license_file" \
  --domain example.com \
  --language en

kk start
kk status
```

Notes:
- Supported languages: `en`, `vi`.
- The license key is validated with the kk license API.
- Use `--license-file` for automation. Avoid `--license <key>` in provisioning scripts because argv can be visible through process listings, shell history, audit tooling, or telemetry.
- Keep temporary license files owner-only (`0600`) and clean them up with a trap so failures do not leave secrets behind.
- Generated `.env` is written with owner-only permissions.
- Existing config files are overwritten after a timestamped backup when `--yes` is used.
- Do not commit generated `.env` or share license/private secrets.
- Generated Compose mounts `/etc/machine-id` read-only for license hardware identity. It is a stable identifier input, not a secret; backend heartbeat and offline-token expiry enforce runtime access.

Unattended mode exits with deterministic codes:

| Code | Meaning |
|------|---------|
| `0` | Success |
| `2` | Flag or input validation failed |
| `3` | License validation failed |
| `4` | Docker preflight failed |
| `5` | Template render or file write failed |

## Commands

| Command | Description |
|---------|-------------|
| `kk init` | Initialize Docker Compose stack with interactive prompts or unattended flags |
| `kk start` | Run preflight checks and start all services |
| `kk stop` | Stop all running services |
| `kk remove` | Remove all containers, networks (use `-v` to also remove volumes) |
| `kk restart` | Restart all running services |
| `kk status` | Display status of all containers |
| `kk update -f` | Pull images, show changed image identities, and recreate containers; `-f` skips confirmation |
| `kk selfupdate --check` | Check or install latest CLI release; use `-f` to skip confirmation |
| `kk config show` | Show language, project directory, and config path |
| `kk completion bash\|zsh\|fish` | Generate shell completion script |

### n8n Commands

| Command | Description |
|---------|-------------|
| `kk n8n install -f` | Render n8n Compose files with defaults when `-f` is set |
| `kk n8n start` | Start the n8n stack |
| `kk n8n stop` | Stop the n8n stack |
| `kk n8n restart` | Restart the n8n stack |
| `kk n8n status` | Show n8n container status |
| `kk n8n logs -f -n 100 -a` | Tail logs; `-a` includes all n8n containers |
| `kk n8n update -f` | Pull and recreate n8n containers; `-f` skips confirmation |
| `kk n8n remove -v` | Remove n8n containers; `-v` also removes volumes after confirmation |

## Supported Components

| Component | Description |
|-----------|-------------|
| **kkengine** | Core service container |
| **MariaDB** | Primary database |
| **Redis** | Cache and session management |
| **SeaweedFS** | Distributed file storage (optional) |
| **Caddy** | Web server and reverse proxy (optional) |

## Requirements

- **Docker** - Installed and running
- **Docker Compose** - Version 2.0+

## Contributing

Contributions welcome! See [Code Standards](./docs/code-standards.md) and [System Architecture](./docs/system-architecture.md).

```bash
make deps
make fmt
make lint
make test
make test-smoke
make build
```

## License

MIT License - see [LICENSE](LICENSE) for details.

## Documentation

- [Project Overview](./docs/project-overview-pdr.md)
- [Codebase Summary](./docs/codebase-summary.md)
- [Code Standards](./docs/code-standards.md)
- [System Architecture](./docs/system-architecture.md)
- [Deployment Guide](./docs/deployment-guide.md)
- [Design Guidelines](./docs/design-guidelines.md)
- [Project Roadmap](./docs/project-roadmap.md)

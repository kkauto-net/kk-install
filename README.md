# kkcli

[![Go Report Card](https://goreportcard.com/badge/github.com/kkauto-net/kk-install)](https://goreportcard.com/report/github.com/kkauto-net/kk-install)
[![Go Reference](https://pkg.go.dev/badge/github.com/kkauto-net/kk-install.svg)](https://pkg.go.dev/github.com/kkauto-net/kk-install)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![CI](https://github.com/kkauto-net/kk-install/actions/workflows/ci.yml/badge.svg)](https://github.com/kkauto-net/kk-install/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/kkauto-net/kk-install)](https://github.com/kkauto-net/kk-install/releases)
![Platform](https://img.shields.io/badge/platform-linux%20%7C%20macos-blue)

A CLI tool for managing kkengine Docker Compose stacks with ease.

## Quick Install

```bash
curl -sSL https://raw.githubusercontent.com/kkauto-net/kk-install/main/scripts/install.sh | bash
```

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

kk init \
  --yes \
  --license LICENSE-ABCDEF0123456789 \
  --domain example.com \
  --language en

kk start
kk status
```

Notes:
- Supported languages: `en`, `vi`.
- The license key is validated with the kk license API.
- Generated `.env` is written with owner-only permissions.
- Existing config files are overwritten after a timestamped backup when `--yes` is used.
- Do not commit generated `.env` or share license/private secrets.

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
| `kk update` | Update to latest version and pull new images |
| `kk selfupdate` | Update kk CLI to the latest version |
| `kk completion` | Generate shell completion script |

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

## License

MIT License - see [LICENSE](LICENSE) for details.

## Documentation

- [Project Overview](./docs/project-overview-pdr.md)
- [Codebase Summary](./docs/codebase-summary.md)
- [Code Standards](./docs/code-standards.md)
- [System Architecture](./docs/system-architecture.md)

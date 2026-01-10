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

## Commands

| Command | Description |
|---------|-------------|
| `kk init` | Initialize Docker Compose stack with interactive prompts |
| `kk start` | Run preflight checks and start all services |
| `kk restart` | Restart all running services |
| `kk status` | Display status of all containers |
| `kk update` | Update to latest version and pull new images |
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

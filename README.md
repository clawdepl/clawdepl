# clawdpl

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
[![npm](https://img.shields.io/npm/v/clawdpl?logo=npm)](https://www.npmjs.com/package/clawdpl)
[![PyPI](https://img.shields.io/pypi/v/clawdpl?logo=python)](https://pypi.org/project/clawdpl/)
[![CI](https://github.com/moltyverse/clawdpl/actions/workflows/ci.yml/badge.svg)](https://github.com/moltyverse/clawdpl/actions/workflows/ci.yml)

Create and manage [OpenClaw](https://openclaw.io) AI Agent orchestrator instances on hosted infrastructure—instantly.

## What is OpenClaw?

OpenClaw is an AI Agent orchestrator that enables you to deploy, configure, and scale intelligent agent workflows. With `clawdpl`, you can spin up new OpenClaw instances in seconds, manage deployments, and monitor your agent infrastructure from the command line.

## Quick Start

The fastest way to use clawdpl is with `npx` or `pipx`—no installation required:

```bash
# Using npx (Node.js)
npx clawdpl init my-project

# Using pipx (Python)
pipx run clawdpl init my-project
```

## Installation

### npx (No Install Required)

```bash
npx clawdpl <command>
```

### pipx (No Install Required)

```bash
pipx run clawdpl <command>
```

### npm (Global Install)

```bash
npm install -g clawdpl
```

### pip (Global Install)

```bash
pip install clawdpl
```

### Go

```bash
go install github.com/moltyverse/clawdpl@latest
```

### From Source

```bash
git clone https://github.com/moltyverse/clawdpl.git
cd clawdpl
make build
```

## Usage

```bash
# Create a new OpenClaw project
clawdpl init my-agent-project

# Navigate to your project
cd my-agent-project

# Deploy to hosted infrastructure
clawdpl deploy

# Check the status of your instance
clawdpl status
```

## Commands

| Command | Description |
|---------|-------------|
| `login` | Authenticate with clawdpl.dev |
| `logout` | Clear stored credentials |
| `new [name]` | Create a new OpenClaw instance |
| `list` | List all deployed instances |
| `status <name>` | Check the status of an instance |
| `start <name>` | Start an instance |
| `stop <name>` | Stop an instance |
| `delete <name>` | Delete an instance |
| `version` | Print version and build information |

Run `clawdpl --help` for a complete list of commands and options.

### Version Command

The `version` command displays detailed build information:

```bash
# Human-readable output
clawdpl version

# JSON output for scripting
clawdpl version --json
```

Example output:

```
clawdpl v1.2.3
  commit:   abc1234 (abc1234567890...)
  built:    2026-02-08T12:34:56Z
  mode:     prod
  platform: linux/amd64
  go:       go1.21.0
```

## Configuration

Configuration can be provided via:

1. **Environment variables** (prefixed with `OPENCLAW_`)
2. **Configuration file** (`~/.openclaw/config.yaml`)
3. **Command-line flags**

Example configuration file:

```yaml
# ~/.openclaw/config.yaml
api_key: your-api-key
default_region: us-east-1
```

## Architecture

```
clawdpl/
├── cmd/                    # CLI commands (Cobra)
├── internal/
│   ├── api/               # OpenClaw API client
│   ├── auth/              # Authentication logic
│   ├── buildinfo/         # Build metadata (version, commit, etc.)
│   ├── config/            # Configuration and credentials
│   └── tui/               # Terminal UI components
├── npm/                   # npm package wrapper
├── python/                # PyPI package wrapper
├── scripts/               # Build and test scripts
├── .github/workflows/     # CI/CD pipelines
├── main.go                # Entry point
├── Makefile               # Build automation
└── go.mod                 # Go module definition
```

The CLI is built in Go for performance and cross-platform compatibility. The npm and PyPI packages are thin wrappers that automatically download the appropriate binary for your platform.

## Development

### Prerequisites

- Go 1.21 or later
- Node.js 16+ (for npm package development)
- Python 3.8+ (for PyPI package development)
- Make

### Building

```bash
# Build the binary (production)
make build

# Build with debug flags enabled
make build-debug

# Build for all platforms
make build-all

# Create release archives
make release
```

### Testing

```bash
# Run all tests
make test

# Test Go code
make test-go

# Test npm package locally
make test-npm

# Test Python package locally
make test-python

# Test npx execution
make test-npx

# Test pipx execution
make test-pipx
```

### Local Development

```bash
# Build and copy binary to wrappers for local testing
make dev

# Build debug binary and copy to wrappers
make dev-debug

# Clean development binaries
make dev-clean
```

### Running Locally

```bash
# Run directly with Go
go run . --help

# Or build and run
make build && ./clawdpl --help
```

### Debug Builds

Debug builds include additional flags for development and testing purposes. These flags are **completely absent** in production builds—they don't appear in `--help` and cannot be used.

To create a debug build:

```bash
make build-debug
```

Debug builds can be identified by running `clawdpl version`, which will show `mode: debug` instead of `mode: prod`.

Debug-only flags include:
- `--unsafe-endpoint`: Override the API endpoint URL (for testing against staging/local servers)

## Releasing

Releases are automated via GitHub Actions. To create a new release:

1. Create a Pull Request with your changes
2. Add a label in the format `release:X.Y.Z` (e.g., `release:1.0.0`)
3. The workflow will automatically:
   - Build binaries for all platforms
   - Create a GitHub Release with the binaries
   - Publish to npm
   - Publish to PyPI
   - Update version numbers in the codebase

### Required Secrets

For the release workflow to work, configure these secrets in your repository:

- `NPM_TOKEN`: npm access token with publish permissions
- `PYPI_TOKEN`: PyPI API token with upload permissions

## Contributing

We welcome contributions! Please see our contributing guidelines:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes using [conventional commits](https://www.conventionalcommits.org/)
4. Push to your branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Commit Message Format

We use conventional commits. Examples:

```
feat(cli): add instance creation command
fix(api): handle timeout errors gracefully
docs: update installation instructions
```

## Roadmap

- [ ] Instance creation and management
- [ ] Multi-region deployment support
- [ ] Team collaboration features
- [ ] Custom agent templates
- [ ] Monitoring and alerting integration
- [ ] CI/CD pipeline integration

## Support

- **Documentation**: [docs.openclaw.io](https://docs.openclaw.io)
- **Issues**: [GitHub Issues](https://github.com/moltyverse/clawdpl/issues)
- **Discussions**: [GitHub Discussions](https://github.com/moltyverse/clawdpl/discussions)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

Built with care by [Moltyverse](https://moltyverse.com)

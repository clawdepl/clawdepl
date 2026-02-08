# clawdpl

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
[![npm](https://img.shields.io/npm/v/clawdpl?logo=npm)](https://www.npmjs.com/package/clawdpl)
[![PyPI](https://img.shields.io/pypi/v/clawdpl?logo=python)](https://pypi.org/project/clawdpl/)

Create and manage [OpenClaw](https://openclaw.io) AI Agent orchestrator instances on hosted infrastructure—instantly.

## What is OpenClaw?

OpenClaw is an AI Agent orchestrator that enables you to deploy, configure, and scale intelligent agent workflows. With `clawdpl`, you can spin up new OpenClaw instances in seconds, manage deployments, and monitor your agent infrastructure from the command line.

## Installation

Choose your preferred installation method:

### npm (Node.js)

```bash
npm install -g clawdpl
```

Or use directly with npx:

```bash
npx clawdpl init my-project
```

### pip (Python)

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
go build -o clawdpl .
```

## Quick Start

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
| `init <name>` | Create a new OpenClaw project |
| `deploy` | Deploy your project to hosted infrastructure |
| `status` | Check the status of your instances |
| `logs` | View logs from your running instances |
| `destroy` | Tear down an instance |
| `config` | Manage configuration settings |

Run `clawdpl --help` for a complete list of commands and options.

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
│   └── api/               # OpenClaw API client
├── npm/                   # npm package wrapper
├── python/                # PyPI package wrapper
├── main.go                # Entry point
└── go.mod                 # Go module definition
```

The CLI is built in Go for performance and cross-platform compatibility. The npm and PyPI packages are thin wrappers that download and execute the appropriate binary for your platform.

## Development

### Prerequisites

- Go 1.21 or later
- Node.js 16+ (for npm package development)
- Python 3.8+ (for PyPI package development)

### Building

```bash
# Build the binary
go build -o clawdpl .

# Run tests
go test ./...

# Build for all platforms
./scripts/build-all.sh
```

### Running Locally

```bash
# Run directly with Go
go run . --help

# Or build and run
go build -o clawdpl . && ./clawdpl --help
```

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

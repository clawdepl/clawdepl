# create-claw-app

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
[![npm](https://img.shields.io/npm/v/create-claw-app?logo=npm)](https://www.npmjs.com/package/create-claw-app)
[![PyPI](https://img.shields.io/pypi/v/create-claw-app?logo=python)](https://pypi.org/project/create-claw-app/)

Create and manage [OpenClaw](https://openclaw.io) AI Agent orchestrator instances on hosted infrastructure—instantly.

## What is OpenClaw?

OpenClaw is an AI Agent orchestrator that enables you to deploy, configure, and scale intelligent agent workflows. With `create-claw-app`, you can spin up new OpenClaw instances in seconds, manage deployments, and monitor your agent infrastructure from the command line.

## Installation

Choose your preferred installation method:

### npm (Node.js)

```bash
npm install -g create-claw-app
```

Or use directly with npx:

```bash
npx create-claw-app init my-project
```

### pip (Python)

```bash
pip install create-claw-app
```

### Go

```bash
go install github.com/moltyverse/create-claw-app@latest
```

### From Source

```bash
git clone https://github.com/moltyverse/create-claw-app.git
cd create-claw-app
go build -o create-claw-app .
```

## Quick Start

```bash
# Create a new OpenClaw project
create-claw-app init my-agent-project

# Navigate to your project
cd my-agent-project

# Deploy to hosted infrastructure
create-claw-app deploy

# Check the status of your instance
create-claw-app status
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

Run `create-claw-app --help` for a complete list of commands and options.

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
create-claw-app/
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
go build -o create-claw-app .

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
go build -o create-claw-app . && ./create-claw-app --help
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
- **Issues**: [GitHub Issues](https://github.com/moltyverse/create-claw-app/issues)
- **Discussions**: [GitHub Discussions](https://github.com/moltyverse/create-claw-app/discussions)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

Built with care by [Moltyverse](https://moltyverse.com)

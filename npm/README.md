# clawdpl

Create and manage OpenClaw AI Agent orchestrator instances from the command line.

This is the npm distribution of clawdpl. For full documentation, see the [main repository](https://github.com/moltyverse/clawdpl).

## Quick Start (No Install)

```bash
npx clawdpl init my-project
```

## Installation

```bash
npm install -g clawdpl
```

## Usage

```bash
# Create a new OpenClaw project
clawdpl init my-project

# Deploy to hosted infrastructure
clawdpl deploy

# Check instance status
clawdpl status

# Get help
clawdpl --help
```

## How It Works

This package is a thin wrapper around the clawdpl Go binary. On installation, it automatically downloads the appropriate binary for your platform from GitHub Releases.

Supported platforms:
- Linux (x64, arm64)
- macOS (x64, arm64)
- Windows (x64, arm64)

## Environment Variables

- `CLAWDPL_BINARY_PATH`: Override the path to the clawdpl binary

## License

MIT - see [LICENSE](https://github.com/moltyverse/clawdpl/blob/main/LICENSE)

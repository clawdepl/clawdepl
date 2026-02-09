# clawdepl

Create and manage OpenClaw AI Agent orchestrator instances from the command line.

This is the npm distribution of clawdepl. For full documentation, see the [main repository](https://github.com/clawdepl/clawdepl).

## Quick Start (No Install)

```bash
npx clawdepl init my-project
```

## Installation

```bash
npm install -g clawdepl
```

## Usage

```bash
# Create a new OpenClaw project
clawdepl init my-project

# Deploy to hosted infrastructure
clawdepl deploy

# Check instance status
clawdepl status

# Get help
clawdepl --help
```

## How It Works

This package is a thin wrapper around the clawdepl Go binary. On installation, it automatically downloads the appropriate binary for your platform from GitHub Releases.

Supported platforms:
- Linux (x64, arm64)
- macOS (x64, arm64)
- Windows (x64, arm64)

## Environment Variables

- `CLAWDPL_BINARY_PATH`: Override the path to the clawdepl binary

## License

MIT - see [LICENSE](https://github.com/clawdepl/clawdepl/blob/main/LICENSE)

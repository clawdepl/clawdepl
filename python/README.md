# clawdpl

Create and manage OpenClaw AI Agent orchestrator instances from the command line.

This is the PyPI distribution of clawdpl. For full documentation, see the [main repository](https://github.com/moltyverse/clawdpl).

## Quick Start (No Install)

```bash
pipx run clawdpl init my-project
```

## Installation

```bash
pip install clawdpl
```

Or with pipx for isolated installation:

```bash
pipx install clawdpl
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

This package is a thin wrapper around the clawdpl Go binary. On first run, it automatically downloads the appropriate binary for your platform from GitHub Releases and caches it locally.

Supported platforms:
- Linux (x64, arm64)
- macOS (x64, arm64)
- Windows (x64, arm64)

## Environment Variables

- `CLAWDPL_BINARY_PATH`: Override the path to the clawdpl binary

## Requirements

- Python 3.8+
- httpx (installed automatically)

## License

MIT - see [LICENSE](https://github.com/moltyverse/clawdpl/blob/main/LICENSE)

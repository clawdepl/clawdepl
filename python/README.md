# clawdepl

Create and manage OpenClaw AI Agent orchestrator instances from the command line.

This is the PyPI distribution of clawdepl. For full documentation, see the [main repository](https://github.com/clawdepl/clawdepl).

## Quick Start (No Install)

```bash
pipx run clawdepl --help
```

## Installation

```bash
pip install clawdepl
```

Or with pipx for isolated installation:

```bash
pipx install clawdepl
```

## Usage

```bash
# Authenticate
clawdepl login

# Create a new instance
clawdepl new my-agent

# List instances
clawdepl list

# Get help
clawdepl --help
```

## How It Works

This package is a thin wrapper around the clawdepl Go binary. On first run, it automatically downloads the appropriate binary for your platform from GitHub Releases and caches it locally.

Supported platforms:
- Linux (x64, arm64)
- macOS (x64, arm64)
- Windows (x64, arm64)

## Environment Variables

- `CLAWDEPL_BINARY_PATH`: Override the path to the clawdepl binary

## Requirements

- Python 3.8+
- httpx (installed automatically)

## License

MIT - see [LICENSE](https://github.com/clawdepl/clawdepl/blob/main/LICENSE)

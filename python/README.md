# clawdepl

Create and manage OpenClaw AI Agent orchestrator instances from the command line.

This is the PyPI distribution of clawdepl. For full documentation, see the [main repository](https://github.com/moltyverse/clawdepl).

## Quick Start (No Install)

```bash
pipx run clawdepl init my-project
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

This package is a thin wrapper around the clawdepl Go binary. On first run, it automatically downloads the appropriate binary for your platform from GitHub Releases and caches it locally.

Supported platforms:
- Linux (x64, arm64)
- macOS (x64, arm64)
- Windows (x64, arm64)

## Environment Variables

- `CLAWDPL_BINARY_PATH`: Override the path to the clawdepl binary

## Requirements

- Python 3.8+
- httpx (installed automatically)

## License

MIT - see [LICENSE](https://github.com/moltyverse/clawdepl/blob/main/LICENSE)

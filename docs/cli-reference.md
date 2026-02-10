# clawdepl CLI Reference

clawdepl is a CLI tool for creating and managing OpenClaw AI Agent orchestrator instances on hosted infrastructure.

## Installation

```bash
# npm
npm install -g clawdepl

# pip
pip install clawdepl

# Or download the binary directly from GitHub releases
```

## Quick Start

```bash
clawdepl                    # Run setup wizard (login + create instance)
clawdepl login              # Authenticate with clawdepl.dev
clawdepl new my-agent       # Create a new instance
clawdepl list               # List all instances
```

## Global Flags

These flags are available on all commands:

| Flag | Description |
|------|-------------|
| `-h, --help` | Help for the command |
| `-v, --version` | Version for clawdepl (root command only) |

## Commands

### Setup Wizard

#### `clawdepl`

When called without arguments, runs the interactive setup wizard that guides you through login and creating your first instance.

```bash
clawdepl
```

---

### Authentication

#### `clawdepl auth`

Parent command for managing authentication with clawdepl.dev.

```bash
clawdepl auth login         # Authenticate
clawdepl auth logout        # Sign out
clawdepl auth status        # Show account info
```

#### `clawdepl auth login`

Authenticate with clawdepl.dev to manage your OpenClaw instances.

| Flag | Description |
|------|-------------|
| `--api-key` | Authenticate with an API key |
| `--info` | Show current user info instead of logging in |
| `--no-browser` | Don't open browser, print URL for manual auth |
| `--token <token>` | Authenticate with a pre-obtained token |

**Examples:**

```bash
clawdepl auth login                    # OAuth via browser (default)
clawdepl auth login --no-browser       # OAuth without browser
clawdepl auth login --token YOUR_TOKEN # Token-based authentication
clawdepl auth login --api-key          # API key authentication
clawdepl auth login --info             # Show current login status
```

#### `clawdepl auth logout`

Log out from clawdepl.dev and clear stored credentials.

Removes the credentials file from `~/.clawdepl/credentials.json`.

```bash
clawdepl auth logout
```

#### `clawdepl auth status`

Display information about the currently logged-in user.

| Flag | Description |
|------|-------------|
| `--json` | Output as JSON (for scripting) |

**Examples:**

```bash
clawdepl auth status          # Human-readable output
clawdepl auth status --json   # JSON output
```

**Output fields:**
- User ID
- Email
- Name
- Authentication method (OAuth or API Key)
- Token expiry (for OAuth)

---

### Authentication Shortcuts

These commands are aliases for the `auth` subcommands for convenience.

#### `clawdepl login`

Alias for `clawdepl auth login`. All flags are identical.

```bash
clawdepl login
clawdepl login --no-browser
clawdepl login --token YOUR_TOKEN
clawdepl login --api-key
clawdepl login --info
```

#### `clawdepl logout`

Alias for `clawdepl auth logout`.

```bash
clawdepl logout
```

#### `clawdepl account`

Alias for `clawdepl auth status`. All flags are identical.

```bash
clawdepl account
clawdepl account --json
```

---

### Instance Management

#### `clawdepl new [name]`

Create a new Molty (AI agent) instance with an interactive wizard.

If a name is provided, the wizard will skip the name prompt. The wizard guides you through:
1. Instance name (if not provided)
2. Claude credential (API key or setup-token)
3. Purpose/description (vibe)

For CI/non-interactive mode, provide:
- `--claude-token <token>`
- `--purpose <text>`
- Optional: `--auth-choice api-key|setup-token`
- Optional: `--model provider/model`

**Examples:**

```bash
clawdepl new              # Interactive wizard
clawdepl new my-agent     # Skip name prompt
clawdepl new my-agent --claude-token "$ANTHROPIC_API_KEY" --purpose "CI instance"
clawdepl new my-agent --claude-token "$CLAUDE_SETUP_TOKEN" --auth-choice setup-token --model anthropic/claude-sonnet-4-5 --purpose "CI instance"
```

#### `clawdepl list`

List all Molty instances associated with your account.

**Aliases:** `ls`

Displays instance name, status, sandbox ID, and creation time in a table format.

```bash
clawdepl list
clawdepl ls
```

#### `clawdepl status <sandbox_id>`

Show detailed status information for a Molty instance.

Displays status, stage, progress, and gateway URL if available.

```bash
clawdepl status sandbox_abc123
```

#### `clawdepl start <sandbox_id>`

Start a stopped Molty instance.

```bash
clawdepl start sandbox_abc123
```

#### `clawdepl stop <sandbox_id>`

Stop a running Molty instance.

```bash
clawdepl stop sandbox_abc123
```

#### `clawdepl delete <sandbox_id> [sandbox_id...]`

Delete one or more Molty instances.

| Flag | Description |
|------|-------------|
| `-y, --yes` | Skip confirmation prompt |

**Examples:**

```bash
clawdepl delete sandbox_abc123           # Delete with confirmation
clawdepl delete sandbox_abc123 -y        # Delete without confirmation
clawdepl delete sandbox_1 sandbox_2      # Delete multiple instances
```

#### `clawdepl ssh <sandbox_name>`

SSH into a running Molty instance by name or sandbox ID.

Automatically provisions temporary SSH credentials (valid for 60 minutes), connects to the sandbox interactively, and revokes the credentials when the session ends. The sandbox must be in a running state to establish SSH access.

**Requirements:**
- OpenSSH client installed on your system
- Instance must be running (use `clawdepl start` if stopped)

**Examples:**

```bash
clawdepl ssh my-agent           # SSH by instance name
clawdepl ssh sandbox_abc123     # SSH by sandbox ID
```

**Behavior:**
- Provisions temporary SSH token (expires in 60 minutes)
- Opens interactive SSH session
- Automatically revokes token on session exit or Ctrl+C
- If revoke fails, token auto-expires after 60 minutes

**Error cases:**
- Not logged in → Prompts to run `clawdepl login`
- Instance not found → Suggests `clawdepl list`
- Instance stopped → Suggests `clawdepl start <name>`
- SSH client missing → Shows installation instructions

---

#### `clawdepl chat <sandbox_name> <message>`

Send a message to a running instance without opening SSH manually.

Creates a temporary session, runs a prompt command in the sandbox, prints output, and closes the session.

| Flag | Description |
|------|-------------|
| `--runner <cmd>` | Base command used in sandbox (default: `claude -p`) |
| `--timeout <sec>` | Execution timeout in seconds (default: `120`) |

**Examples:**

```bash
clawdepl chat wifey "Hi, introduce yourself"
clawdepl chat wifey "What can you do?" --runner "claude -p"
```

---

### Utility Commands

#### `clawdepl version`

Print detailed version and build information.

| Flag | Description |
|------|-------------|
| `--json` | Output version info as JSON |

**Examples:**

```bash
clawdepl version          # Human-readable output
clawdepl version --json   # JSON output for scripting
```

**Output includes:**
- Version number
- Git commit hash
- Build date
- Build mode (debug/prod)
- Platform (OS/architecture)
- Go version

#### `clawdepl completion <shell>`

Generate shell autocompletion scripts.

**Supported shells:** bash, zsh, fish, powershell

```bash
# Bash
clawdepl completion bash > /etc/bash_completion.d/clawdepl

# Zsh
clawdepl completion zsh > "${fpath[1]}/_clawdepl"

# Fish
clawdepl completion fish > ~/.config/fish/completions/clawdepl.fish

# PowerShell
clawdepl completion powershell > clawdepl.ps1
```

---

## Command Summary

| Command | Aliases | Description |
|---------|---------|-------------|
| `clawdepl` | - | Setup wizard (login + new) |
| `clawdepl auth` | - | Manage authentication |
| `clawdepl auth login` | `login` | Authenticate with clawdepl.dev |
| `clawdepl auth logout` | `logout` | Sign out and clear credentials |
| `clawdepl auth status` | `account` | Show account information |
| `clawdepl new [name]` | - | Create new instance |
| `clawdepl list` | `ls` | List all instances |
| `clawdepl status <id>` | - | Show instance status |
| `clawdepl start <id>` | - | Start an instance |
| `clawdepl stop <id>` | - | Stop an instance |
| `clawdepl delete <id>` | - | Delete instance(s) |
| `clawdepl ssh <name>` | - | SSH into running instance |
| `clawdepl version` | - | Show version info |
| `clawdepl completion` | - | Generate shell completions |

---

## Configuration

### Credentials Location

Credentials are stored in `~/.clawdepl/credentials.json`.

### Authentication Methods

1. **OAuth (default)** - Opens browser for authentication
2. **OAuth (no-browser)** - Prints URL for manual authentication
3. **Token** - Direct token authentication for CI/CD
4. **API Key** - API key authentication

---

## Examples

### First-time Setup

```bash
# Run the setup wizard
clawdepl

# Or step by step:
clawdepl login
clawdepl new my-first-agent
```

### Managing Instances

```bash
# List all instances
clawdepl list

# Check status of an instance
clawdepl status 8ee265f1-a440-42fe-bcde-44077d8525cc

# Stop an instance
clawdepl stop 8ee265f1-a440-42fe-bcde-44077d8525cc

# Start it again
clawdepl start 8ee265f1-a440-42fe-bcde-44077d8525cc

# Delete an instance
clawdepl delete 8ee265f1-a440-42fe-bcde-44077d8525cc -y
```

### CI/CD Usage

```bash
# Authenticate with a token (no browser needed)
clawdepl login --token $CLAWDEPL_TOKEN

# Or use the auth command
clawdepl auth login --token $CLAWDEPL_TOKEN

# Check account status in JSON format
clawdepl account --json
```

### Scripting

```bash
# Get version info as JSON
clawdepl version --json | jq '.version'

# Get account info as JSON
clawdepl account --json | jq '.email'
```

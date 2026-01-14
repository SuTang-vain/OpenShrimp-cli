# OpenShrimp

A unified CLI and Desktop UI tool for managing AI development tools (Claude, Gemini, OpenCode) on macOS and Linux.

![OpenShrimp UI Preview](https://via.placeholder.com/800x400?text=OpenShrimp+UI)

## Features

- **Tool Discovery**: Automatically detect AI tools installed on your system
- **Cleanup**: Clean up temporary files with configurable retention periods
- **Model Switching**: Switch between different AI models (Claude, MiniMax, GLM)
- **Configuration Backup**: Backup and restore tool configurations
- **Context Sharing**: Unified context management across AI tools
- **Web UI**: Beautiful desktop interface built with Tauri + Vue 3

## Installation

### From Source

```bash
git clone https://github.com/SuTang-vain/OpenShrimp-cli
cd OpenShrimp-cli
make build
sudo mv ai-mgr /usr/local/bin/
```

### Homebrew (Coming Soon)

```bash
brew install sutang-vain/open-shrimp/open-shrimp
```

## CLI Usage

```bash
# Show help
ai-mgr --help

# Scan for AI tools
ai-mgr scan

# Clean up temporary files
ai-mgr cleanup
ai-mgr cleanup --days 3

# Switch AI model
ai-mgr switch claude-sonnet-4

# Backup configurations
ai-mgr backup

# Restore from backup
ai-mgr restore backup_20240114.tar.gz
```

## Desktop UI

### Quick Start

```bash
# Install dependencies
make deps

# Start daemon (required for UI)
./ai-mgr daemon &

# Start UI development server
make ui-dev
# Open http://localhost:3000
```

### Build Desktop App

```bash
# Install Rust + Tauri
make tauri-install

# Build Tauri app
make tauri-build
```

### Production Build

```bash
# Build frontend
make ui-build

# Start daemon with static UI
./ai-mgr daemon
# Open http://127.0.0.1:19999
```

## Commands

| Command | Description |
|---------|-------------|
| `scan` | Scan for AI tools on your system |
| `cleanup` | Clean up temporary files |
| `check` | Health check for AI tools |
| `stats` | Show disk usage statistics |
| `switch` | Switch between AI models |
| `link` | Manage symbolic links |
| `backup` | Backup configurations |
| `restore` | Restore configurations |
| `context` | Manage conversation history |
| `daemon` | Start HTTP server for Web UI |
| `version` | Show version information |

## Configuration

Default config: `~/.ai-manager/config.yaml`

```yaml
version: "1.0.0"
tools:
  claude:
    name: Claude Code
    path: "~/.claude"
    enabled: true
  gemini:
    name: Gemini CLI
    path: "~/.gemini"
    enabled: true

models:
  claude-sonnet-4:
    name: Claude Sonnet 4
    provider: anthropic
  minimax-m2.1:
    name: MiniMax M2.1
    provider: minimax
```

## Web API

When running in daemon mode:

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/health` | GET | Health check |
| `/api/tools` | GET | List tools |
| `/api/tools/{name}/cleanup` | POST | Clean tool |
| `/api/models` | GET | List models |
| `/api/switch` | POST | Switch model |
| `/api/backups` | GET/POST | Manage backups |
| `/api/links` | GET/POST/DELETE | Manage links |
| `/api/stats` | GET | Get statistics |
| `/ws` | WebSocket | Real-time updates |

## Supported Tools

| Tool | Default Path |
|------|--------------|
| Claude Code | ~/.claude |
| Gemini CLI | ~/.gemini |
| OpenCode | ~/.config/opencode |

## Development

```bash
# Build binary
make build

# Run tests
make test

# Install all dependencies
make deps

# Build for all platforms
make release
```

## Architecture

```
main.go          → Entry point
cmd/daemon/      → HTTP + WebSocket server
internal/cli/    → Cobra commands
internal/*/      → Feature modules
ui/              → Vue 3 frontend
tauri/           → Tauri desktop config
```

## License

MIT License

# OpenShrimp

A unified CLI and Desktop UI tool for managing AI development tools (Claude, Gemini, OpenCode) on macOS and Linux.

![OpenShrimp UI Preview](https://via.placeholder.com/800x400?text=OpenShrimp+UI)

## Features

- **Tool Discovery**: Automatically detect AI tools installed on your system
- **Cleanup**: Clean up temporary files with configurable retention periods
- **Model Switching**: Switch between different AI models (Claude, MiniMax, GLM)
- **Configuration Backup**: Backup and restore tool configurations
- **Context Sharing**: Unified context management across AI tools
- **Credential Management**: Secure API key storage with system keychain integration
- **Scheduled Tasks**: Automated cleanup and backup with cron support
- **Web UI**: Beautiful desktop interface built with Tauri + Vue 3
- **CLI**: Full-featured command-line interface

## Installation

### From Source

```bash
git clone https://github.com/sutang/ai-manager
cd ai-manager
make build
sudo mv ai-mgr /usr/local/bin/
```

### Homebrew (Coming Soon)

```bash
brew install sutang-vain/open-shrimp/open-shrimp
```

### Docker

```bash
docker pull openshrimp/ai-manager:latest
docker run -p 19999:19999 -v ~/.ai-manager:/home/app/.ai-manager openshrimp/ai-manager:latest
```

## CLI Usage

```bash
# Show help
ai-mgr --help

# Scan for AI tools
ai-mgr scan
ai-mgr scan --verbose

# Clean up temporary files
ai-mgr cleanup
ai-mgr cleanup --days 3

# Switch AI model
ai-mgr switch claude-sonnet-4
ai-mgr switch --list

# Manage API credentials securely
ai-mgr credentials --list
ai-mgr credentials --set --model claude-sonnet-4 --key ANTHROPIC_API_KEY --value sk-xxx --provider anthropic

# Backup configurations
ai-mgr backup
ai-mgr backup --list

# Restore from backup
ai-mgr restore backup_20240114.tar.gz

# Schedule automatic tasks
ai-mgr scheduler --list
ai-mgr scheduler --add daily-cleanup --type cleanup --schedule "0 0 * * *" --enabled

# Start daemon for Web UI
ai-mgr daemon
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
| `scheduler` | Manage scheduled tasks |
| `credentials` | Manage API credentials securely |
| `daemon` | Start HTTP server for Web UI |
| `version` | Show version information |

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

## Configuration

Default config: `~/.ai-manager/config.yaml`

```yaml
version: "1.0.0"
tools:
  claude:
    name: Claude Code
    path: "~/.claude"
    config_path: settings.json
    data_path: projects
    temp_paths:
      - debug
      - shell-snapshots
    enabled: true
  gemini:
    name: Gemini CLI
    path: "~/.gemini"
    enabled: true

models:
  claude-sonnet-4:
    name: Claude Sonnet 4
    provider: anthropic
    api_endpoint: https://api.anthropic.com
    model_id: claude-sonnet-4-20250514
  minimax-m2.1:
    name: MiniMax M2.1
    provider: minimax
    api_endpoint: https://api.minimaxi.com/anthropic
    model_id: miniMax-M2.1-200k

defaults:
  model: claude-sonnet-4
  cleanup_days: 7

scheduler:
  enabled: true
  tasks:
    daily-cleanup:
      type: cleanup
      schedule: "0 0 * * *"
      enabled: true
```

## Web API

When running in daemon mode:

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/health` | GET | Health check |
| `/api/tools` | GET | List tools |
| `/api/tools/{name}/stats` | GET | Get tool stats |
| `/api/tools/{name}/cleanup` | POST | Clean tool |
| `/api/models` | GET | List models |
| `/api/switch` | POST | Switch model |
| `/api/backups` | GET/POST | List/create backups |
| `/api/backups/{id}/restore` | POST | Restore backup |
| `/api/backups/{id}` | DELETE | Delete backup |
| `/api/links` | GET/POST/DELETE | Manage links |
| `/api/scheduler` | GET/POST | List/create tasks |
| `/api/scheduler/{id}` | PUT/DELETE | Update/delete tasks |
| `/api/scheduler/{id}/run` | POST | Run task manually |
| `/api/credentials` | GET/POST | List/set credentials |
| `/api/credentials/{model}/{key}` | DELETE | Delete credential |
| `/api/credentials/{model}` | GET | Get model credentials |
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

# Run linter
make lint

# Install shell completion
make completion
```

## Architecture

```
main.go              → Entry point
cmd/daemon/          → HTTP + WebSocket server
internal/cli/        → Cobra commands
internal/backup/     → Backup management
internal/cleanup/    → Cleanup operations
internal/config/     → YAML configuration
internal/context/    → SQLite context storage
internal/credentials/→ Credential management (Keychain)
internal/discovery/  → Tool discovery
internal/link/       → Symbolic link management
internal/models/     → Shared models
internal/scheduler/  → Cron task scheduler
internal/switcher/   → Model switching
ui/                  → Vue 3 frontend
  src/components/    → UI components
  src/App.vue        → Main app
  vite.config.ts     → Vite configuration
tauri/               → Tauri desktop config
.github/workflows/   → GitHub Actions CI/CD
```

## CI/CD Pipeline

- **CI Workflow**: Runs on every push/PR
  - Go tests with coverage
  - Multi-platform builds
  - UI type check and build
  - Security scanning (SBOM)
  - Linting (golangci-lint, ESLint)

- **Release Workflow**: Runs on version tag creation
  - Builds all platform binaries
  - Creates GitHub release
  - Updates Homebrew formula
  - Pushes Docker image

## License

MIT License

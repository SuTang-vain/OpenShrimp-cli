# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

OpenShrimp - A unified CLI and Desktop UI tool for managing AI development tools (Claude, Gemini, OpenCode, VSCode) on macOS and Linux. Built with Go using the Cobra framework + Tauri + Vue 3.

## Build Commands

```bash
# Build Go binary
make build

# Run tests
make test

# Install all dependencies (Go + Node.js)
make deps

# Build for all platforms
make release

# Run CLI locally
make run

# Daemon mode (starts HTTP server on port 19999)
make daemon-start
make daemon-stop

# Web UI (requires Node.js)
make ui-deps      # Install UI dependencies
make ui-dev       # Start dev server (port 3000)
make ui-build     # Build production UI

# Tauri Desktop App (requires Rust)
make tauri-install  # Install Rust + Tauri CLI
make tauri-dev      # Run desktop app in dev mode
make tauri-build    # Build desktop app
```

## Architecture

```
main.go          → Entry point, calls cli.Run()
cmd/daemon/      → HTTP + WebSocket server for UI
internal/cli/    → Cobra command definitions and orchestration
internal/config/ → YAML config loading with default fallback
internal/*/      → Feature packages (backup, cleanup, context, discovery, link, models, switcher, utils)
ui/              → Vue 3 frontend (Vite + TailwindCSS)
tauri/           → Tauri desktop app configuration
```

### Key Patterns

- **Commands**: Each CLI command is defined in `internal/cli/commands.go` using Cobra
- **Config**: `internal/config/config.go` loads `~/.ai-manager/config.yaml`, returns defaults if missing
- **Feature Modules**: Each feature area (cleanup, switch, backup, etc.) has its own package with:
  - A manager/handler struct that takes `*Config`
  - Methods that implement the business logic
  - Results returned to CLI for formatting
- **Path Expansion**: Use `utils.ExpandPath()` to handle `~` and environment variables
- **Web API**: `cmd/daemon/server.go` provides REST API and WebSocket for UI

### Adding a New Feature

1. Create `internal/<feature>/feature.go` with a handler struct
2. Add command definition in `internal/cli/commands.go`
3. Register in `internal/cli/root.go:Run()`
4. Add API endpoints in `cmd/daemon/server.go` (if needed for UI)
5. Add Vue component in `ui/src/components/` (if needed for UI)

## Configuration

Default config path: `~/.ai-manager/config.yaml`

The config defines:
- Tools: AI tools to manage (name, paths, temp locations)
- Models: Available AI models with providers/endpoints
- Retention: Cleanup retention policies in days

## Supported Commands

| Command | Purpose |
|---------|---------|
| `scan` | Discover AI tools on the system |
| `cleanup` | Remove temp files older than N days |
| `check` | Health check for configurations |
| `stats` | Disk usage statistics |
| `switch` | Change AI model configuration |
| `link` | Manage symbolic links |
| `backup` | Backup configurations |
| `restore` | Restore from backup |
| `context` | Manage conversation history (SQLite) |
| `version` | Show version |
| `daemon` | Start HTTP server for Web UI |

## Web API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/health` | GET | Health check |
| `/api/tools` | GET | List all tools |
| `/api/tools/{name}/cleanup` | POST | Clean a tool |
| `/api/models` | GET | List models |
| `/api/switch` | POST | Switch model |
| `/api/backups` | GET | List backups |
| `/api/backups` | POST | Create backup |
| `/api/backups/{id}/restore` | POST | Restore backup |
| `/api/backups/{id}` | DELETE | Delete backup |
| `/api/links` | GET | List links |
| `/api/links` | POST | Create link |
| `/api/links/{name}` | DELETE | Remove link |
| `/api/stats` | GET | Get stats |
| `/ws` | WS | WebSocket for real-time updates |

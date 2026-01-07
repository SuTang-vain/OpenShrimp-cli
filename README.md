# AI Tools Manager

A unified CLI tool for managing AI development tools (Claude, Gemini, OpenCode, VSCode) on macOS and Linux.

## Features

- **Tool Discovery**: Automatically detect AI tools installed on your system
- **Cleanup**: Clean up temporary files with configurable retention periods
- **Health Check**: Verify tool configurations and detect issues
- **Statistics**: View disk usage across all AI tools
- **Configuration Management**: Centralized YAML configuration

## Installation

### From Source

```bash
git clone https://github.com/yourusername/ai-manager
cd ai-manager
go build -o ai-mgr .
sudo mv ai-mgr /usr/local/bin/
```

### From Binary

Download the latest release from [GitHub Releases](https://github.com/yourusername/ai-manager/releases).

## Usage

```bash
# Show help
ai-mgr --help

# Scan for AI tools
ai-mgr scan
ai-mgr scan -v  # verbose output

# Clean up temporary files (default: 7 days)
ai-mgr cleanup
ai-mgr cleanup --days 3  # keep last 3 days

# Health check
ai-mgr check

# Show disk usage statistics
ai-mgr stats

# Show version
ai-mgr version
```

## Configuration

The default configuration file is at `~/.ai-manager/config.yaml`.

```yaml
version: "1.0.0"
home_dir: "~/.ai-manager"

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
    api_endpoint: "https://api.anthropic.com"

retention:
  temp_files: 7      # days
  debug_logs: 7      # days
  shell_snapshots: 30 # days
```

## Supported Tools

| Tool | Default Path | Configuration |
|------|--------------|---------------|
| Claude Code | ~/.claude | settings.json |
| Gemini CLI | ~/.gemini | settings.json |
| OpenCode | ~/.config/opencode | settings.json |

## License

MIT

# OpenShrimp CLI

A unified CLI tool for managing AI development tools (Claude, Gemini, OpenCode, VSCode) on macOS and Linux. Part of the OpenShrimp ecosystem.

## Features

- **Tool Discovery**: Automatically detect AI tools installed on your system
- **Cleanup**: Clean up temporary files with configurable retention periods
- **Health Check**: Verify tool configurations and detect issues
- **Statistics**: View disk usage across all AI tools
- **Configuration Management**: Centralized YAML configuration
- **Model Switching**: Switch between different AI models
- **Context Sharing**: Unified context management across AI tools

## Installation

### From Source

```bash
git clone https://github.com/SuTang-vain/OpenShrimp-cli
cd OpenShrimp-cli
go build -o ai-mgr .
sudo mv ai-mgr /usr/local/bin/
```

### Homebrew (Coming Soon)

```bash
brew install sutang-vain/open-shrimp/open-shrimp-cli
```

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

# Switch AI model
ai-mgr switch claude-sonnet-4

# Show version
ai-mgr version
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
| `version` | Show version information |

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
  opencode:
    name: OpenCode
    path: "~/.config/opencode"
    enabled: true

models:
  claude-sonnet-4:
    name: Claude Sonnet 4
    provider: anthropic
    api_endpoint: "https://api.anthropic.com"
  minimax-m2.1:
    name: MiniMax M2.1
    provider: minimax
    api_endpoint: "https://api.minimaxi.com/anthropic"
  glm-4.7:
    name: GLM-4.7
    provider: zhipu
    api_endpoint: "https://open.bigmodel.cn/api/anthropic"

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
| VSCode | ~/Library/.../workspaceStorage | User settings |

## Environment Variables

| Variable | Description |
|----------|-------------|
| `AI_MGR_CONFIG` | Path to config file |
| `ANTHROPIC_API_KEY` | Anthropic API key |
| `MINIMAX_API_KEY` | MiniMax API key |
| `ZHIPU_API_KEY` | Zhipu AI API key |

## Development

```bash
# Build
make build

# Test
make test

# Install dependencies
make deps

# Release build for all platforms
make release
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License - see [LICENSE](LICENSE) for details.

## Related Projects

- [OpenShrimp UI](https://github.com/SuTang-vain/OpenShrimp-ui) - Web Dashboard
- [AI Central](https://github.com/SuTang-vain/AI-Central) - Centralized AI config management

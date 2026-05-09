---
name: openai-codex
description: >
  Manual (not a tool) for OpenAI Codex CLI — OpenAI's coding agent that runs
  locally from your terminal. Built in Rust for speed and efficiency. Supports
  headless remote control, Vim editing, plugin management, hooks, and Chrome
  browser integration. Read this when the human asks to use OpenAI Codex CLI,
  wants to compare it with Claude Code, or needs help with installation and
  configuration.
version: 1.0.0
---

# OpenAI Codex CLI

> **OpenAI's coding agent — run locally from your terminal.**
> Built in Rust for speed. Open source. ~4 million weekly active users (as of April 2026).

## Installation

```bash
npm install -g @openai/codex@0.130.0
```

Update existing installation:
```bash
codex update
# or
npm i -g @openai/codex@latest
```

## Configuration

### API Key
Set your OpenAI API key:
```bash
export OPENAI_API_KEY="your-api-key"
```

Or configure in `~/.codex/config.toml`:
```toml
[api]
key = "your-api-key"
```

### Models
Codex CLI supports multiple models:
- GPT-5.5 (latest, recommended)
- GPT-5.4
- GPT-5.3-Codex (specialized for coding)

Configure in `config.toml`:
```toml
[model]
default = "gpt-5.5"
```

### Bedrock Auth
For AWS Bedrock, use console-login credentials:
```bash
aws login
codex exec "your prompt"
```

## Key Features

### 1. Remote Control
New in 0.130.0 — headless, remotely controllable app-server:
```bash
codex remote-control
```
- Start a headless app-server
- Control Codex remotely
- Page large threads with different view modes (unloaded/summary/full)

### 2. Vim Editing
Full Vim modal editing in the TUI:
```bash
codex exec "your prompt"
# In TUI:
/vim                    # Toggle Vim mode
:set default-mode=insert  # Set default mode
```

### 3. Plugin Management
Workspace sharing and marketplace:
```bash
codex plugins list      # List installed plugins
codex plugins install   # Install from marketplace
codex plugins share     # Share with workspace
```

Features:
- Workspace sharing with access controls
- Source filtering and local share path tracking
- Marketplace removal/upgrades
- Remote bundle sync
- Admin-disabled status handling

### 4. Hooks
Browseable and toggleable hooks:
```bash
codex hooks list        # List available hooks
codex hooks toggle      # Toggle hook on/off
```

Capabilities:
- Before/after compaction support
- PreToolUse context injection
- Codex Apps auth integration
- MCP elicitations through TUI/Guardian flows

### 5. Chrome Extension
Browser integration without takeover:
- Works in parallel across tabs
- Background operation
- User controls which websites Codex can use
- Install from Chrome Web Store

### 6. App-Server
Thread management and pagination:
```bash
codex exec "your prompt"
# In TUI:
# - Resume/fork picker
# - Raw scrollback mode
# - /ide context injection
# - /diff workspace-aware diffing
```

## Usage Examples

### Basic Usage
```bash
# Start interactive session
codex exec "Create a Python script that reads CSV files"

# With specific model
codex exec --model gpt-5.5 "Refactor this function"

# In specific directory
codex exec --dir /path/to/project "Fix the bug in main.py"
```

### Remote Control
```bash
# Start headless server
codex remote-control

# Connect from another terminal
codex connect localhost:8080
```

### Plugin Management
```bash
# List plugins
codex plugins list

# Install plugin
codex plugins install @openai/plugin-name

# Share with workspace
codex plugins share ./my-plugin
```

### Hooks
```bash
# List hooks
codex hooks list

# Toggle hook
codex hooks toggle my-hook

# Run with hooks
codex exec --hooks my-hook "your prompt"
```

## Integration with LingTai

### Workflow Integration
Codex CLI can be used alongside Claude Code for different tasks:

| Task | Claude Code | OpenAI Codex |
|------|-------------|--------------|
| Complex reasoning | ✅ Excellent | ✅ Good |
| Local file operations | ✅ Good | ✅ Excellent |
| Browser integration | ❌ No | ✅ Chrome extension |
| Remote control | ❌ No | ✅ Yes |
| Plugin ecosystem | ❌ Limited | ✅ Rich marketplace |

### When to Use Codex CLI
- **Browser automation**: Use Chrome extension for web tasks
- **Remote development**: Use remote-control for headless operation
- **Plugin ecosystem**: Leverage marketplace for specialized tools
- **Vim users**: Native Vim editing support

### When to Use Claude Code
- **Complex reasoning**: Deep analysis and multi-step problem solving
- **LingTai integration**: Native integration with LingTai kernel
- **Cost efficiency**: Uses Claude Max subscription

## Comparison with Claude Code

| Feature | OpenAI Codex CLI | Claude Code |
|---------|------------------|-------------|
| Language | Rust | TypeScript |
| Open Source | ✅ Yes | ❌ No |
| Vim Support | ✅ Native | ❌ No |
| Browser Extension | ✅ Chrome | ❌ No |
| Remote Control | ✅ Yes | ❌ No |
| Plugin Marketplace | ✅ Rich | ❌ Limited |
| LingTai Integration | ❌ No | ✅ Native |
| Cost | API usage | Claude Max subscription |

## Troubleshooting

### Common Issues

1. **Installation fails**
   ```bash
   # Clear npm cache
   npm cache clean --force
   # Reinstall
   npm install -g @openai/codex@0.130.0
   ```

2. **API key not found**
   ```bash
   # Check environment variable
   echo $OPENAI_API_KEY
   # Or check config file
   cat ~/.codex/config.toml
   ```

3. **Plugin installation fails**
   ```bash
   # Check marketplace connectivity
   codex plugins search
   # Clear plugin cache
   rm -rf ~/.codex/plugins/cache
   ```

## Resources

- **GitHub**: https://github.com/openai/codex
- **Documentation**: https://developers.openai.com/codex
- **Changelog**: https://developers.openai.com/codex/changelog
- **Chrome Extension**: Available on Chrome Web Store

## Version History

- **0.130.0** (May 8, 2026): Remote control, plugin hooks, Bedrock auth
- **0.129.0** (May 7, 2026): Vim editing, Chrome extension, plugin management
- **0.128.0** (May 5, 2026): /goal command, Ralph loop

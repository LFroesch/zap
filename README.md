# zap ⚡

A lightning-fast TUI file registry for developers. Organize and instantly access your most important files across all projects.

## What is zap?

zap is your personal file shortcut manager. Instead of hunting through directories for that config file, database schema, or API spec you need to edit, just register it once and `zap` to it instantly.

## Features

- **Lightning Fast** - Open any registered file in your preferred editor with one keystroke
- **Project Organization** - Group files by project for easy browsing
- **Multiple File Types** - JSON, YAML, TOML, configs, docs, scripts - anything text-based
- **Smart Editor Detection** - Automatically uses VS Code, nano, vim, or vi
- **Live Editing** - Edit file metadata directly in the TUI

## Installation

```bash
# Build
go build -o zap main.go

# Install globally
cp zap ~/.local/bin/

# Make sure ~/.local/bin is in PATH
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

## Usage

```bash
zap
```

### Quick Commands

- `↑↓` - Navigate files
- `space/enter` - Open file in editor
- `e` - Edit file details
- `n/a` - Add new file
- `d` - Delete file
- `r` - Refresh
- `q` - Quit

## Examples

Perfect for managing:
- Config files (`nginx.conf`, `docker-compose.yml`)
- Data files (`users.json`, `api-endpoints.yaml`) 
- Documentation (`README.md`, `API.md`)
- Scripts (`deploy.sh`, `backup.sql`)
- Dotfiles (`.vimrc`, `.bashrc`)

## File Storage

Files are registered in `~/.config/zap/zap-registry.json` but zap doesn't move or copy your files - it just creates shortcuts to them.

## Why zap?

Stop wasting time navigating to frequently used files. Register them once, access them instantly. ⚡
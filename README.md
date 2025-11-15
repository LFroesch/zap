# zap ⚡

A lightning-fast, feature-rich TUI file registry for developers. Organize and instantly access your most important files across all projects with powerful search, fuzzy finding, and smart navigation.

## What is zap?

zap is your personal file shortcut manager on steroids. Instead of hunting through directories for that config file, database schema, or API spec you need to edit, just register it once and `zap` to it instantly. With advanced features like fuzzy finding, recently opened tracking, and vim-style navigation, it's the ultimate productivity tool for developers who work with many files across multiple projects.

## Features

### Core Features
- **Lightning Fast** - Open any registered file in your preferred editor with one keystroke
- **Project Organization** - Group files by project for easy browsing
- **Multiple File Types** - JSON, YAML, TOML, configs, docs, scripts - anything text-based
- **Smart Editor Detection** - Automatically uses VS Code, nvim, vim, nano, or vi
- **Live Status Indicators** - See at a glance which files exist, are broken, or were recently opened

### Search & Discovery
- **Powerful Search** - Press `/` to search across all file metadata (name, path, project, description)
- **Fuzzy Finding** - Press `Ctrl+F` for lightning-fast fuzzy file finding
- **Live Search Results** - See matches update in real-time as you type
- **Smart Filtering** - Searches across name, project, type, path, and description

### Navigation
- **Vim-Style Keys** - Use `j/k` for navigation, `g/G` for top/bottom
- **Page Navigation** - `PageUp/PageDown` or `Ctrl+U/Ctrl+D` for fast scrolling
- **Arrow Keys** - Traditional `↑/↓` navigation also supported
- **Home/End** - Jump to first or last file instantly
- **Column Scrolling** - Use `←/→` or `h/l` to scroll through columns

### Smart Features
- **Auto-detect File Type** - Automatically detects file type from extension
- **Duplicate Prevention** - Prevents registering the same file twice
- **Recently Opened Tracking** - Press `s` to toggle sort by recently opened
- **Path Validation** - Shows broken file paths with ❌ indicator
- **Atomic Saves** - Config changes are saved atomically to prevent corruption

### Organization & Workflow
- **Sorted Views** - Sort by project (default) or recently opened
- **Visual Project Headers** - Clear project groupings with 📂 headers
- **Status Icons** - ✓ for opened files, ❌ for broken paths
- **Quick Edit Mode** - Tab through fields when editing metadata
- **Confirmation Prompts** - Safety confirmation before deleting files

### Help & Usability
- **Built-in Help** - Press `?` to see all available commands
- **Context-aware UI** - Footer shows relevant commands for current mode
- **Status Messages** - Color-coded feedback for all operations
- **Path Expansion** - Supports `~/` for home directory

## Installation

```bash
# Build
go build -o zap main.go

# Install globally (Linux/macOS)
sudo cp zap /usr/local/bin/

# Or install to user directory
cp zap ~/.local/bin/

# Make sure ~/.local/bin is in PATH
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

## Usage

```bash
zap
```

### Quick Start

1. **Add your first file**: Press `n` or `a`
2. **Fill in details**: Tab through Name, Project, Type, Path, Description
3. **Save**: Press Enter
4. **Open file**: Navigate with arrows and press Space or Enter
5. **Search**: Press `/` to search or `Ctrl+F` for fuzzy find

### All Commands

#### Navigation
- `↑/k` - Move up
- `↓/j` - Move down
- `PageUp / Ctrl+u` - Page up
- `PageDown / Ctrl+d` - Page down
- `Home / g` - Go to top
- `End / G` - Go to bottom
- `← / h` - Scroll columns left
- `→ / l` - Scroll columns right

#### Actions
- `space / enter` - Open file in editor
- `e` - Edit file metadata
- `n / a` - Add new file
- `d` - Delete file (with confirmation)
- `r` - Refresh list from disk

#### Search & Filter
- `/` - Start search mode
- `Ctrl+f` - Start fuzzy find mode
- `esc` - Clear search/exit mode
- `enter` - Apply search filter

#### Sorting
- `s` - Toggle sort (project/recent)

#### System
- `?` - Show help screen
- `q / Ctrl+c` - Quit

### Edit Mode Navigation

When editing or adding a file:
- `tab` - Next field
- `shift+tab` - Previous field
- `enter` - Save changes
- `esc` - Cancel (removes unsaved new entries)

## Examples

Perfect for managing:
- **Config files** (`nginx.conf`, `docker-compose.yml`, `.env`)
- **Data files** (`users.json`, `api-endpoints.yaml`, `schema.sql`)
- **Documentation** (`README.md`, `API.md`, `CHANGELOG.md`)
- **Scripts** (`deploy.sh`, `backup.sql`, `setup.py`)
- **Dotfiles** (`.vimrc`, `.bashrc`, `.gitconfig`)
- **Project files** (`package.json`, `Cargo.toml`, `go.mod`)

## File Storage

Files are registered in `~/.config/zap/zap-registry.json` but zap doesn't move or copy your files - it just creates shortcuts to them. The registry stores:

```json
{
  "configs": [
    {
      "name": "API Config",
      "path": "/home/user/projects/api/config.yaml",
      "type": "yaml",
      "project": "MyAPI",
      "description": "Main API configuration",
      "last_opened": "2025-11-15T10:30:00Z"
    }
  ]
}
```

## Architecture

The codebase is organized into clean, modular packages:

```
zap/
├── main.go                 # Main application & UI logic
├── internal/
│   ├── models/            # Data structures
│   │   └── config.go
│   ├── storage/           # Config persistence
│   │   └── storage.go
│   ├── editor/            # Editor integration
│   │   └── editor.go
│   └── ui/                # UI components
│       ├── help.go
│       └── styles.go
```

## Why zap?

Stop wasting time navigating to frequently used files. As developers, we often work across multiple projects with important files scattered everywhere. zap gives you instant access to all of them from a single, beautiful interface.

### Before zap:
```bash
cd ~/projects/api
vim config/database.yaml
# Later...
cd ~/projects/frontend
code .env.production
# Later...
cd ~/dotfiles
nano .vimrc
```

### With zap:
```bash
zap
# Press / to search "database" → Enter → Opens instantly
# Press / to search "env" → Enter → Opens instantly
# Press / to search "vim" → Enter → Opens instantly
```

## Features at a Glance

| Feature | Status |
|---------|--------|
| Multi-editor support | ✅ |
| Fuzzy finding | ✅ |
| Recently opened tracking | ✅ |
| Search & filter | ✅ |
| Vim-style navigation | ✅ |
| Project grouping | ✅ |
| Path validation | ✅ |
| Duplicate prevention | ✅ |
| Atomic saves | ✅ |
| Auto file-type detection | ✅ |
| Help screen | ✅ |
| Live search | ✅ |
| Performance caching | ✅ |

## Contributing

Contributions welcome! This is a focused tool for developer productivity. When contributing:

- Keep it fast
- Keep it keyboard-driven
- Keep the code clean and modular
- Test thoroughly

## License

MIT

## Tips & Tricks

1. **Use descriptive names**: Make files easy to find with clear, searchable names
2. **Add good descriptions**: The description field is searchable - use it!
3. **Organize by project**: Group related files by project for better organization
4. **Leverage fuzzy find**: `Ctrl+F` then type just the key letters (e.g., "dc" finds "docker-compose.yml")
5. **Check recently opened**: Press `s` to see your most-used files at the top
6. **Use projects wisely**: Group files by project, not by type

---

Built with ❤️ using [Bubble Tea](https://github.com/charmbracelet/bubbletea) by Charm

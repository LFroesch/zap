# zap

A lightning-fast TUI file registry for developers. Organize and instantly access your most important files across all projects.

## Quick Install

Supported platforms: Linux and macOS. On Windows, use WSL.

Recommended (installs to `~/.local/bin`):

```bash
curl -fsSL https://raw.githubusercontent.com/LFroesch/zap/main/install.sh | bash
```

Or download a binary from [GitHub Releases](https://github.com/LFroesch/zap/releases).

Or install with Go:

```bash
go install github.com/LFroesch/zap@latest
```

Or build from source:

```bash
make install
```

Command:

```bash
zap
```
## What is zap?

zap is a personal file shortcut manager. Register files once, then open them instantly from a single interface with search, sorting, and vim-style navigation.

## Features

- **Instant access** - Open any registered file in your configured editor with one keystroke
- **Project organization** - Group files by project
- **Search** - Live search across all file metadata
- **5 sort modes** - Project, Recent, Name, Type, Path
- **Vim-style navigation** - j/k, g/G, ctrl+d/u
- **Configurable editor** - Set your preferred editor in the config file
- **Auto file-type detection** - Detects type from extension
- **Duplicate prevention** - Prevents registering the same file twice
- **Atomic saves** - Config changes saved atomically to prevent corruption

## Installation

```bash
go install github.com/LFroesch/zap@latest
```

Make sure `$GOPATH/bin` (usually `~/go/bin`) is in your PATH:
```bash
export PATH="$HOME/go/bin:$PATH"
```

## Usage

```bash
zap
```

### Quick Start

1. Press `N` to add a file
2. Tab through Name, Project, Type, Path, Description
3. Press Enter to save
4. Navigate with j/k and press `o` or Enter to open

## Editor Configuration

zap uses the `editor` field in `~/.config/zap/zap-registry.json` to determine which editor to use. It defaults to `code` (VS Code) if not set.

To change your editor, either:
- Press `,` in zap to open the config file and add/edit the `"editor"` field
- Or edit `~/.config/zap/zap-registry.json` directly

```json
{
  "editor": "code",
  "configs": [...]
}
```

Supported values include any editor command in your PATH:
- `code` - VS Code (default)
- `cursor` - Cursor
- `nvim` - Neovim
- `vim` - Vim
- `nano` - Nano
- `emacs` - Emacs

Terminal editors (nvim, vim, vi, nano, emacs) run inside the TUI. GUI editors (code, cursor, etc.) launch in the background.

## Keybindings

### Navigation

| Key | Action |
|-----|--------|
| `j/k`, `up/down` | Navigate |
| `g/G` | First/last item |
| `ctrl+d/u` | Half-page scroll |

### Actions

| Key | Action |
|-----|--------|
| `enter/o` | Open file in editor |
| `O` | Open parent directory in editor |
| `e` | Edit file metadata |
| `N` | Add new file |
| `D` | Delete file (confirmation) |
| `y` | Copy path to clipboard |
| `r` | Refresh list |

### Search & Sort

| Key | Action |
|-----|--------|
| `/` | Search |
| `S` | Cycle sort: Project/Recent/Name/Type/Path |

### Edit Mode

| Key | Action |
|-----|--------|
| `tab/shift+tab` | Next/previous field |
| `enter` | Save |
| `esc` | Cancel |

### System

| Key | Action |
|-----|--------|
| `,` | Open config |
| `?` | Help |
| `q/ctrl+c` | Quit |

## File Storage

Files are registered in `~/.config/zap/zap-registry.json`. Zap doesn't move or copy your files - it creates shortcuts to them.

```json
{
  "editor": "code",
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

```
zap/
├── main.go
├── model.go
├── update.go
├── view.go
├── helpers.go
├── internal/
│   ├── models/config.go
│   ├── storage/storage.go
│   ├── editor/editor.go
│   └── ui/
│       ├── help.go
│       └── styles.go
```

## License

[AGPL-3.0](LICENSE)

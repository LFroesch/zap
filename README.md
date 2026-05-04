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
- **4 sort modes** - Project, Recent, Name, Path
- **Dual-panel layout** - Browse files on the left with metadata + preview on the right
- **Responsive terminal sizing** - Fills the current terminal viewport and adapts cleanly on resize
- **Vim-style navigation** - j/k, g/G, ctrl+d/u
- **Editor via env** - Uses `$VISUAL` / `$EDITOR`, falls back to `code`
- **Auto file-type detection** - Detects type from extension
- **Duplicate prevention** - Prevents registering the same file twice
- **Atomic saves** - Config changes saved atomically to prevent corruption

## Usage

### Quick Start

1. Press `N` to add a file
2. Tab through Name, Project, Path, Description
3. Press Enter to save
4. Navigate with j/k and press `o` or Enter to open

## Editor Configuration

zap resolves the editor via `$VISUAL` → `$EDITOR` → `code`.

```bash
export EDITOR=nvim      # terminal editor
export VISUAL=cursor    # GUI editor (checked first)
```

Add to your `~/.zshrc` or `~/.bashrc`. Terminal editors (nvim, vim, nano, emacs) run inside the TUI; GUI editors (code, cursor, etc.) launch in the background.

## Keybindings

### Navigation

| Key | Action |
|-----|--------|
| `j/k`, `up/down` | Navigate left file list |
| `g/G` | First/last item |
| `ctrl+d/u` | Half-page list scroll |
| `J/K` | Scroll right preview pane |
| `pgup/pgdn` | Page scroll right preview pane |
| `ctrl+home/end` | Preview top / bottom |

### Actions

| Key | Action |
|-----|--------|
| `enter/o` | Open file in external editor |
| `E` | Edit selected file inline |
| `O` | Open parent directory in editor |
| `e` | Edit selected file metadata |
| `N` | Add new file |
| `D` | Delete file (confirmation) |
| `y` | Copy path to clipboard |
| `r` | Refresh list |

### Search & Sort

| Key | Action |
|-----|--------|
| `/` | Search |
| `S` | Cycle sort: Project/Recent/Name/Path |

### Edit Mode

| Key | Action |
|-----|--------|
| `tab/shift+tab` | Next/previous field |
| `enter` | Save |
| `esc` | Cancel |

### Inline File Edit Mode

| Key | Action |
|-----|--------|
| `ctrl+s` | Save file |
| `ctrl+d` | Delete current line |
| `esc` | Cancel |

### System

| Key | Action |
|-----|--------|
| `,` | Open config |
| `?` | Help |
| `q/ctrl+c` | Quit |

### Help Mode

| Key | Action |
|-----|--------|
| `j/k`, `up/down` | Scroll help |
| `pgup/pgdn` | Page scroll help |
| `g/G` | Top/bottom of help |
| `esc`, `?`, `q` | Close help |

## File Storage

Files are registered in `~/.config/zap/zap-registry.json`. Zap doesn't move or copy your files - it creates shortcuts to them.

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

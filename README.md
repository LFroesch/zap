# zap

Personal file registry for the terminal. `zap` lets you register important files once, organize them by project, preview them quickly, and reopen them without hunting through directories every time.

![zap hero screenshot](assets/screenshots/hero.png)

**Live demo:** [froesch.dev](https://froesch.dev)

## Install

Supported platforms: Linux and macOS.

Windows release binaries and installer entrypoints are shipped, but native Windows support is unverified.

Recommended:

```bash
curl -fsSL https://raw.githubusercontent.com/LFroesch/zap/main/install.sh | bash
```

Windows:

```powershell
./install.ps1
```

```bat
install.cmd
```

Other options:

```bash
go install github.com/LFroesch/zap@latest
make install
```

Run:

```bash
zap
zap --version
```

## What It Stores

Registered files are saved in:

```text
~/.config/zap/zap-registry.json
```

`zap` does not move or copy your files. It only stores metadata and paths.

## Features

- Register files with a name, project, path, and description
- Search across saved file metadata
- Sort by project, recent, name, or path
- Preview file content in a right-hand pane
- Open the file or its parent directory in your editor
- Edit file metadata or edit the file inline
- Prevent duplicate registrations and save registry changes atomically

Editor resolution order:

```text
$VISUAL -> $EDITOR -> code
```

## Quick Start

1. Press `N`
2. Fill in the file metadata
3. Save
4. Press `enter` or `o` to open the selected file

## Controls

| Key | Action |
|-----|--------|
| `j/k` | Move |
| `g/G` | Top or bottom |
| `/` | Search |
| `S` | Change sort |
| `enter`, `o` | Open file |
| `O` | Open parent directory |
| `N` | Add file |
| `e` | Edit metadata |
| `E` | Edit file inline |
| `D` | Delete |
| `y` | Copy path |
| `r` | Refresh |
| `,` | Open config |
| `?` | Help |
| `q` | Quit |

## License

[AGPL-3.0](LICENSE)

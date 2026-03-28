# goclip

![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go&logoColor=white)
![Platforms](https://img.shields.io/badge/platform-macOS%20%7C%20Linux%20%7C%20Windows-1f6feb)
![Interface](https://img.shields.io/badge/interface-CLI%20%2B%20TUI-2ea44f)

`goclip` is a local-first clipboard history manager for the terminal. It gives you a fast command-line workflow for saving and searching clips, plus a polished Bubble Tea interface for browsing, adding, copying, and deleting items without leaving your terminal.

## Why goclip

- Keep a searchable history of copied text on your machine
- Jump between quick CLI commands and a full-screen TUI
- Install with a single terminal command on macOS, Linux, or Windows PowerShell
- Export, trim, and remove your clipboard history cleanly when you need to

## Highlights

| Capability | What you get |
| --- | --- |
| Local storage | Clipboard history is stored in BoltDB on your machine |
| Terminal-first workflow | Add clips from args, pipes, search queries, or the interactive UI |
| Fast TUI | Live search, keyboard navigation, add, copy, and delete actions |
| Cross-platform delivery | Release builds and installer scripts for macOS, Linux, and Windows |
| Clean removal | `goclip uninstall` removes the binary plus local history/config |
| Configurable | Tune max history, DB location, and preview length with YAML |

## Installation

`goclip` is designed to be installed without cloning the repository. The installers download the latest tagged release from GitHub Releases.

### macOS / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/Ramyprojs/go-clip/main/scripts/install.sh | sh
```

### Windows PowerShell

```powershell
irm https://raw.githubusercontent.com/Ramyprojs/go-clip/main/scripts/install.ps1 | iex
```

### Install a specific version

```bash
curl -fsSL https://raw.githubusercontent.com/Ramyprojs/go-clip/main/scripts/install.sh | GOCLIP_VERSION=v0.2.0 sh
```

```powershell
$env:GOCLIP_VERSION = "v0.2.0"
irm https://raw.githubusercontent.com/Ramyprojs/go-clip/main/scripts/install.ps1 | iex
```

### Build from source

```bash
git clone https://github.com/Ramyprojs/go-clip.git
cd go-clip
make build
```

### Verify the install

```bash
goclip version
goclip
```

## Quick Start

```bash
goclip add "API token for staging"
printf "ssh user@host\n" | goclip add
goclip list
goclip search ssh
goclip
```

Running `goclip` with no subcommand launches the TUI by default.

## TUI Controls

| Key | Action |
| --- | --- |
| `A` | Start add mode and type a new clip |
| `Enter` | Copy the selected clip, or save a new clip while add mode is active |
| `D` | Delete the selected clip |
| `Up` / `Down` | Move through the clip list |
| `Q` / `Ctrl+C` | Quit |

## Commands

| Command | Description | Example |
| --- | --- | --- |
| `goclip` | Launch the TUI | `goclip` |
| `goclip ui` | Launch the TUI explicitly | `goclip ui` |
| `goclip add` | Save a clip from an argument or stdin | `echo "hello" \| goclip add` |
| `goclip list` | Print recent clips | `goclip list --limit 10` |
| `goclip search` | Fuzzy-search history | `goclip search token` |
| `goclip delete` | Delete a clip by displayed index | `goclip delete 3` |
| `goclip clear` | Remove all stored clips after confirmation | `goclip clear` |
| `goclip export` | Export clips to JSON or text | `goclip export --format=json --output=clips.json` |
| `goclip version` | Print the current version | `goclip version` |
| `goclip uninstall` | Remove the binary and local data | `goclip uninstall` |

## Configuration

`goclip` reads configuration from `~/.goclip/config.yaml`.

| Option | Default | Description |
| --- | --- | --- |
| `max_history` | `500` | Maximum number of stored clips before older items are trimmed |
| `db_path` | `~/.goclip/history.db` | Custom location for the BoltDB history file |
| `preview_length` | `60` | Preview length used by list, search, and TUI views |

Example:

```yaml
max_history: 500
db_path: ~/.goclip/history.db
preview_length: 60
```

## Uninstall

To remove `goclip` completely:

```bash
goclip uninstall
```

That command removes:

- the installed `goclip` binary
- local clipboard history
- local configuration in `~/.goclip`

## Development

### Make targets

```bash
make build
make cross-build
make run
make test
make clean
make install
```

### Release flow

Pushing a tag like `v0.2.0` triggers [`.github/workflows/release.yml`](.github/workflows/release.yml), which builds release archives for Linux, macOS, and Windows and uploads them to GitHub Releases. The one-line install scripts use those release assets.

## Screenshots

Add terminal screenshots or an asciinema demo here to show the TUI in action.

# goclip

`goclip` is a terminal clipboard manager written in Go. It stores clipboard history locally, supports fuzzy search from the command line, and includes an interactive Bubble Tea TUI for browsing, copying, and deleting clips.

## Features

- Local clipboard history stored in BoltDB
- `goclip add` for saving text from arguments or piped stdin
- `goclip list`, `search`, `delete`, `clear`, and `export` commands
- Interactive TUI with live search, copy, delete, and quit keybindings
- Configurable history size, DB path, and preview length through YAML config

## Installation

### Clone and build

```bash
git clone https://github.com/Ramyprojs/go-clip.git
cd go-clip
make build
```

The binary will be written to `bin/goclip`.

### Install with Go

```bash
go install .
```

## Quick Start

```bash
goclip add "hello from the terminal"
printf "hello from stdin\n" | goclip add
goclip list
goclip search hello
goclip
```

Running `goclip` with no subcommand launches the TUI by default.

## Configuration

`goclip` reads configuration from `~/.goclip/config.yaml`.

Available options:

- `max_history`: maximum number of stored clips before older items are trimmed
- `db_path`: path to the BoltDB history file
- `preview_length`: preview length used by list, search, and TUI views

Example:

```yaml
max_history: 500
db_path: ~/.goclip/history.db
preview_length: 60
```

## Commands

### `goclip`

Launch the interactive TUI.

```bash
goclip
```

### `goclip ui`

Launch the interactive TUI explicitly.

```bash
goclip ui
```

### `goclip add`

Add a clip from an argument or stdin.

```bash
goclip add "important snippet"
echo "copied from a pipe" | goclip add
```

### `goclip list`

List stored clips in terminal output or JSON.

```bash
goclip list
goclip list --limit 10
goclip list --json
```

### `goclip search`

Fuzzy-search stored clips.

```bash
goclip search ssh
goclip search "api token"
```

### `goclip delete`

Delete a clip by its displayed index.

```bash
goclip delete 3
```

### `goclip clear`

Delete all stored history after confirmation.

```bash
goclip clear
```

### `goclip export`

Export the full history to JSON or plain text.

```bash
goclip export --format=json --output=clips.json
goclip export --format=txt --output=clips.txt
```

### `goclip version`

Print the application version.

```bash
goclip version
```

## Make Targets

```bash
make build
make run
make test
make clean
make install
```

## Screenshots

TUI screenshots can be added here once you capture the interactive interface.

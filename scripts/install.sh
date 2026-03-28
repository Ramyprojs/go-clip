#!/usr/bin/env sh
set -eu

REPO="${GOCLIP_REPO:-Ramyprojs/go-clip}"
VERSION="${GOCLIP_VERSION:-latest}"
INSTALL_DIR="${GOCLIP_INSTALL_DIR:-}"

usage() {
	cat <<'EOF'
Install goclip from GitHub Releases.

Usage:
  install.sh [--bin-dir <dir>] [--version <tag>]

Environment overrides:
  GOCLIP_INSTALL_DIR   Directory to install the goclip binary into
  GOCLIP_VERSION       Release tag to install (default: latest)
  GOCLIP_REPO          GitHub repo in owner/name form
EOF
}

path_contains() {
	case ":$PATH:" in
		*":$1:"*) return 0 ;;
		*) return 1 ;;
	esac
}

choose_install_dir() {
	if [ -n "$INSTALL_DIR" ]; then
		printf '%s\n' "$INSTALL_DIR"
		return
	fi

	if command -v goclip >/dev/null 2>&1; then
		existing_dir=$(dirname "$(command -v goclip)")
		if [ -w "$existing_dir" ]; then
			printf '%s\n' "$existing_dir"
			return
		fi
	fi

	if [ -n "${HOME:-}" ] && path_contains "$HOME/.local/bin"; then
		printf '%s\n' "$HOME/.local/bin"
		return
	fi

	if [ -n "${HOME:-}" ] && path_contains "$HOME/bin"; then
		printf '%s\n' "$HOME/bin"
		return
	fi

	if [ -d /usr/local/bin ] && [ -w /usr/local/bin ] && path_contains "/usr/local/bin"; then
		printf '%s\n' "/usr/local/bin"
		return
	fi

	printf '%s\n' "${HOME:-.}/.local/bin"
}

detect_os() {
	case "$(uname -s)" in
		Linux) printf '%s\n' "linux" ;;
		Darwin) printf '%s\n' "darwin" ;;
		*)
			printf '%s\n' "unsupported operating system: $(uname -s)" >&2
			exit 1
			;;
	esac
}

detect_arch() {
	case "$(uname -m)" in
		x86_64|amd64) printf '%s\n' "amd64" ;;
		arm64|aarch64) printf '%s\n' "arm64" ;;
		*)
			printf '%s\n' "unsupported architecture: $(uname -m)" >&2
			exit 1
			;;
	esac
}

download() {
	url=$1
	destination=$2

	if command -v curl >/dev/null 2>&1; then
		curl -fsSL "$url" -o "$destination"
		return
	fi

	if command -v wget >/dev/null 2>&1; then
		wget -qO "$destination" "$url"
		return
	fi

	printf '%s\n' "curl or wget is required to install goclip" >&2
	exit 1
}

while [ $# -gt 0 ]; do
	case "$1" in
		-b|--bin-dir)
			if [ $# -lt 2 ]; then
				printf '%s\n' "missing value for $1" >&2
				exit 1
			fi
			INSTALL_DIR=$2
			shift 2
			;;
		-v|--version)
			if [ $# -lt 2 ]; then
				printf '%s\n' "missing value for $1" >&2
				exit 1
			fi
			VERSION=$2
			shift 2
			;;
		-h|--help)
			usage
			exit 0
			;;
		*)
			printf '%s\n' "unknown argument: $1" >&2
			exit 1
			;;
	esac
done

OS=$(detect_os)
ARCH=$(detect_arch)
ARCHIVE="goclip_${OS}_${ARCH}.tar.gz"

if [ "$VERSION" = "latest" ]; then
	URL="https://github.com/${REPO}/releases/latest/download/${ARCHIVE}"
else
	URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE}"
fi

TMPDIR=$(mktemp -d 2>/dev/null || mktemp -d -t goclip)
cleanup() {
	rm -rf "$TMPDIR"
}
trap cleanup EXIT INT TERM

INSTALL_DIR=$(choose_install_dir)
ARCHIVE_PATH="$TMPDIR/$ARCHIVE"
BINARY_PATH="$INSTALL_DIR/goclip"

printf '%s\n' "Downloading ${URL}"
if ! download "$URL" "$ARCHIVE_PATH"; then
	printf '%s\n' "Unable to download a goclip release. Publish a tagged GitHub release before using this installer." >&2
	exit 1
fi

mkdir -p "$INSTALL_DIR"
tar -xzf "$ARCHIVE_PATH" -C "$TMPDIR"
cp "$TMPDIR/goclip" "$BINARY_PATH"
chmod 755 "$BINARY_PATH"

printf '%s\n' "Installed goclip to $BINARY_PATH"
if ! path_contains "$INSTALL_DIR"; then
	printf '%s\n' "Add $INSTALL_DIR to your PATH to run goclip from any terminal."
fi

"$BINARY_PATH" version

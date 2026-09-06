#!/usr/bin/env sh
# This file written by AI :)
#
# Install the latest ttype release.
#
#   curl -fsSL https://raw.githubusercontent.com/alirezaudev/ttype/main/install.sh | sh
#
# Environment:
#   TTYPE_VERSION   install this version instead of the latest (e.g. 1.0.0)
#   TTYPE_PREFIX    install prefix (default: /usr/local, or ~/.local if that
#                   is not writable)

set -eu

REPO="alirezaudev/ttype"
BINARY="ttype"

die() {
	echo "install: $*" >&2
	exit 1
}

need() {
	command -v "$1" >/dev/null 2>&1 || die "$1 is required"
}

need uname
need tar
need mktemp

if command -v curl >/dev/null 2>&1; then
	fetch() { curl -fsSL "$1" -o "$2"; }
	fetch_stdout() { curl -fsSL "$1"; }
elif command -v wget >/dev/null 2>&1; then
	fetch() { wget -qO "$2" "$1"; }
	fetch_stdout() { wget -qO- "$1"; }
else
	die "curl or wget is required"
fi

case "$(uname -s)" in
	Linux) os="linux" ;;
	Darwin) os="darwin" ;;
	*) die "unsupported operating system: $(uname -s)" ;;
esac

case "$(uname -m)" in
	x86_64 | amd64) arch="amd64" ;;
	arm64 | aarch64) arch="arm64" ;;
	*) die "unsupported architecture: $(uname -m)" ;;
esac

version="${TTYPE_VERSION:-}"
if [ -z "$version" ]; then
	version=$(fetch_stdout "https://api.github.com/repos/$REPO/releases/latest" |
		sed -n 's/.*"tag_name": *"v\{0,1\}\([^"]*\)".*/\1/p' | head -n 1)
	[ -n "$version" ] || die "could not resolve the latest release"
fi

prefix="${TTYPE_PREFIX:-}"
if [ -z "$prefix" ]; then
	if [ -w /usr/local/bin ] 2>/dev/null; then
		prefix="/usr/local"
	else
		prefix="$HOME/.local"
	fi
fi

archive="${BINARY}_${version}_${os}_${arch}.tar.gz"
base="https://github.com/$REPO/releases/download/v$version"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT INT TERM

echo "Downloading $BINARY $version ($os/$arch)..."
fetch "$base/$archive" "$tmp/$archive" || die "download failed: $base/$archive"

# Verify the checksum when a tool for it exists; skipping is better than failing
# on a minimal system, but a mismatch always stops the install.
if fetch "$base/checksums.txt" "$tmp/checksums.txt" 2>/dev/null; then
	if command -v sha256sum >/dev/null 2>&1; then
		sum=$(sha256sum "$tmp/$archive" | cut -d' ' -f1)
	elif command -v shasum >/dev/null 2>&1; then
		sum=$(shasum -a 256 "$tmp/$archive" | cut -d' ' -f1)
	else
		sum=""
	fi

	if [ -n "$sum" ]; then
		want=$(grep " \*\{0,1\}$archive\$" "$tmp/checksums.txt" | cut -d' ' -f1)
		[ -n "$want" ] || die "no checksum published for $archive"
		[ "$sum" = "$want" ] || die "checksum mismatch for $archive"
		echo "Checksum verified."
	fi
fi

tar -xzf "$tmp/$archive" -C "$tmp"

binary=$(find "$tmp" -type f -name "$BINARY" -perm -u+x | head -n 1)
[ -n "$binary" ] || die "the archive contains no $BINARY binary"

install_file() {
	src="$1"
	dest="$2"
	mode="$3"

	dir=$(dirname "$dest")
	if mkdir -p "$dir" 2>/dev/null && cp "$src" "$dest" 2>/dev/null; then
		chmod "$mode" "$dest"
		return 0
	fi

	command -v sudo >/dev/null 2>&1 || die "cannot write to $dir"
	echo "Writing to $dir needs sudo."
	sudo mkdir -p "$dir"
	sudo cp "$src" "$dest"
	sudo chmod "$mode" "$dest"
}

install_file "$binary" "$prefix/bin/$BINARY" 755

manpage=$(find "$tmp" -type f -name "$BINARY.1" | head -n 1)
if [ -n "$manpage" ]; then
	install_file "$manpage" "$prefix/share/man/man1/$BINARY.1" 644
fi

echo "Installed $BINARY $version to $prefix/bin/$BINARY"

case ":$PATH:" in
	*":$prefix/bin:"*) ;;
	*) echo "Note: $prefix/bin is not on your PATH." ;;
esac

#!/usr/bin/env sh
# This file written by AI :)
#
# Remove ttype and, if asked, the data it keeps.
#
#   sh uninstall.sh            remove the binary and man page
#   sh uninstall.sh --purge    also remove history, bests, replays and languages

set -eu

BINARY="ttype"
purge=0

for arg in "$@"; do
	case "$arg" in
		--purge) purge=1 ;;
		*) echo "uninstall: unknown option $arg" >&2; exit 1 ;;
	esac
done

remove() {
	target="$1"
	[ -e "$target" ] || return 0

	if rm -rf "$target" 2>/dev/null; then
		echo "  removed $target"
		return 0
	fi
	if command -v sudo >/dev/null 2>&1; then
		sudo rm -rf "$target"
		echo "  removed $target"
		return 0
	fi
	echo "  could not remove $target" >&2
}

for prefix in "/usr/local" "$HOME/.local" "/usr"; do
	remove "$prefix/bin/$BINARY"
	remove "$prefix/share/man/man1/$BINARY.1"
done

if [ "$purge" -eq 1 ]; then
	remove "${XDG_DATA_HOME:-$HOME/.local/share}/$BINARY"
	remove "${XDG_CONFIG_HOME:-$HOME/.config}/$BINARY"
	remove "${XDG_CACHE_HOME:-$HOME/.cache}/$BINARY"
else
	echo "Kept your history and settings. Pass --purge to remove them too."
fi

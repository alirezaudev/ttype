# This file written by AI :)
#
# Throwaway config and data dirs, sourced by every tape after _settings.tape.
# All of it is hidden from the recording.
#
# ttype reads and writes real settings, history and bests. Without this the GIF
# would show whatever the person recording happens to have saved, and every take
# would land in their history. The pinned config also marks the first run as
# done, so the welcome screen does not open in front of each demo.
#
# XDG_CONFIG_HOME and XDG_DATA_HOME are ignored on macOS, where ttype uses
# ~/Library instead. Record on Linux.

Hide
Set TypingSpeed 1ms
Type 'export PATH="$PWD/bin:$PATH"'
Enter
Type 'export XDG_CONFIG_HOME="$(mktemp -d)" XDG_DATA_HOME="$(mktemp -d)"'
Enter
Type 'mkdir -p "$XDG_CONFIG_HOME/ttype"'
Enter
Type `printf '%s' '{"onboarded":true}' > "$XDG_CONFIG_HOME/ttype/config.json"`
Enter
Type 'clear'
Enter
Set TypingSpeed 120ms
Show

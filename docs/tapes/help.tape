# This file written by AI :)
#
# Discoverability: the `?` overlay and the settings panel behind it.
#
# Both are built from the keymaps in internal/tui/keys.go, so this one goes
# stale the moment a binding changes. Re-record it whenever keys.go changes.

Output docs/help.gif

Source docs/tapes/_settings.tape
# Taller than the shared default on purpose. The overlay drops whole sections
# to fit a short terminal — during test, then results, then the flag examples —
# and at the usual 31 rows only the first section survives. This is the height
# where it stops apologising with "more in man ttype".
Set Height 940
Source docs/tapes/_isolate.tape

Set TypingSpeed 35ms
Type "ttype --words 5 --seed 42"
Enter
Sleep 1200ms

# ? opens help only before the first keystroke.
Type "?"
Sleep 5000ms

# S goes from the overlay straight into the settings panel.
Type "S"
Sleep 2000ms

Down
Sleep 500ms
Down
Sleep 500ms
Down
Sleep 500ms
Down
Sleep 800ms
Down
Sleep 800ms

# Left and right adjust the focused row.
Right
Sleep 1500ms

# Esc leaves without applying, back to the run it was opened from.
Escape
Sleep 1200ms
Ctrl+C
Sleep 500ms

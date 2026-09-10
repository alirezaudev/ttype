# This file written by AI :)
#
# Code modes, the thing a terminal typing test can do that a word list cannot.
# Three of them typed end to end, then the picker that lists all eight.
#
# Each seed pins its snippet. Re-derive with `ttype --mode <m> --words <n>
# --seed <s>` and copy what is on screen before touching any Type line.
#
#   go     seed 13  n=2  io.Copy(dst, src) atomic.AddInt64(&count, 1)
#   sql    seed  6  n=1  UPDATE accounts SET balance = balance - 100 WHERE id = 42;
#   shell  seed  9  n=2  ss -tulpn | grep LISTEN sha256sum dist/* > checksums.txt
#
# Space is a literal character in code modes — it does not commit a word — so
# these are typed exactly as they appear.

Output docs/modes.gif

Source docs/tapes/_settings.tape
Source docs/tapes/_isolate.tape

Set TypingSpeed 35ms
Type "ttype --mode go --words 2 --seed 13 --width 60"
Enter
Sleep 800ms
Set TypingSpeed 120ms
Type "io.Copy(dst, src) atomic.AddInt64(&count, 1)"
Sleep 1600ms
Ctrl+C
Sleep 400ms

Set TypingSpeed 35ms
Type "ttype --mode sql --words 1 --seed 6 --width 60"
Enter
Sleep 800ms
Set TypingSpeed 120ms
Type "UPDATE accounts SET balance = balance - 100 WHERE id = 42;"
Sleep 1600ms
Ctrl+C
Sleep 400ms

Set TypingSpeed 35ms
Type "ttype --mode shell --words 2 --seed 9 --width 60"
Enter
Sleep 800ms
Set TypingSpeed 120ms
Type "ss -tulpn | grep LISTEN sha256sum dist/* > checksums.txt"
Sleep 1600ms
Ctrl+C
Sleep 400ms

# The picker, reached through the help overlay. `?` only opens help before the
# first keystroke; after that it is typed input, because code targets contain
# question marks. help.tape gives the overlay a proper look, here it is only
# the doorway, so it does not linger.
Set TypingSpeed 35ms
Type "ttype --words 5 --seed 42"
Enter
Sleep 1000ms
Type "?"
Sleep 1300ms
# Uppercase M: the overlay and results shortcuts are uppercase so a keystroke
# still in flight when a run ends cannot trigger one.
Type "M"
Sleep 2500ms
Down
Sleep 400ms
Down
Sleep 400ms
Down
Sleep 1200ms
Escape
Sleep 600ms
Ctrl+C
Sleep 500ms

# This file written by AI :)
#
# Replay: ttype records every keystroke and plays it back at the original
# timing. Nothing else in a terminal typing test does this, so it gets its own
# recording.
#
# The hidden block runs three real tests to fill history — there is no fixture
# to fake, because the replay is the recording. Each one types at a different
# speed so the table shows three different scores instead of the same number
# three times (VHS types at a fixed interval, so the interval is the wpm).
#
#   seed 101  n=8  dynamics sexual green movie machines surprise advisory notify
#   seed 102  n=8  correct optimization settings until processor heaven mark sensitive
#   seed 103  n=8  reality reg directly gb disorders zealand too britain

Output docs/replay.gif

Source docs/tapes/_settings.tape
Source docs/tapes/_isolate.tape

Hide
Set TypingSpeed 35ms
Type "ttype --words 8 --seed 101 --width 58"
Enter
Sleep 800ms
Set TypingSpeed 150ms
Type "dynamics sexual green movie machines surprise advisory notify"
Sleep 900ms
Ctrl+C
Sleep 400ms

Set TypingSpeed 35ms
Type "ttype --words 8 --seed 102 --width 58"
Enter
Sleep 800ms
Set TypingSpeed 130ms
Type "correct optimization settings until processor heaven mark sensitive"
Sleep 900ms
Ctrl+C
Sleep 400ms

Set TypingSpeed 35ms
Type "ttype --words 8 --seed 103 --width 58"
Enter
Sleep 800ms
Set TypingSpeed 115ms
Type "reality reg directly gb disorders zealand too britain"
Sleep 900ms
Ctrl+C
Sleep 400ms

Type "clear"
Enter
Show

Set TypingSpeed 35ms
Type "ttype history"
Enter
Sleep 2200ms

Down
Sleep 700ms
Down
Sleep 1200ms

# Enter plays the selected run back keystroke by keystroke, at the speed it was
# originally typed.
Enter
Sleep 5000ms

# 2 doubles the speed, space pauses, r starts over.
Type "2"
Sleep 3500ms

Escape
Sleep 1500ms
Type "q"
Sleep 600ms

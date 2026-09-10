# This file written by AI :)
#
# The three modifiers: blind, zen and the min-wpm threshold.
#
#   blind   seed 103  n=8   reality reg directly gb disorders zealand too britain
#   zen     seed 106  n=8   customers bowl organization ct till catering popular home
#   min-wpm seed  35  n=12  file farmers evolution sections so account ...
#
# The min-wpm run types three words and then stops. The threshold is only
# checked once five seconds have passed, and the check runs on the tick, so
# letting it fire while idle keeps stray keystrokes off the results screen.

Output docs/gameplay.gif

Source docs/tapes/_settings.tape
Source docs/tapes/_isolate.tape

# Blind: characters stay dots until the cursor is past the word, and the live
# line drops to raw speed. Net wpm and accuracy would leak the mistakes blind
# mode exists to hide.
Set TypingSpeed 35ms
Type "ttype --blind --words 8 --seed 103 --width 58"
Enter
Sleep 1000ms
Set TypingSpeed 120ms
Type "reality reg directly gb disorders zealand too britain"
Sleep 2400ms
Ctrl+C
Sleep 400ms

# Zen: no HUD at all.
Set TypingSpeed 35ms
Type "ttype --zen --words 8 --seed 106 --width 58"
Enter
Sleep 1000ms
Set TypingSpeed 120ms
Type "customers bowl organization ct till catering popular home"
Sleep 2400ms
Ctrl+C
Sleep 400ms

# Min-wpm: type, hesitate, fail.
Set TypingSpeed 35ms
Type "ttype --min-wpm 60 --words 12 --seed 35 --width 58"
Enter
Sleep 1000ms
Set TypingSpeed 120ms
Type "file farmers evolution "
Sleep 4500ms
Sleep 2500ms
Ctrl+C
Sleep 500ms

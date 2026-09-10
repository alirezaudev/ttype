# This file written by AI :)
#
# The README hero: one word test, one real typo, the results screen.
#
# --seed 35 pins the words, and every Type line below is that text verbatim.
# The seed and the typed text are one unit — change either without the other
# and the demo records a screen full of red. To re-derive:
#
#   ttype --words 12 --seed 35 --width 62
#
# and copy what is on screen. Seed 35 gives:
#
#   file farmers evolution sections so account ringtones design resume semi
#   additional spent
#
# --words rather than --time so the run ends on the last keystroke instead of
# on a clock the tape would have to stay in sync with.

Output docs/demo.gif

Source docs/tapes/_settings.tape
Source docs/tapes/_isolate.tape

Set TypingSpeed 35ms
Type "ttype --words 12 --seed 35 --width 62"
Enter
Sleep 1200ms
Set TypingSpeed 120ms

Type "file farmers evolution sections so account ringtones "

# A mistake, caught and corrected. This is what puts a red cell in the results
# chart, a row in the character heatmap, and accuracy under 100%.
Type "desi"
Sleep 150ms
Type "n"
Sleep 500ms
Backspace
Sleep 250ms
Type "gn resume semi additional spent"

# The last keystroke ends the run. Results ignore input for 700ms so a
# keystroke still in flight cannot trigger a shortcut; let that settle.
Sleep 4500ms
Ctrl+C
Sleep 500ms

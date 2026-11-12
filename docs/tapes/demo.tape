# This file written by AI :)
#
# The README hero: one sentence, one real typo, the results screen.
#
# --text pins what is typed, and the Type lines below are that text verbatim.
# The flag and the typed text are one unit: change either without the other
# and the demo records a screen full of red.
#
# --seed only keeps the seed on the results screen short, and --width keeps the
# test narrow enough for the recording to zoom in on it while it is typed.

Output docs/demo.svg

Set TypingSpeed 35ms
Type `ttype --seed 7 --width 50 --text "it doesn't matter where I am, it's all in my mind"`
Enter
Sleep 1200ms
Set TypingSpeed 120ms

Type "it doesn't matter where I am, it's all in my "

# A mistake, caught and corrected. This is what puts a red cell in the results
# chart, a row in the character heatmap, and accuracy under 100%.
Type "mi"
Sleep 150ms
Type "m"
Sleep 500ms
Backspace
Sleep 250ms
Type "nd"

# The last keystroke ends the run. Results ignore input for 700ms so a
# keystroke still in flight cannot trigger a shortcut; let that settle.
Sleep 4500ms
Ctrl+C
Sleep 500ms

# This file written by AI :)
#
# The three built-in themes, each on the two screens that matter: the coloured
# typing area and the results hero.
#
#   default  seed 35  file farmers evolution sections so account
#   monokai  seed  7  maximum iraq building watch rhode villa
#   dracula  seed 12  posters donald hear credits cst contest
#
# A different seed per theme on purpose. The same six words three times reads
# as a stutter rather than a comparison.

Output docs/themes.gif

Source docs/tapes/_settings.tape
Source docs/tapes/_isolate.tape

Set TypingSpeed 35ms
Type "ttype --theme default --words 6 --seed 35 --width 56"
Enter
Sleep 900ms
Set TypingSpeed 120ms
Type "file farmers evolution sections so account"
Sleep 2200ms
Ctrl+C
Sleep 400ms

Set TypingSpeed 35ms
Type "ttype --theme monokai --words 6 --seed 7 --width 56"
Enter
Sleep 900ms
Set TypingSpeed 120ms
Type "maximum iraq building watch rhode villa"
Sleep 2200ms
Ctrl+C
Sleep 400ms

Set TypingSpeed 35ms
Type "ttype --theme dracula --words 6 --seed 12 --width 56"
Enter
Sleep 900ms
Set TypingSpeed 120ms
Type "posters donald hear credits cst contest"
Sleep 2600ms
Ctrl+C
Sleep 500ms

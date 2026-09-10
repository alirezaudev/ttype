# This file written by AI :)
#
# Shared VHS settings. Sourced first by every tape, never recorded on its own.
#
# It comes before _isolate.tape so a tape can override the geometry in between:
#
#   Source docs/tapes/_settings.tape
#   Set Height 800                      # this one needs more rows
#   Source docs/tapes/_isolate.tape
#
# 1140x640 at font size 16 is roughly 102x32 cells. Wide enough for the results
# screen to keep its side-by-side layout and print the whole shortcut line
# without dropping a binding, short enough that the GIF is still reasonable to
# put in a README.
#
# TypingSpeed is not cosmetic. VHS types at a fixed interval, and that interval
# is the WPM the recording shows:
#
#   wpm = (60000 / interval_ms) / 5      ->  120ms is about 100 wpm
#
# Under ~60ms it reads as broken rather than fast. Shell commands are typed at
# 35ms because they are stage directions, not the demo.

Set Shell bash
Set FontSize 16
Set Width 1140
Set Height 640
Set Padding 20
Set Theme "Catppuccin Mocha"
Set TypingSpeed 120ms

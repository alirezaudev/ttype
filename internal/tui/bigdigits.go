package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// A 3-row box-drawing font for the two hero numbers on the result screen.
// Every glyph is 3 cells wide so multi-digit values line up.
var bigGlyphs = map[rune][3]string{
	'0': {"┌─┐", "│ │", "└─┘"},
	'1': {" ┐ ", " │ ", " ┴ "},
	'2': {"╶─┐", "┌─┘", "└─╴"},
	'3': {"╶─┐", " ─┤", "╶─┘"},
	'4': {"╷ ╷", "└─┤", "  ╵"},
	'5': {"┌─╴", "└─┐", "╶─┘"},
	'6': {"┌─╴", "├─┐", "└─┘"},
	'7': {"╶─┐", "  │", "  ╵"},
	'8': {"┌─┐", "├─┤", "└─┘"},
	'9': {"┌─┐", "└─┤", "╶─┘"},
	'%': {"▘ ╱", " ╱ ", "╱ ▗"},
}

// Anything outside the font falls back to itself, centered in a 3-cell column.
func renderBigDigits(text string, style lipgloss.Style) string {
	var rows [3]strings.Builder
	for i, r := range text {
		glyph, ok := bigGlyphs[r]
		if !ok {
			glyph = [3]string{"   ", " " + string(r) + " ", "   "}
		}
		for row := range rows {
			if i > 0 {
				rows[row].WriteByte(' ')
			}
			rows[row].WriteString(glyph[row])
		}
	}

	lines := make([]string, len(rows))
	for i := range rows {
		lines[i] = style.Render(rows[i].String())
	}
	return strings.Join(lines, "\n")
}

package tui

import (
	"fmt"
	"time"

	"github.com/mattn/go-runewidth"
)

type lineSpan struct {
	start int
	end   int
}

func wordWrapIndices(text []rune, width int) []lineSpan {
	if len(text) == 0 {
		return nil
	}

	if width < 1 {
		width = 1
	}

	var lines []lineSpan
	start := 0
	for start < len(text) {
		// Chinese, Japanese and Korean take two cells each, so the line ends
		// where the cells run out, not after width characters.
		end, cells := start, 0
		for end < len(text) && cells+runewidth.RuneWidth(text[end]) <= width {
			cells += runewidth.RuneWidth(text[end])
			end++
		}
		if end == len(text) {
			// Last line, or the only line if the text fits on a single line.
			lines = append(lines, lineSpan{start: start, end: len(text)})
			break
		}
		if end == start {
			end++ // A character wider than the whole line still has to go somewhere.
		}
		breakAt := end
		space := false
		for i := end; i > start; i-- {
			if text[i-1] == ' ' {
				breakAt = i // Break at the last space so the next word stays together.
				space = true
				break
			}
		}

		if !space {
			breakAt = end // No space found, so hard-break at the width.
		}

		lines = append(lines, lineSpan{start: start, end: breakAt})
		start = breakAt

		for start < len(text) && text[start] == ' ' {
			start++ // Skip leading spaces on the next line.
		}
	}

	return lines
}

func cursorLine(lines []lineSpan, cursor int) int {
	if len(lines) == 0 {
		return 0
	}

	for i, line := range lines {
		if cursor >= line.start && cursor < line.end {
			return i
		}
		if cursor == line.end && i == len(lines)-1 {
			return i
		}
	}

	return len(lines) - 1
}

func visibleLineWindow(lines []lineSpan, cursor, maxLines int) (from, to int) {
	if maxLines < 1 {
		maxLines = 1
	}

	if len(lines) == 0 {
		return 0, 0
	}

	current := cursorLine(lines, cursor)
	from = current - maxLines/2
	if from < 0 {
		from = 0
	}

	to = from + maxLines
	if to > len(lines) {
		to = len(lines)
		from = to - maxLines
		if from < 0 {
			from = 0
		}
	}

	return from, to
}

func formatClock(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	secs := int(d.Round(time.Second).Seconds())
	return fmt.Sprintf("%d:%02d", secs/60, secs%60)
}

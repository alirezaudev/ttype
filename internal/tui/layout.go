package tui

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

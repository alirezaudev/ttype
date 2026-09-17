package tui

import (
	"testing"

	"github.com/mattn/go-runewidth"
)

func TestCursorLine(t *testing.T) {
	t.Parallel()

	lines := []lineSpan{{start: 0, end: 5}, {start: 5, end: 10}}

	tests := []struct {
		name   string
		lines  []lineSpan
		cursor int
		want   int
	}{
		{name: "no lines", lines: nil, cursor: 5, want: 0},
		{name: "start of first line", lines: lines, cursor: 0, want: 0},
		{name: "inside first line", lines: lines, cursor: 3, want: 0},
		{name: "boundary belongs to next line", lines: lines, cursor: 5, want: 1},
		{name: "inside last line", lines: lines, cursor: 7, want: 1},
		{name: "end of last line", lines: lines, cursor: 10, want: 1},
		{name: "past the end", lines: lines, cursor: 100, want: 1},
		{name: "before the start", lines: lines, cursor: -1, want: 1},
	}

	for _, test := range tests {
		if got := cursorLine(test.lines, test.cursor); got != test.want {
			t.Errorf("%s: cursorLine(cursor=%d) = %d, want %d", test.name, test.cursor, got, test.want)
		}
	}
}

func TestVisibleLineWindow(t *testing.T) {
	t.Parallel()

	five := []lineSpan{{start: 0, end: 5}, {start: 5, end: 10}, {start: 10, end: 15}, {start: 15, end: 20}, {start: 20, end: 25}}
	three := []lineSpan{{start: 0, end: 5}, {start: 5, end: 10}, {start: 10, end: 15}}

	tests := []struct {
		name     string
		lines    []lineSpan
		cursor   int
		maxLines int
		wantFrom int
		wantTo   int
	}{
		{name: "no lines", lines: nil, cursor: 0, maxLines: 10, wantFrom: 0, wantTo: 0},
		{name: "fewer lines than window", lines: three, cursor: 0, maxLines: 10, wantFrom: 0, wantTo: 3},
		{name: "maxLines below one becomes one", lines: three, cursor: 0, maxLines: 0, wantFrom: 0, wantTo: 1},
		{name: "cursor near top", lines: five, cursor: 0, maxLines: 3, wantFrom: 0, wantTo: 3},
		{name: "cursor in the middle", lines: five, cursor: 12, maxLines: 3, wantFrom: 1, wantTo: 4},
		{name: "cursor near bottom", lines: five, cursor: 24, maxLines: 3, wantFrom: 2, wantTo: 5},
		{name: "window wider than lines", lines: three, cursor: 14, maxLines: 10, wantFrom: 0, wantTo: 3},
	}

	for _, test := range tests {
		from, to := visibleLineWindow(test.lines, test.cursor, test.maxLines)
		if from != test.wantFrom || to != test.wantTo {
			t.Errorf("%s: visibleLineWindow(cursor=%d, maxLines=%d) = (%d, %d), want (%d, %d)",
				test.name, test.cursor, test.maxLines, from, to, test.wantFrom, test.wantTo)
		}
	}
}

func TestWrapCountsWideCharactersTwice(t *testing.T) {
	t.Parallel()

	text := []rune("你好 世界 我们 学习 中文 今天")
	for _, line := range wordWrapIndices(text, 10) {
		if w := runewidth.StringWidth(string(text[line.start:line.end])); w > 10 {
			t.Fatalf("line %q is %d cells wide, want at most 10", string(text[line.start:line.end]), w)
		}
	}
}

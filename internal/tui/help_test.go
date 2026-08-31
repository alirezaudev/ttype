package tui

import (
	"strings"
	"testing"
)

func TestHelpOverlayFitsEveryTerminal(t *testing.T) {
	t.Parallel()

	for _, size := range [][2]int{{100, 40}, {80, 24}, {60, 18}, {40, 12}, {30, 8}, {20, 5}} {
		help := NewHelpOverlay(defaultTheme())
		help.setSize(size[0], size[1])
		view := help.View()

		if got := strings.Count(view, "\n") + 1; got > size[1] {
			t.Errorf("%dx%d: overlay is %d rows tall", size[0], size[1], got)
		}
		for _, line := range strings.Split(view, "\n") {
			if got := len([]rune(stripANSI(line))); got > size[0] {
				t.Errorf("%dx%d: line is %d columns wide: %q", size[0], size[1], got, stripANSI(line))
				break
			}
		}
	}
}

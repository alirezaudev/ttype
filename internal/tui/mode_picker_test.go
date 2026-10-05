package tui

import (
	"strings"
	"testing"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/mattn/go-runewidth"
)

func TestModePickerRowsLineUp(t *testing.T) {
	t.Parallel()

	for _, current := range domain.AllTextModes() {
		p := NewModePicker(current, defaultTheme())
		p.setSize(80, 24)
		lines := strings.Split(stripANSI(p.View()), "\n")

		col := -1
		for _, line := range lines {
			trimmed := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), ">"))
			if !isMode(trimmed) {
				continue
			}
			start := runewidth.StringWidth(line[:strings.Index(line, trimmed)])
			if col == -1 {
				col = start
			} else if start != col {
				t.Fatalf("selected %s: names start at different columns:\n%s", current, strings.Join(lines, "\n"))
			}
		}
		if col == -1 {
			t.Fatalf("no mode rows in:\n%s", strings.Join(lines, "\n"))
		}
	}
}

func isMode(s string) bool {
	for _, mode := range domain.AllTextModes() {
		if s == string(mode) {
			return true
		}
	}
	return false
}

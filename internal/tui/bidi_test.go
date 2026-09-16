package tui

import (
	"strings"
	"testing"

	"github.com/alirezaudev/ttype/internal/domain"
)

func TestShouldReorderRTL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		env  map[string]string
		want bool
	}{
		{"plain terminal", nil, true},
		{"jetbrains", map[string]string{"TERMINAL_EMULATOR": "JetBrains-JediTerm"}, true},
		{"gnome terminal", map[string]string{"VTE_VERSION": "7600"}, false},
		{"konsole", map[string]string{"KONSOLE_VERSION": "230804"}, false},
		{"forced on", map[string]string{"VTE_VERSION": "7600", "TTYPE_BIDI": "ttype"}, true},
		{"forced off", map[string]string{"TTYPE_BIDI": "terminal"}, false},
	}

	for _, test := range tests {
		getenv := func(name string) string { return test.env[name] }
		if got := shouldReorderRTL(getenv); got != test.want {
			t.Errorf("%s: shouldReorderRTL = %v, want %v", test.name, got, test.want)
		}
	}
}

func visualString(text string) string {
	runes := []rune(text)
	var b strings.Builder
	for _, i := range visualOrder(runes, 0, len(runes)) {
		b.WriteRune(mirrorRune(runes[i]))
	}
	return b.String()
}

func TestVisualOrderReversesOnlyTheRightToLeftText(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"سلام دنیا":      "ایند مالس",
		"سلام 2026 دنیا": "ایند 2026 مالس",
		"(שלום)":         "(םולש)",
		"سلام go دنیا":   "ایند go مالس",
	}
	for text, want := range tests {
		if got := visualString(text); got != want {
			t.Errorf("visualOrder(%q) = %q, want %q", text, got, want)
		}
	}
}

func TestRightToLeft(t *testing.T) {
	t.Parallel()

	for text, want := range map[string]bool{
		"سلام hello": true,
		"hello سلام": false,
		"2026 שלום":  true,
		"123 ...":    false,
	} {
		if got := rightToLeft([]rune(text)); got != want {
			t.Errorf("rightToLeft(%q) = %v, want %v", text, got, want)
		}
	}
}

func TestPersianIsDrawnRightToLeft(t *testing.T) {
	t.Parallel()

	cfg := domain.TestConfig{Kind: domain.TestKindTimed, Duration: domain.Duration60}
	m := newTestModelFor(t, "سلام دنیا", cfg)
	m.setSize(40, 20)

	m.reorderRTL = true
	lines := strings.Split(stripANSI(m.renderWords()), "\n")
	if !strings.HasSuffix(lines[0], visualString("سلام دنیا")) {
		t.Fatalf("line = %q, want the words in display order, right-aligned", lines[0])
	}

	m.reorderRTL = false
	if got := stripANSI(m.renderWords()); got != "سلام دنیا" {
		t.Fatalf("with the terminal reordering, line = %q, want the text as is", got)
	}
}

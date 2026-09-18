package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/engine"
)

func newHUDSession(t *testing.T, cfg domain.TestConfig) (*engine.Session, *engine.FakeClock) {
	t.Helper()

	clock := engine.NewFakeClock(time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC))
	s, err := engine.NewSession(cfg, fixedSource("hello world example"), clock)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	return s, clock
}

func TestHUDLiveLineKeepsFixedWidth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		wpm, raw, acc float64
		errs          int
	}{
		{5, 7, 100, 0},
		{55, 77, 98, 9},
		{555, 777, 9, 99},
		{9999, 9999, 100, 9999},
	}

	want := len(hudLiveLine(0, 0, 0, 0, false))
	for _, test := range tests {
		line := hudLiveLine(test.wpm, test.raw, test.acc, test.errs, false)
		if len(line) != want {
			t.Errorf("hudLiveLine(%v) = %q, width %d, want %d", test, line, len(line), want)
		}
	}
}

// The first keystroke lands a few milliseconds into the run; reading the pace
// straight off that elapsed time used to peg the line at 999.
func TestHUDPaceHoldsThroughTheFirstSecond(t *testing.T) {
	t.Parallel()

	cfg := domain.TestConfig{Kind: domain.TestKindTimed, Duration: domain.Duration60}
	s, clock := newHUDSession(t, cfg)

	s.InputRune('h')
	clock.Advance(40 * time.Millisecond)

	hud := stripANSI(renderHUD(s, defaultTheme(), 76, cfg, false))
	if !strings.Contains(hud, "wpm 12 ") || !strings.Contains(hud, "raw 12 ") {
		t.Fatalf("HUD = %q, want one character rated over a whole second", hud)
	}
}

func TestHUDTimerFollowsSessionState(t *testing.T) {
	t.Parallel()

	cfg := domain.TestConfig{Kind: domain.TestKindTimed, Duration: domain.Duration60}
	s, clock := newHUDSession(t, cfg)

	ready := stripANSI(renderHUD(s, defaultTheme(), 76, cfg, false))
	if !strings.Contains(ready, "1:00") {
		t.Fatalf("ready HUD = %q, want the full duration", ready)
	}

	s.InputRune('h')
	clock.Advance(5 * time.Second)

	active := stripANSI(renderHUD(s, defaultTheme(), 76, cfg, false))
	if !strings.Contains(active, "0:55") {
		t.Fatalf("active HUD = %q, want a countdown", active)
	}
	if len(ready) != len(active) {
		t.Fatalf("HUD width changed on the first keystroke: %d → %d", len(ready), len(active))
	}
}

func TestHUDErrCountsCorrectedMistakes(t *testing.T) {
	t.Parallel()

	cfg := domain.TestConfig{Kind: domain.TestKindTimed, Duration: domain.Duration60}
	s, _ := newHUDSession(t, cfg)

	s.InputRune('x')
	s.Backspace()
	s.InputRune('h')

	hud := stripANSI(renderHUD(s, defaultTheme(), 76, cfg, false))
	if !strings.Contains(hud, "err 1") {
		t.Fatalf("HUD = %q, want err 1 (a fixed mistake still counts)", hud)
	}
}

func TestHUDWordsModeShowsProgress(t *testing.T) {
	t.Parallel()

	cfg := domain.TestConfig{Kind: domain.TestKindWords, WordCount: 50}
	s, _ := newHUDSession(t, cfg)

	hud := stripANSI(renderHUD(s, defaultTheme(), 76, cfg, false))
	if !strings.Contains(hud, " 0/50 · 0:00") {
		t.Fatalf("words HUD = %q, want padded progress and a clock", hud)
	}
}

func TestHUDNarrowWidthDropsLiveStats(t *testing.T) {
	t.Parallel()

	cfg := domain.TestConfig{Kind: domain.TestKindTimed, Duration: domain.Duration60}
	s, _ := newHUDSession(t, cfg)

	hud := stripANSI(renderHUD(s, defaultTheme(), 20, cfg, false))
	if strings.Contains(hud, "wpm") {
		t.Fatalf("HUD = %q, want the live stats dropped at 20 columns", hud)
	}
	if strings.Contains(hud, "\n") {
		t.Fatalf("HUD = %q, want one line, not a wrap", hud)
	}
	if !strings.Contains(hud, "1:00") {
		t.Fatalf("HUD = %q, want the timer kept", hud)
	}
}

func TestBlindHUDShowsRawOnly(t *testing.T) {
	t.Parallel()

	line := hudLiveLine(80, 90, 97, 4, true)
	if strings.Contains(line, "wpm") || strings.Contains(line, "acc") || strings.Contains(line, "err") {
		t.Fatalf("blind live line = %q, want raw only", line)
	}
	if !strings.Contains(line, "raw") {
		t.Fatalf("blind live line = %q, want the raw speed", line)
	}
}

func TestZenHUDIsEmpty(t *testing.T) {
	t.Parallel()

	cfg := domain.TestConfig{Kind: domain.TestKindTimed, Duration: domain.Duration60, Zen: true}
	s, _ := newHUDSession(t, cfg)

	if got := renderHUD(s, defaultTheme(), 76, cfg, false); got != "" {
		t.Fatalf("zen HUD = %q, want empty", got)
	}
}

// Batching must not change a single visible character.
func TestBatchedRenderMatchesTheCharacters(t *testing.T) {
	t.Parallel()

	const target = "the quick brown fox jumps over the lazy dog"
	cfg := domain.TestConfig{Kind: domain.TestKindTimed, Duration: domain.Duration60}
	m := newTestModelFor(t, target, cfg)
	for _, r := range "the qwick brown " {
		m.session.InputRune(r)
	}

	got := stripANSI(m.renderWords())
	want := strings.Join(visibleTargetLines(m), "\n")
	if got != want {
		t.Fatalf("rendered %q, want %q", got, want)
	}
}

func visibleTargetLines(m TestModel) []string {
	target := m.session.TargetRunes()
	lines := wordWrapIndices(target, m.typingWidth())
	from, to := visibleLineWindow(lines, m.session.Cursor(), 3)

	out := make([]string, 0, to-from)
	for i := from; i < to; i++ {
		out = append(out, string(target[lines[i].start:lines[i].end]))
	}
	return out
}

func TestWrapCacheRecomputesOnResize(t *testing.T) {
	t.Parallel()

	target := []rune("the quick brown fox jumps over the lazy dog")
	cache := &wrapCache{}

	wide := cache.wrap(target, string(target), 60, 0, nil)
	if len(cache.wrap(target, string(target), 60, 0, nil)) != len(wide) {
		t.Fatal("a repeat wrap at the same width should reuse the cache")
	}

	narrow := cache.wrap(target, string(target), 20, 0, nil)
	if len(narrow) <= len(wide) {
		t.Fatalf("narrow wrap = %d lines, wide = %d; expected more lines", len(narrow), len(wide))
	}

	other := []rune("a different target entirely")
	if got := cache.wrap(other, string(other), 20, 0, nil); string(other[got[0].start:got[0].end]) == "" {
		t.Fatal("a new target should be re-wrapped")
	}
}

func TestExtraLettersAreDrawnAfterTheWord(t *testing.T) {
	t.Parallel()

	cfg := domain.TestConfig{Kind: domain.TestKindTimed, Duration: domain.Duration60}
	m := newTestModelFor(t, "cat dog", cfg)
	for _, r := range "catxx" {
		m.session.InputRune(r)
	}
	if got := stripANSI(m.renderWords()); got != "catxx dog" {
		t.Fatalf("rendered %q, want %q", got, "catxx dog")
	}

	m.cfg.Blind = true
	if got := stripANSI(m.renderWords()); got != "······dog" {
		t.Fatalf("blind rendered %q, want %q", got, "······dog")
	}
}

func TestExtraLettersWrapTheWord(t *testing.T) {
	t.Parallel()

	text := []rune("aaaa bbbb cccc")
	extras := map[int]int{9: 3}
	lines := wrapWithExtras(text, 10, func(i int) int { return extras[i] })

	if got := string(text[lines[0].start:lines[0].end]); got != "aaaa " {
		t.Fatalf("first line = %q, want %q", got, "aaaa ")
	}
}

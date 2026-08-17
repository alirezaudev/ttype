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

func TestHUDTimerFollowsSessionState(t *testing.T) {
	t.Parallel()

	cfg := domain.TestConfig{Kind: domain.TestKindTimed, Duration: domain.Duration60}
	s, clock := newHUDSession(t, cfg)

	ready := stripANSI(renderHUD(s, defaultTheme(), 76, cfg))
	if !strings.Contains(ready, "1:00") {
		t.Fatalf("ready HUD = %q, want the full duration", ready)
	}

	s.InputRune('h')
	clock.Advance(5 * time.Second)

	active := stripANSI(renderHUD(s, defaultTheme(), 76, cfg))
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

	hud := stripANSI(renderHUD(s, defaultTheme(), 76, cfg))
	if !strings.Contains(hud, "err 1") {
		t.Fatalf("HUD = %q, want err 1 (a fixed mistake still counts)", hud)
	}
}

func TestHUDWordsModeShowsProgress(t *testing.T) {
	t.Parallel()

	cfg := domain.TestConfig{Kind: domain.TestKindWords, WordCount: 50}
	s, _ := newHUDSession(t, cfg)

	hud := stripANSI(renderHUD(s, defaultTheme(), 76, cfg))
	if !strings.Contains(hud, " 0/50 · 0:00") {
		t.Fatalf("words HUD = %q, want padded progress and a clock", hud)
	}
}

func TestHUDNarrowWidthDropsLiveStats(t *testing.T) {
	t.Parallel()

	cfg := domain.TestConfig{Kind: domain.TestKindTimed, Duration: domain.Duration60}
	s, _ := newHUDSession(t, cfg)

	hud := stripANSI(renderHUD(s, defaultTheme(), 20, cfg))
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

	if got := renderHUD(s, defaultTheme(), 76, cfg); got != "" {
		t.Fatalf("zen HUD = %q, want empty", got)
	}
}

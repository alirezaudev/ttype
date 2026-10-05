package engine_test

import (
	"testing"
	"time"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/engine"
)

func newCodeSession(t *testing.T, target string, mode domain.TextMode) *engine.Session {
	t.Helper()

	clock := engine.NewFakeClock(time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC))
	cfg := domain.TestConfig{
		Kind:     domain.TestKindTimed,
		Duration: domain.Duration(60),
		TextMode: mode,
	}
	s, err := engine.NewSession(cfg, fixedSource(target), clock)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	return s
}

func TestCodeModeMissedLetterStaysInItsWord(t *testing.T) {
	t.Parallel()

	s := newCodeSession(t, "SELECT id, name FROM users", domain.TextModeSQL)
	typeString(s, "SELCT id, name FROM users")

	for _, w := range s.Words() {
		if w.Missed != (w.Expected == "SELECT") {
			t.Fatalf("word %q missed = %v", w.Expected, w.Missed)
		}
	}
}

func TestCodeModeIndentationIsTyped(t *testing.T) {
	t.Parallel()

	s := newCodeSession(t, "if ok:     return x", domain.TextModePython)
	typeString(s, "if ok: z")

	if got := s.Cursor(); got != 7 {
		t.Fatalf("cursor = %d, want 7", got)
	}

	typeString(s, "    return x")
	if got := string(s.Input()); got != s.Target() {
		t.Fatalf("input = %q, want %q", got, s.Target())
	}
	if _, incorrect := s.Keystrokes(); incorrect != 1 {
		t.Fatalf("incorrect keystrokes = %d, want 1", incorrect)
	}
}

// Backend reads like prose but is a code mode.
func TestModeIsCode(t *testing.T) {
	t.Parallel()

	cases := map[domain.TextMode]bool{
		domain.TextModeWords:     false,
		domain.TextModeSentences: false,
		domain.TextModeCustom:    false,
		domain.TextModeSQL:       true,
		domain.TextModeGo:        true,
		domain.TextModeBackend:   true,
		domain.TextModePython:    true,
		domain.TextModeShell:     true,
		domain.TextModeRegex:     true,
		"":                       false,
	}

	for mode, want := range cases {
		if got := mode.IsCode(); got != want {
			t.Fatalf("%q is code = %v, want %v", mode, got, want)
		}
	}
}

func TestCodeModeCapsThresholdIsLonger(t *testing.T) {
	t.Parallel()

	s := newCodeSession(t, "abc", domain.TextModeSQL)
	typeString(s, "AB")

	if s.CapsLockSuspected() {
		t.Fatal("two inversions should not trip the warning in a code mode")
	}

	typeString(s, "C")
	if !s.CapsLockSuspected() {
		t.Fatal("three inversions should trip it")
	}
}

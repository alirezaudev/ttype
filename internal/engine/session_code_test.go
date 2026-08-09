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

// Indentation and runs of spaces are real content in code, so Space has to be
// an ordinary character there rather than a word commit.
func TestCodeModeSpaceIsLiteral(t *testing.T) {
	t.Parallel()

	s := newCodeSession(t, "select id", domain.TextModeSQL)
	typeString(s, "s ")

	if got := s.Cursor(); got != 2 {
		t.Fatalf("cursor = %d, want 2 (space must consume one position, not commit a word)", got)
	}
	if got := string(s.Input()); got != "s " {
		t.Fatalf("input = %q, want %q — no skip sentinel belongs in code modes", got, "s ")
	}
}

// The sticky separator only exists to stop an overrun cascading between words.
// With no word commit there is nothing to protect, so the extra letter is
// buffered as an ordinary error instead of being pinned out.
func TestCodeModeDoesNotPinExtraCharsAtASpace(t *testing.T) {
	t.Parallel()

	s := newCodeSession(t, "a b", domain.TextModeSQL)
	typeString(s, "ax")

	if got := s.Cursor(); got != 2 {
		t.Fatalf("cursor = %d, want 2", got)
	}
	if got := string(s.Input()); got != "ax" {
		t.Fatalf("input = %q, want %q", got, "ax")
	}
}

func TestModeSpaceCommitClassification(t *testing.T) {
	t.Parallel()

	cases := map[domain.TextMode]bool{
		domain.TextModeWords:     true,
		domain.TextModeSentences: true,
		domain.TextModeSQL:       false,
		"":                       true,
	}

	for mode, want := range cases {
		if got := mode.CommitsWordsOnSpace(); got != want {
			t.Fatalf("%q commits on space = %v, want %v", mode, got, want)
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

package engine_test

import (
	"strings"
	"testing"
	"time"

	"github.com/alirezaudev/ttype/internal/engine"
)

func newTestSession(t *testing.T, target string, duration time.Duration) (*engine.Session, *engine.FakeClock) {
	t.Helper()

	clock := engine.NewFakeClock(time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC))
	s := engine.NewSession(target, duration, clock)
	return s, clock
}

func TestSessionStartsOnFirstKeypress(t *testing.T) {
	t.Parallel()

	wantedDuration := 15 * time.Second
	s, clock := newTestSession(t, "abc", wantedDuration)

	if s.Remaining() != wantedDuration {
		t.Fatalf("remaining = %v, want %v", s.Remaining(), wantedDuration)
	}

	if s.Cursor() != 0 {
		t.Fatalf("cursor = %d, want 0", s.Cursor())
	}

	if s.Input() != nil {
		t.Fatalf("input = %v, want nil", s.Input())
	}

	s.InputRune('a')
	if s.Cursor() != 1 {
		t.Fatalf("cursor = %d, want 1", s.Cursor())
	}

	clock.Advance(2 * time.Second)
	if s.Remaining() != 13*time.Second {
		t.Fatalf("elapsed = %v, want 13s", s.Remaining())
	}

}

func TestSessionTickBeforeStartDoesNotFinish(t *testing.T) {
	t.Parallel()

	s, clock := newTestSession(t, "abc", 15*time.Second)

	clock.Advance(16 * time.Second)

	if s.Tick() {
		t.Fatal("Tick should not finish a session that has not started")
	}

	if s.Finished() {
		t.Fatal("session should not be finished before the first keystroke")
	}

	if got := s.Remaining(); got != 15*time.Second {
		t.Fatalf("Remaining() = %v, want 15s", got)
	}
}

func TestSessionTimerExpiry(t *testing.T) {
	t.Parallel()

	s, clock := newTestSession(t, strings.Repeat("a", 200), 15*time.Second)
	s.InputRune('a')
	clock.Advance(16 * time.Second)

	if !s.Tick() {
		t.Fatal("Tick should finish session")
	}

	if !s.Finished() {
		t.Fatalf("Test is not finished")
	}

	if s.Remaining() != 0 {
		t.Fatalf("duration = %v", s.Remaining())
	}
}

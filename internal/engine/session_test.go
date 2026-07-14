package engine_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/alirezaudev/ttype/internal/engine"
)

func newTestSession(t *testing.T, target string, duration time.Duration) (*engine.Session, *engine.FakeClock) {
	t.Helper()

	clock := engine.NewFakeClock(time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC))
	s, err := engine.NewSession(func() (string, error) { return target, nil }, duration, clock)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
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

func TestSessionRestartResetsAndReloads(t *testing.T) {
	t.Parallel()

	calls := 0
	newTarget := func() (string, error) {
		calls++
		return fmt.Sprintf("run%d", calls), nil
	}
	clock := engine.NewFakeClock(time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC))
	s, err := engine.NewSession(newTarget, 15*time.Second, clock)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}

	s.InputRune('r')
	clock.Advance(2 * time.Second)

	if err := s.Restart(); err != nil {
		t.Fatalf("Restart: %v", err)
	}

	if s.Cursor() != 0 || s.Input() != nil {
		t.Fatalf("cursor=%d input=%v", s.Cursor(), s.Input())
	}
	if got := s.Target(); got != "run2" {
		t.Fatalf("Target() = %q, want %q", got, "run2")
	}
	if s.Remaining() != 15*time.Second {
		t.Fatalf("Remaining() = %v, want 15s", s.Remaining())
	}
	if calls != 2 {
		t.Fatalf("newTarget called %d times, want 2", calls)
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

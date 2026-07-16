package tui

import (
	"testing"
	"time"

	"github.com/alirezaudev/ttype/internal/engine"
	tea "github.com/charmbracelet/bubbletea"
)

func newModelSession(t *testing.T) (*engine.Session, *engine.FakeClock) {
	t.Helper()

	clock := engine.NewFakeClock(time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC))
	cfg := engine.Config{Kind: engine.TestKindTimed, Duration: 15 * time.Second}
	s, err := engine.NewSession(func() (string, error) { return "abc", nil }, cfg, clock)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	return s, clock
}

func TestFormatClock(t *testing.T) {
	t.Parallel()

	if got := formatClock(10 * time.Second); got != "0:10" {
		t.Fatalf("formatClock(10s) = %q, want 0:10", got)
	}
	if got := formatClock(9900 * time.Millisecond); got != "0:10" {
		t.Fatalf("formatClock(9.9s) = %q, want 0:10", got)
	}
	if got := formatClock(9400 * time.Millisecond); got != "0:09" {
		t.Fatalf("formatClock(9.4s) = %q, want 0:09", got)
	}
	if got := formatClock(90 * time.Second); got != "1:30" {
		t.Fatalf("formatClock(90s) = %q, want 1:30", got)
	}
	if got := formatClock(-time.Second); got != "0:00" {
		t.Fatalf("formatClock(-1s) = %q, want 0:00", got)
	}
}

func TestEnterMidTestDoesNotScheduleTick(t *testing.T) {
	t.Parallel()

	session, _ := newModelSession(t)
	m := NewTestModel(session)
	session.InputRune('a')

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Fatal("enter should not schedule another tick")
	}
	if session.Finished() {
		t.Fatal("session should restart, not finish")
	}
}

func TestEnterAfterFinishDoesNotScheduleTick(t *testing.T) {
	t.Parallel()

	session, clock := newModelSession(t)
	m := NewTestModel(session)
	session.InputRune('a')
	clock.Advance(16 * time.Second)
	session.Tick()

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Fatal("enter should not schedule another tick")
	}
	if session.Finished() {
		t.Fatal("enter should restart the session")
	}
}

func TestTickAlwaysReschedules(t *testing.T) {
	t.Parallel()

	session, clock := newModelSession(t)
	m := NewTestModel(session)
	session.InputRune('a')
	clock.Advance(16 * time.Second)

	_, cmd := m.Update(tickMsg(time.Now()))
	if cmd == nil {
		t.Fatal("tick should always schedule the next tick")
	}
	if !session.Finished() {
		t.Fatal("session should finish when time is up")
	}
}

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

	tests := []struct {
		in       time.Duration
		expected string
	}{
		{in: 10 * time.Second, expected: "0:10"},
		{in: 9900 * time.Millisecond, expected: "0:10"},
		{in: 9400 * time.Millisecond, expected: "0:09"},
		{in: 90 * time.Second, expected: "1:30"},
		{in: -time.Second, expected: "0:00"},
	}

	for _, test := range tests {
		if got := formatClock(test.in); got != test.expected {
			t.Fatalf("formatClock(%s) = %q, want %s", test.in.String(), got, test.expected)
		}

	}
}

func TestEnterMidTestDoesNotScheduleTick(t *testing.T) {
	t.Parallel()

	session, _ := newModelSession(t)
	m := NewTestModel(session, engine.Config{}, defaultTheme())
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
	m := NewTestModel(session, engine.Config{}, defaultTheme())
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
	m := NewTestModel(session, engine.Config{}, defaultTheme())
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

func TestTypingWidth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		width int
		want  int
	}{
		{name: "typical terminal clamps to width-4", width: 80, want: 76},
		{name: "capped at 100", width: 120, want: 100},
		{name: "at the cap boundary", width: 104, want: 100},
		{name: "narrow clamps to width-4", width: 22, want: 18},
		{name: "tiny hits the floor", width: 13, want: 10},
	}

	for _, test := range tests {
		m := TestModel{width: test.width}
		if got := m.typingWidth(); got != test.want {
			t.Errorf("%s: width=%d -> %d, want %d", test.name, test.width, got, test.want)
		}
	}
}

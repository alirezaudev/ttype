package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/engine"
	tea "github.com/charmbracelet/bubbletea"
)

type fakeSource struct {
	gen func() (string, error)
}

func (f fakeSource) Generate(domain.GenerateOptions) (string, error) {
	return f.gen()
}

func fixedSource(target string) fakeSource {
	return fakeSource{gen: func() (string, error) { return target, nil }}
}

func newModelSession(t *testing.T) (*engine.Session, *engine.FakeClock) {
	t.Helper()

	clock := engine.NewFakeClock(time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC))
	cfg := domain.TestConfig{Kind: domain.TestKindTimed, Duration: 15}
	s, err := engine.NewSession(cfg, fixedSource("abc"), clock)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	return s, clock
}

func newTestModelFor(t *testing.T, target string, cfg domain.TestConfig) TestModel {
	t.Helper()

	clock := engine.NewFakeClock(time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC))
	s, err := engine.NewSession(cfg, fixedSource(target), clock)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	m := NewTestModel(s, cfg, defaultTheme())
	m.capsProbe = nil
	m.setSize(80, 24)
	return m
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
	m := NewTestModel(session, domain.TestConfig{}, defaultTheme())
	session.InputRune('a')

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Fatal("enter should not schedule another tick")
	}
	if session.State() == domain.SessionFinished {
		t.Fatal("session should restart, not finish")
	}
}

func TestEnterAfterFinishDoesNotScheduleTick(t *testing.T) {
	t.Parallel()

	session, clock := newModelSession(t)
	m := NewTestModel(session, domain.TestConfig{}, defaultTheme())
	session.InputRune('a')
	clock.Advance(16 * time.Second)
	session.Tick()

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Fatal("enter should not schedule another tick")
	}
	if session.State() == domain.SessionFinished {
		t.Fatal("enter should restart the session")
	}
}

func TestTickAlwaysReschedules(t *testing.T) {
	t.Parallel()

	session, clock := newModelSession(t)
	m := NewTestModel(session, domain.TestConfig{}, defaultTheme())
	session.InputRune('a')
	clock.Advance(16 * time.Second)

	_, cmd := m.Update(tickMsg(time.Now()))
	if cmd == nil {
		t.Fatal("tick should always schedule the next tick")
	}
	if session.State() != domain.SessionFinished {
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

func TestBlindModeHidesTheWordInProgress(t *testing.T) {
	t.Parallel()

	cfg := domain.TestConfig{Kind: domain.TestKindTimed, Duration: domain.Duration60, Blind: true}
	m := newTestModelFor(t, "the cat sat", cfg)
	for _, r := range "the ca" {
		m.session.InputRune(r)
	}

	view := stripANSI(m.View())
	if strings.Contains(view, "ca") {
		t.Fatalf("view = %q, want the word in progress hidden", view)
	}
	if !strings.Contains(view, "the") {
		t.Fatalf("view = %q, want the committed word revealed", view)
	}
}

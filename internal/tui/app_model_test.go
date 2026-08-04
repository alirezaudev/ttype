package tui

import (
	"testing"
	"time"

	"github.com/alirezaudev/ttype/internal/engine"
	tea "github.com/charmbracelet/bubbletea"
)

func newAppModel(t *testing.T) (AppModel, *engine.FakeClock) {
	t.Helper()
	clock := engine.NewFakeClock(time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC))
	cfg := engine.Config{Kind: engine.TestKindTimed, Duration: 15 * time.Second}
	newTarget := func() (string, error) { return "abc def", nil }
	session, err := engine.NewSession(newTarget, cfg, clock)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	return NewAppModel(cfg, newTarget, session), clock
}

func TestWindowSizeFansOutToTestModel(t *testing.T) {
	t.Parallel()

	m, _ := newAppModel(t)
	next, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	app := next.(AppModel)

	if app.test.width != 120 || app.test.height != 40 {
		t.Fatalf("test model size = %dx%d, want 120x40", app.test.width, app.test.height)
	}
}

func TestOpenSettingsFromResult(t *testing.T) {
	t.Parallel()

	m, clock := newAppModel(t)
	m.test.session.InputRune('a')
	clock.Advance(16 * time.Second)
	next, _ := m.Update(tickMsg(time.Now()))
	m = next.(AppModel)

	if m.phase != phaseResult {
		t.Fatalf("phase = %v, want phaseResult", m.phase)
	}

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	m = next.(AppModel)

	if m.phase != phaseSettings {
		t.Fatalf("phase = %v after s, want phaseSettings", m.phase)
	}
}

func TestSettingsApplyRestarts(t *testing.T) {
	t.Parallel()

	m, clock := newAppModel(t)
	m.test.session.InputRune('a')
	clock.Advance(16 * time.Second)
	next, _ := m.Update(tickMsg(time.Now()))
	m = next.(AppModel)

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	m = next.(AppModel)

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(AppModel)

	if m.phase != phaseTest {
		t.Fatalf("phase = %v after apply, want phaseTest", m.phase)
	}
	if m.test.session.Finished() {
		t.Fatal("restarted session should not be finished")
	}
}

func TestSettingsCancelReturnsToResult(t *testing.T) {
	t.Parallel()

	m, clock := newAppModel(t)
	m.test.session.InputRune('a')
	clock.Advance(16 * time.Second)
	next, _ := m.Update(tickMsg(time.Now()))
	m = next.(AppModel)

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	m = next.(AppModel)

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(AppModel)

	if m.phase != phaseResult {
		t.Fatalf("phase = %v after esc, want phaseResult", m.phase)
	}
}

func TestRestartAfterFinishTransitionsToTest(t *testing.T) {
	t.Parallel()

	m, clock := newAppModel(t)
	m.test.session.InputRune('a')
	clock.Advance(16 * time.Second)

	next, _ := m.Update(tickMsg(time.Now()))
	m = next.(AppModel)

	if m.phase != phaseResult {
		t.Fatalf("phase = %v after finish, want phaseResult", m.phase)
	}

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(AppModel)

	if m.phase != phaseTest {
		t.Fatalf("phase = %v after restart, want phaseTest", m.phase)
	}
	if m.test.session.Finished() {
		t.Fatal("restarted session should not be finished")
	}
}

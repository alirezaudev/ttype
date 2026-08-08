package tui

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/engine"
	tea "github.com/charmbracelet/bubbletea"
)

func runeKey(r rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
}

func newAppModel(t *testing.T) (AppModel, *engine.FakeClock) {
	t.Helper()
	clock := engine.NewFakeClock(time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC))
	cfg := domain.TestConfig{Kind: domain.TestKindTimed, Duration: 15}
	source := fixedSource("abc def")
	session, err := engine.NewSession(cfg, source, clock)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	return NewAppModel(cfg, source, nil, session, nil), clock
}

func finishTest(m AppModel, clock *engine.FakeClock) AppModel {
	m.test.session.InputRune('a')
	clock.Advance(16 * time.Second)
	next, _ := m.Update(tickMsg(time.Now()))
	m = next.(AppModel)
	m.finishedAt = m.finishedAt.Add(-resultsKeyGrace)
	return m
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
	m = finishTest(m, clock)

	if m.phase != phaseResult {
		t.Fatalf("phase = %v, want phaseResult", m.phase)
	}

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	m = next.(AppModel)

	if m.phase != phaseSettings {
		t.Fatalf("phase = %v after s, want phaseSettings", m.phase)
	}
}

func TestSettingsApplyRestarts(t *testing.T) {
	t.Parallel()

	m, clock := newAppModel(t)
	m = finishTest(m, clock)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	m = next.(AppModel)

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(AppModel)

	if m.phase != phaseTest {
		t.Fatalf("phase = %v after apply, want phaseTest", m.phase)
	}
	if m.test.session.State() == domain.SessionFinished {
		t.Fatal("restarted session should not be finished")
	}
}

func TestSettingsCancelReturnsToTest(t *testing.T) {
	t.Parallel()

	m, _ := newAppModel(t)
	m.test.session.InputRune('a')
	next, _ := m.Update(tickMsg(time.Now().Add(1 * time.Second)))
	m = next.(AppModel)

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlS}) // settings
	m = next.(AppModel)

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc}) // return back to test
	m = next.(AppModel)

	if m.phase != phaseTest {
		t.Fatalf("phase = %v after esc, want phaseTest (%v)", m.phase, phaseTest)
	}
}

func TestSettingsCancelReturnsToResult(t *testing.T) {
	t.Parallel()

	m, clock := newAppModel(t)
	m = finishTest(m, clock)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	m = next.(AppModel)

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(AppModel)

	if m.phase != phaseResult {
		t.Fatalf("phase = %v after esc, want phaseResult (%v)", m.phase, phaseResult)
	}
}

func TestRestartAfterFinishTransitionsToTest(t *testing.T) {
	t.Parallel()

	m, clock := newAppModel(t)
	m = finishTest(m, clock)

	if m.phase != phaseResult {
		t.Fatalf("phase = %v after finish, want phaseResult", m.phase)
	}

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(AppModel)

	if m.phase != phaseTest {
		t.Fatalf("phase = %v after restart, want phaseTest", m.phase)
	}
	if m.test.session.State() == domain.SessionFinished {
		t.Fatal("restarted session should not be finished")
	}
}

func TestResultsGraceSwallowsKeysThenReleases(t *testing.T) {
	t.Parallel()

	m, clock := newAppModel(t)
	m.test.session.InputRune('a')
	clock.Advance(16 * time.Second)
	next, _ := m.Update(tickMsg(time.Now()))
	m = next.(AppModel)

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(AppModel)
	if m.phase != phaseResult {
		t.Fatal("enter within the grace window restarted the test")
	}

	m.finishedAt = m.finishedAt.Add(-resultsKeyGrace)
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(AppModel)
	if m.phase != phaseTest {
		t.Fatalf("phase = %v after the grace window, want phaseTest", m.phase)
	}
}

func TestResultsGraceLetsEscapeThrough(t *testing.T) {
	t.Parallel()

	m, clock := newAppModel(t)
	m.test.session.InputRune('a')
	clock.Advance(16 * time.Second)
	next, _ := m.Update(tickMsg(time.Now()))
	m = next.(AppModel)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("esc within the grace window should still quit")
	}
}

func TestCopyOnResultsReportsClipboardFailureTruthfully(t *testing.T) {
	m, clock := newAppModel(t)
	m = finishTest(m, clock)

	original := copyToClipboardFn
	t.Cleanup(func() { copyToClipboardFn = original })

	var copied string
	copyToClipboardFn = func(text string) error {
		copied = text
		return nil
	}

	next, cmd := m.Update(runeKey('C'))
	m = next.(AppModel)
	if cmd == nil {
		t.Fatal("C produced no command")
	}
	next, _ = m.Update(cmd())
	m = next.(AppModel)

	if !strings.Contains(copied, "WPM") {
		t.Fatalf("share line = %q", copied)
	}
	if m.notice.kind != noticeInfo || m.notice.empty() {
		t.Fatalf("notice = %+v, want an info notice", m.notice)
	}

	copyToClipboardFn = func(string) error { return errors.New("no clipboard tool") }
	next, cmd = m.Update(runeKey('C'))
	m = next.(AppModel)
	next, _ = m.Update(cmd())
	m = next.(AppModel)

	if m.notice.kind != noticeError || !strings.Contains(m.notice.text, "clipboard unavailable") {
		t.Fatalf("notice = %+v, want the clipboard failure", m.notice)
	}
}

func TestNextKeyDismissesTheNotice(t *testing.T) {
	t.Parallel()

	m, clock := newAppModel(t)
	m = finishTest(m, clock)
	m.notice = infoNotice("Copied to clipboard")

	next, _ := m.Update(runeKey('x'))
	m = next.(AppModel)

	if !m.notice.empty() {
		t.Fatalf("notice = %+v, want it dismissed", m.notice)
	}
}

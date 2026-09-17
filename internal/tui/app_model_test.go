package tui

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/engine"
	"github.com/alirezaudev/ttype/internal/storage"
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
	return NewAppModel(Options{Config: cfg, Provider: source, Session: session}), clock
}

func finishTest(m AppModel, clock *engine.FakeClock) AppModel {
	m.test.session.InputRune('a')
	clock.Advance(16 * time.Second)
	next, _ := m.Update(tickMsg{loop: m.test.tickLoop})
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

	next, _ := m.Update(runeKey('S'))
	m = next.(AppModel)

	if m.phase != phaseSettings {
		t.Fatalf("phase = %v after s, want phaseSettings", m.phase)
	}
}

func TestSettingsApplyRestarts(t *testing.T) {
	t.Parallel()

	m, clock := newAppModel(t)
	m = finishTest(m, clock)

	next, _ := m.Update(runeKey('S'))
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
	next, _ := m.Update(tickMsg{loop: m.test.tickLoop})
	m = next.(AppModel)

	next, _ = m.Update(runeKey('S')) // settings
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

	next, _ := m.Update(runeKey('S'))
	m = next.(AppModel)

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(AppModel)

	if m.phase != phaseResult {
		t.Fatalf("phase = %v after esc, want phaseResult (%v)", m.phase, phaseResult)
	}
}

// The tick in flight lands on the settings panel and is dropped, so closing it
// has to start the clock again or a timed run never ends.
func TestTimedRunEndsAfterSettingsMidRun(t *testing.T) {
	t.Parallel()

	m, clock := newAppModel(t)
	next, _ := m.Update(runeKey('a'))
	m = next.(AppModel)
	inFlight := tickMsg{loop: m.test.tickLoop}

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	m = next.(AppModel)
	next, _ = m.Update(inFlight)
	m = next.(AppModel)
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(AppModel)
	if m.phase != phaseTest || cmd == nil {
		t.Fatalf("esc from settings: phase = %v, tick scheduled = %v", m.phase, cmd != nil)
	}

	clock.Advance(16 * time.Second)
	next, _ = m.Update(cmd())
	if m = next.(AppModel); m.phase != phaseResult {
		t.Fatalf("phase = %v once time is up, want phaseResult", m.phase)
	}
}

func TestTimedRunEndsAfterHelp(t *testing.T) {
	t.Parallel()

	m, clock := newAppModel(t)
	inFlight := tickMsg{loop: m.test.tickLoop}
	next, _ := m.Update(runeKey('?'))
	m = next.(AppModel)
	if m.phase != phaseHelp {
		t.Fatalf("phase = %v after ?, want phaseHelp", m.phase)
	}
	next, _ = m.Update(inFlight)
	m = next.(AppModel)
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(AppModel)
	if cmd == nil {
		t.Fatal("closing help scheduled no tick")
	}

	next, _ = m.Update(runeKey('a'))
	m = next.(AppModel)
	clock.Advance(16 * time.Second)
	next, _ = m.Update(cmd())
	if m = next.(AppModel); m.phase != phaseResult {
		t.Fatalf("phase = %v once time is up, want phaseResult", m.phase)
	}
}

// Settings closed before its tick landed: that tick reaches the test, and must
// not keep a second loop going next to the new one.
func TestTickFromAnOldLoopStops(t *testing.T) {
	t.Parallel()

	m, _ := newAppModel(t)
	old := tickMsg{loop: m.test.tickLoop}
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	m = next.(AppModel)
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(AppModel)

	if _, cmd := m.Update(old); cmd != nil {
		t.Fatal("a tick from before settings scheduled another")
	}
	if _, cmd := m.Update(tickMsg{loop: m.test.tickLoop}); cmd == nil {
		t.Fatal("the current loop stopped ticking")
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
	next, _ := m.Update(tickMsg{loop: m.test.tickLoop})
	m = next.(AppModel)

	for _, r := range "rL " {
		next, _ = m.Update(runeKey(r))
		m = next.(AppModel)
		if m.phase != phaseResult {
			t.Fatalf("%q within the grace window left the result screen", r)
		}
	}

	m.finishedAt = m.finishedAt.Add(-resultsKeyGrace)
	next, _ = m.Update(runeKey('r'))
	m = next.(AppModel)
	if m.phase != phaseTest {
		t.Fatalf("phase = %v after the grace window, want phaseTest", m.phase)
	}
}

// Nobody types tab or enter into the text, so they need no grace.
func TestResultsRestartKeysSkipTheGrace(t *testing.T) {
	t.Parallel()

	for _, keyType := range []tea.KeyType{tea.KeyTab, tea.KeyEnter} {
		m, clock := newAppModel(t)
		m.test.session.InputRune('a')
		clock.Advance(16 * time.Second)
		next, _ := m.Update(tickMsg{loop: m.test.tickLoop})
		m = next.(AppModel)

		next, _ = m.Update(tea.KeyMsg{Type: keyType})
		if m = next.(AppModel); m.phase != phaseTest {
			t.Fatalf("%v right after the result: phase = %v, want a new test", keyType, m.phase)
		}
	}
}

func TestResultsGraceLetsEscapeThrough(t *testing.T) {
	t.Parallel()

	m, clock := newAppModel(t)
	m.test.session.InputRune('a')
	clock.Advance(16 * time.Second)
	next, _ := m.Update(tickMsg{loop: m.test.tickLoop})
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

func TestHelpOverlayOnlyOpensBeforeTyping(t *testing.T) {
	t.Parallel()

	m, _ := newAppModel(t)
	next, _ := m.Update(runeKey('?'))
	m = next.(AppModel)
	if m.phase != phaseHelp {
		t.Fatalf("phase = %v, want phaseHelp", m.phase)
	}

	next, _ = m.Update(runeKey('x'))
	m = next.(AppModel)
	if m.phase != phaseTest {
		t.Fatalf("phase = %v after dismissing help, want phaseTest", m.phase)
	}

	m.test.session.InputRune('a')
	next, _ = m.Update(runeKey('?'))
	m = next.(AppModel)
	if m.phase == phaseHelp {
		t.Fatal("? opened the overlay mid-test instead of typing it")
	}
	if got := string(m.test.session.Input()); !strings.HasSuffix(got, "?") {
		t.Fatalf("input = %q, want the ? typed", got)
	}
}

func TestHelpOverlayListsTheBindings(t *testing.T) {
	t.Parallel()

	help := NewHelpOverlay(defaultTheme())
	help.setSize(100, 40)
	view := stripANSI(help.View())

	for _, want := range []string{"ctrl+bksp/ctrl+w", "delete word", "C", "copy result", "--zen"} {
		if !strings.Contains(view, want) {
			t.Errorf("help overlay missing %q", want)
		}
	}
}

func TestModePickerAppliesAndRestarts(t *testing.T) {
	t.Parallel()

	m, clock := newAppModel(t)
	m = finishTest(m, clock)

	next, _ := m.Update(runeKey('M'))
	m = next.(AppModel)
	if m.phase != phaseModePicker {
		t.Fatalf("phase = %v, want phaseModePicker", m.phase)
	}

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = next.(AppModel)
	picked := m.modePicker.Selected()

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(AppModel)

	if m.cfg.TextMode != picked {
		t.Fatalf("mode = %q, want %q", m.cfg.TextMode, picked)
	}
	if m.phase != phaseTest {
		t.Fatalf("phase = %v after picking a mode, want phaseTest", m.phase)
	}
}

func TestWelcomeAppliesDefaultsAndStartsTheTest(t *testing.T) {
	t.Parallel()

	m, _ := newAppModel(t)
	m.phase = phaseWelcome
	m.welcome = NewWelcome(m.cfg, m.theme)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = next.(AppModel)
	picked := m.welcome.Config().TextMode

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(AppModel)

	if m.phase != phaseTest {
		t.Fatalf("phase = %v after the welcome screen, want phaseTest", m.phase)
	}
	if m.cfg.TextMode != picked {
		t.Fatalf("mode = %q, want %q", m.cfg.TextMode, picked)
	}
}

func TestVersionCheckReachesTheFooter(t *testing.T) {
	t.Parallel()

	m, _ := newAppModel(t)
	next, _ := m.Update(VersionCheckedMsg{Info: domain.VersionInfo{
		Local: "1.0.0", Latest: "1.1.0", UpdateAvailable: true,
	}})
	m = next.(AppModel)

	if !m.version.UpdateAvailable || !m.test.version.UpdateAvailable {
		t.Fatal("the version check did not reach the models")
	}

	m.setSize(80, 24)
	if !strings.Contains(stripANSI(m.View()), "new") {
		t.Fatalf("footer does not flag the update:\n%s", stripANSI(m.View()))
	}
}

func TestCtrlSOpensSettingsDuringATest(t *testing.T) {
	t.Parallel()

	m, _ := newAppModel(t)
	m.test.session.InputRune('a')

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	m = next.(AppModel)

	if m.phase != phaseSettings {
		t.Fatalf("phase = %v after ctrl+s mid-test, want phaseSettings", m.phase)
	}
	if got := string(m.test.session.Input()); got != "a" {
		t.Fatalf("input = %q, want ctrl+s not to be typed", got)
	}
}

func TestCtrlSStillOpensSettingsFromResults(t *testing.T) {
	t.Parallel()

	m, clock := newAppModel(t)
	m = finishTest(m, clock)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	m = next.(AppModel)

	if m.phase != phaseSettings {
		t.Fatalf("phase = %v, want phaseSettings", m.phase)
	}
}

func TestReplayFromTheResultScreen(t *testing.T) {
	t.Parallel()

	m, clock := newAppModel(t)
	m = finishTest(m, clock)

	next, cmd := m.Update(runeKey('p'))
	m = next.(AppModel)
	if m.phase != phaseReplay || cmd == nil {
		t.Fatalf("phase = %v, cmd nil = %v; want the replay playing", m.phase, cmd == nil)
	}

	// The run's only keystroke came at its first instant.
	for range 3 {
		next, _ = m.Update(replayTickMsg(time.Now()))
		m = next.(AppModel)
	}
	if !m.replay.done() {
		t.Fatal("the replay did not reach the end")
	}
	if view := stripANSI(m.View()); !strings.Contains(view, "replay") {
		t.Fatalf("view = %q, want the replay screen", view)
	}

	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(AppModel)
	if m.phase != phaseResult {
		t.Fatalf("phase = %v after esc, want back on the result screen", m.phase)
	}
}

func appModelWithStore(t *testing.T, cfg domain.TestConfig, noSave bool) (AppModel, *engine.FakeClock, *storage.JSONStore) {
	t.Helper()

	dir := t.TempDir()
	store, err := storage.NewJSONStore(storage.Dirs{Data: dir + "/data", Config: dir + "/config"})
	if err != nil {
		t.Fatalf("NewJSONStore: %v", err)
	}
	clock := engine.NewFakeClock(time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC))
	source := fixedSource("abc def")
	session, err := engine.NewSession(cfg, source, clock)
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	m := NewAppModel(Options{Config: cfg, Provider: source, Session: session, Store: store, NoSave: noSave})
	return m, clock, store
}

func TestNoSaveKeepsTheRunOutOfHistory(t *testing.T) {
	t.Parallel()

	cfg := domain.TestConfig{Kind: domain.TestKindTimed, Duration: 15}
	for _, noSave := range []bool{false, true} {
		m, clock, store := appModelWithStore(t, cfg, noSave)
		m = finishTest(m, clock)

		results, err := store.ListResults(10)
		if err != nil {
			t.Fatalf("ListResults: %v", err)
		}
		if want := map[bool]int{false: 1, true: 0}[noSave]; len(results) != want {
			t.Fatalf("noSave=%v: %d results saved, want %d", noSave, len(results), want)
		}
		if _, err := store.LoadReplay(m.result.ID); noSave && err == nil {
			t.Fatal("--no-save still wrote a replay")
		}
	}
}

// The next plain ttype would have no text, so custom must never become the
// saved mode, and neither must the length of the text.
func TestCustomTextIsNotSavedAsTheDefault(t *testing.T) {
	t.Parallel()

	cfg := domain.TestConfig{Kind: domain.TestKindWords, WordCount: 2, TextMode: domain.TextModeCustom, Theme: "dracula"}
	m, _, store := appModelWithStore(t, cfg, false)
	if err := store.SaveSettings(domain.Settings{DefaultMode: domain.TextModeGo, DefaultDuration: 30, Language: "spanish"}); err != nil {
		t.Fatalf("SaveSettings: %v", err)
	}

	m.persistConfigDefaults()

	settings, err := store.LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings: %v", err)
	}
	if settings.DefaultMode != domain.TextModeGo || settings.DefaultWordCount != 0 || settings.Language != "spanish" {
		t.Fatalf("settings = %+v, want mode, length and language untouched", settings)
	}
	if settings.Theme != "dracula" {
		t.Fatalf("theme = %q, want the theme still saved", settings.Theme)
	}
}

type offlineSource struct{}

func (offlineSource) Generate(opts domain.GenerateOptions) (string, error) {
	if opts.Language != "" {
		return "", errors.New("offline")
	}
	return "abc def", nil
}

func TestPickingALanguageThatCannotLoad(t *testing.T) {
	t.Parallel()

	m, clock := newAppModel(t)
	m.provider = offlineSource{}
	m = finishTest(m, clock)

	next, _ := m.Update(runeKey('L'))
	m = next.(AppModel)
	m.languagePicker, _, _, _ = m.languagePicker.Update(LanguagesLoadedMsg{IDs: []string{"french"}})
	m.languagePicker, _, _, _ = m.languagePicker.Update(tea.KeyMsg{Type: tea.KeyDown})
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = next.(AppModel)

	if m.phase != phaseLanguagePicker || m.cfg.Language != "" {
		t.Fatalf("phase %v, language %q: want the picker still open and nothing chosen", m.phase, m.cfg.Language)
	}
	if !strings.Contains(m.languagePicker.View(), "french isn't downloaded") {
		t.Fatal("the picker doesn't say why")
	}
}

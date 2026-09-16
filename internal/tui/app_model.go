package tui

import (
	"fmt"
	"time"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/engine"
	"github.com/alirezaudev/ttype/internal/storage"
	"github.com/alirezaudev/ttype/internal/text/langcache"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

const resultsKeyGrace = 700 * time.Millisecond

type appPhase int

const (
	phaseWelcome appPhase = iota
	phaseTest
	phaseResult
	phaseSettings
	phaseLanguagePicker
	phaseModePicker
	phaseHelp
	phaseReplay
)

type sizable interface {
	setSize(width, height int)
}

type AppModel struct {
	cfg            domain.TestConfig
	provider       engine.TextSource
	langCache      *langcache.Cache
	store          storage.Store
	phase          appPhase
	navStack       []appPhase
	test           TestModel
	result         domain.Result
	pbUpdate       storage.PBUpdate
	recording      domain.Replay
	replay         ReplayModel
	finishedAt     time.Time
	notice         statusNotice
	settings       SettingsPanel
	languagePicker LanguagePicker
	modePicker     ModePicker
	welcome        Welcome
	help           HelpOverlay
	theme          Theme
	version        domain.VersionInfo
	checkVersion   func() domain.VersionInfo
	quitOnFinish   bool
	finished       bool
	width          int
	height         int
}

// Options carries everything the app model needs from the composition root.
// tui never imports app, so the wiring comes in through here.
type Options struct {
	Config   domain.TestConfig
	Provider engine.TextSource
	Cache    *langcache.Cache
	Store    storage.Store
	Session  *engine.Session
	Version  domain.VersionInfo
	Welcome  bool
	// QuitOnFinish leaves the results screen out of the way so the caller can
	// print the result itself (--output json).
	QuitOnFinish bool
	// CheckVersion runs once in the background; nil skips the check.
	CheckVersion func() domain.VersionInfo
}

func NewAppModel(opts Options) AppModel {
	theme := ResolveTheme(opts.Config.Theme)
	m := AppModel{
		cfg:          opts.Config,
		provider:     opts.Provider,
		langCache:    opts.Cache,
		store:        opts.Store,
		version:      opts.Version,
		checkVersion: opts.CheckVersion,
		quitOnFinish: opts.QuitOnFinish,
		theme:        theme,
		test:         NewTestModel(opts.Session, opts.Config, theme, opts.Version),
	}
	if opts.Welcome {
		m.phase = phaseWelcome
		m.welcome = NewWelcome(opts.Config, theme)
	} else {
		m.phase = phaseTest
	}
	return m
}

func (m *AppModel) setSize(width, height int) {
	m.width, m.height = width, height
	for _, s := range m.sizables() {
		s.setSize(width, height)
	}
}

func (m *AppModel) sizables() []sizable {
	return []sizable{&m.test, &m.settings, &m.languagePicker, &m.modePicker, &m.welcome, &m.help, &m.replay}
}

func (m *AppModel) pushPhase(next appPhase) {
	m.navStack = append(m.navStack, m.phase)
	m.phase = next
}

func (m *AppModel) popPhase() tea.Cmd {
	if len(m.navStack) == 0 {
		return tea.Quit
	}

	end := len(m.navStack) - 1
	prev := m.navStack[end]
	m.navStack = m.navStack[:end]
	if prev == phaseTest && m.test.session.State() == domain.SessionFinished {
		return m.restartTest()
	}

	m.phase = prev
	return nil
}

// Result reports the finished run, if the session got that far.
func (m AppModel) Result() (domain.Result, bool) {
	return m.result, m.finished
}

func (m AppModel) Init() tea.Cmd {
	if m.checkVersion == nil {
		return m.test.Init()
	}
	return tea.Batch(m.test.Init(), checkVersionCmd(m.checkVersion))
}

// VersionCheckedMsg carries the result of the background release check.
type VersionCheckedMsg struct{ Info domain.VersionInfo }

func checkVersionCmd(check func() domain.VersionInfo) tea.Cmd {
	return func() tea.Msg { return VersionCheckedMsg{Info: check()} }
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.setSize(msg.Width, msg.Height)
		return m, nil
	case VersionCheckedMsg:
		m.version = msg.Info
		m.test.version = msg.Info
		return m, nil
	case OpenLanguagePickerMsg:
		return m, m.openLanguagePicker()
	case OpenSettingsMsg:
		return m, m.openSettings()
	case OpenModePickerMsg:
		return m, m.openModePicker()
	case clipboardCopiedMsg:
		if msg.err != nil {
			m.notice = errorNotice("clipboard unavailable — run ttype doctor")
		} else {
			m.notice = infoNotice("Copied to clipboard")
		}
		return m, nil
	case tea.MouseMsg:
		return m, footerClickCmd(msg, m.width, m.height, m.cfg, m.version)
	case browserOpenedMsg:
		if msg.err != nil {
			m.notice = errorNotice("could not open the browser")
		}
		return m, nil
	case tea.KeyMsg:
		if isQuitKey(msg) {
			return m, tea.Quit
		}
		// The welcome screen is the root: it owns esc, which confirms.
		if m.phase == phaseWelcome {
			return m.updateWelcome(msg)
		}
		// q and Q are typed input on the test screen, so only esc backs out there.
		if isEscKey(msg) || (m.phase != phaseTest && isBackKey(msg)) {
			return m, m.popPhase()
		}

		// Keystrokes still in flight when the timer fires must not press
		// anything on the screen that just appeared.
		if m.phase == phaseResult && m.inResultsGrace() {
			return m, nil
		}
		m.notice = statusNotice{}

		// ctrl+s opens settings from anywhere a test or its result is on
		// screen; the pickers and overlays own their own keys.
		if key.Matches(msg, appKeys.Settings) && (m.phase == phaseTest || m.phase == phaseResult) {
			return m, m.openSettings()
		}

		// The help key only exists before the first keystroke — after that "?"
		// is a character the target may well contain.
		if m.phase == phaseTest && key.Matches(msg, testKeys.Help) &&
			m.test.session.State() == domain.SessionReady {
			return m, m.openHelp()
		}

		if m.phase == phaseResult {
			switch {
			case key.Matches(msg, resultsKeys.Restart):
				return m, m.restartTest()
			case key.Matches(msg, resultsKeys.Replay):
				return m, m.openReplay()
			case key.Matches(msg, resultsKeys.Copy):
				return m, copyResultCmd(m.result)
			case key.Matches(msg, resultsKeys.Settings):
				return m, m.openSettings()
			case key.Matches(msg, resultsKeys.Mode):
				return m, m.openModePicker()
			case key.Matches(msg, resultsKeys.Language):
				return m, m.openLanguagePicker()
			case key.Matches(msg, resultsKeys.Update) && m.version.UpdateAvailable:
				return m, openURLCmd(GitHubURL + "/releases")
			}
		}
	}

	switch m.phase {
	case phaseWelcome:
		return m.updateWelcome(msg)
	case phaseTest:
		return m.updateTest(msg)
	case phaseResult:
		return m, nil
	case phaseSettings:
		return m.updateSettings(msg)
	case phaseLanguagePicker:
		return m.updateLanguagePicker(msg)
	case phaseModePicker:
		return m.updateModePicker(msg)
	case phaseHelp:
		return m.updateHelp(msg)
	case phaseReplay:
		next, cmd, done := m.replay.Update(msg)
		m.replay = next
		if done {
			return m, m.popPhase()
		}
		return m, cmd
	}

	return m, nil
}

func (m *AppModel) openSettings() tea.Cmd {
	m.settings = NewSettingsPanel(m.cfg, m.theme)
	m.settings.setSize(m.width, m.height)
	m.pushPhase(phaseSettings)
	return nil
}

func (m *AppModel) openHelp() tea.Cmd {
	m.help = NewHelpOverlay(m.theme)
	m.help.setSize(m.width, m.height)
	m.pushPhase(phaseHelp)
	return nil
}

func (m *AppModel) openModePicker() tea.Cmd {
	m.modePicker = NewModePicker(m.cfg.TextMode, m.theme)
	m.modePicker.setSize(m.width, m.height)
	m.pushPhase(phaseModePicker)
	return nil
}

// openReplay plays the run that just finished, straight from memory, so it
// works even when replays are not saved.
func (m *AppModel) openReplay() tea.Cmd {
	if len(m.recording.Events) == 0 {
		m.notice = errorNotice("nothing to replay")
		return nil
	}
	replay, err := NewReplayModel(m.result, m.recording, m.theme)
	if err != nil {
		m.notice = errorNotice(err.Error())
		return nil
	}
	m.replay = replay
	m.replay.setSize(m.width, m.height)
	m.pushPhase(phaseReplay)
	return m.replay.Init()
}

func (m *AppModel) openLanguagePicker() tea.Cmd {
	m.languagePicker = NewLanguagePicker(m.langCache, m.cfg.Language, m.theme)
	m.languagePicker.setSize(m.width, m.height)
	m.pushPhase(phaseLanguagePicker)
	return m.languagePicker.Init()
}

func (m AppModel) updateLanguagePicker(msg tea.Msg) (tea.Model, tea.Cmd) {
	next, cmd, done, apply := m.languagePicker.Update(msg)
	m.languagePicker = next
	if !done {
		return m, cmd
	}
	if apply {
		m.cfg.Language = m.languagePicker.Selected()
		m.settings.cfg.Language = m.cfg.Language
	}
	return m, tea.Batch(cmd, m.popPhase())
}

func (m AppModel) updateWelcome(msg tea.Msg) (tea.Model, tea.Cmd) {
	next, cmd, done := m.welcome.Update(msg)
	m.welcome = next
	if !done {
		return m, cmd
	}

	m.cfg = m.welcome.Config()
	m.theme = ResolveTheme(m.cfg.Theme)
	m.persistConfigDefaults()
	m.markOnboarded()
	return m, tea.Batch(cmd, m.restartTest())
}

func (m *AppModel) markOnboarded() {
	if m.store == nil {
		return
	}
	settings, err := m.store.LoadSettings()
	if err != nil {
		return
	}
	settings.Onboarded = true
	_ = m.store.SaveSettings(settings)
}

func (m AppModel) updateModePicker(msg tea.Msg) (tea.Model, tea.Cmd) {
	next, cmd, done, apply := m.modePicker.Update(msg)
	m.modePicker = next
	if !done {
		return m, cmd
	}
	if !apply {
		return m, tea.Batch(cmd, m.popPhase())
	}
	m.cfg.TextMode = m.modePicker.Selected()
	m.persistConfigDefaults()
	return m, tea.Batch(cmd, m.restartTest())
}

func (m AppModel) updateHelp(msg tea.Msg) (tea.Model, tea.Cmd) {
	next, cmd, done := m.help.Update(msg)
	m.help = next
	if !done {
		return m, cmd
	}
	return m, tea.Batch(cmd, m.popPhase())
}

func (m AppModel) updateTest(msg tea.Msg) (tea.Model, tea.Cmd) {
	next, cmd := m.test.Update(msg)
	m.test = next.(TestModel)
	if m.test.session.State() == domain.SessionFinished {
		m.result, m.pbUpdate, m.notice = m.finishResult()
		m.recording = domain.Replay{Target: m.test.session.Target(), Events: m.test.session.Events()}
		m.finished = true
		if m.quitOnFinish {
			return m, tea.Quit
		}
		m.phase = phaseResult
		m.finishedAt = time.Now()
		return m, tea.Batch(cmd, tea.ClearScreen)
	}
	return m, cmd
}

func (m AppModel) inResultsGrace() bool {
	if m.finishedAt.IsZero() {
		return false
	}
	return time.Since(m.finishedAt) < resultsKeyGrace
}

func (m AppModel) finishResult() (domain.Result, storage.PBUpdate, statusNotice) {
	result, err := m.test.session.Result()
	if err != nil {
		return result, storage.PBUpdate{}, errorNotice(fmt.Sprintf("result not saved: %s", err))
	}
	if m.store == nil {
		return result, storage.PBUpdate{}, statusNotice{}
	}

	pb, err := m.store.SaveResult(result)
	if err != nil {
		return result, storage.PBUpdate{}, errorNotice(fmt.Sprintf("result not saved: %s", err))
	}
	m.saveReplay(result.ID)
	return result, pb, statusNotice{}
}

// Best effort: a missing recording only costs the playback, not the result.
func (m AppModel) saveReplay(id string) {
	replays, ok := m.store.(storage.ReplayStore)
	if !ok {
		return
	}
	events := m.test.session.Events()
	if len(events) == 0 {
		return
	}
	_ = replays.SaveReplay(id, domain.Replay{Target: m.test.session.Target(), Events: events})
}

func (m AppModel) updateSettings(msg tea.Msg) (tea.Model, tea.Cmd) {
	next, cmd, done, apply := m.settings.Update(msg)
	m.settings = next
	if !done {
		return m, cmd
	}
	if !apply {
		return m, tea.Batch(cmd, m.popPhase())
	}
	m.cfg = m.settings.cfg
	m.theme = ResolveTheme(m.cfg.Theme)
	m.persistConfigDefaults()
	return m, m.restartTest()
}

func (m *AppModel) persistConfigDefaults() {
	if m.store == nil {
		return
	}
	settings, err := m.store.LoadSettings()
	if err != nil {
		return
	}
	settings.Theme = m.cfg.Theme
	settings.DefaultWidth = m.cfg.Width
	settings.Language = m.cfg.Language
	settings.DefaultMode = m.cfg.TextMode
	settings.Punctuation = m.cfg.Punctuation
	settings.Numbers = m.cfg.Numbers
	settings.Blind = m.cfg.Blind
	settings.Zen = m.cfg.Zen
	settings.DefaultMinWPM = m.cfg.MinWPM
	if m.cfg.IsWordsMode() {
		settings.DefaultWordCount = m.cfg.WordCount
	} else {
		settings.DefaultWordCount = 0
		if m.cfg.Duration <= 0 {
			settings.DefaultDuration = domain.Duration60
		} else {
			settings.DefaultDuration = m.cfg.Duration
		}
	}
	_ = m.store.SaveSettings(settings)
}

func (m *AppModel) restartTest() tea.Cmd {
	session, err := engine.NewSession(m.cfg, m.provider, nil)
	if err != nil {
		return nil
	}
	m.test = NewTestModel(session, m.cfg, m.theme, m.version)
	m.test.setSize(m.width, m.height)
	m.phase = phaseTest
	m.navStack = nil
	m.finishedAt = time.Time{}
	m.notice = statusNotice{}
	return tea.Batch(tea.ClearScreen, m.test.Init())
}

func (m AppModel) View() string {
	switch m.phase {
	case phaseWelcome:
		return m.welcome.View()
	case phaseSettings:
		return m.settings.View()
	case phaseLanguagePicker:
		return m.languagePicker.View()
	case phaseModePicker:
		return m.modePicker.View()
	case phaseHelp:
		return m.help.View()
	case phaseReplay:
		return m.replay.View()
	case phaseResult:
		return renderResult(m.result, m.pbUpdate, m.theme, m.width, m.height, m.version, m.notice)
	default:
		return m.test.View()
	}
}

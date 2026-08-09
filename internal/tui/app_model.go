package tui

import (
	"fmt"
	"time"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/engine"
	"github.com/alirezaudev/ttype/internal/storage"
	"github.com/alirezaudev/ttype/internal/text/langcache"
	tea "github.com/charmbracelet/bubbletea"
)

const resultsKeyGrace = 700 * time.Millisecond

type appPhase int

const (
	phaseTest appPhase = iota
	phaseResult
	phaseSettings
	phaseLanguagePicker
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
	finishedAt     time.Time
	notice         statusNotice
	settings       SettingsPanel
	languagePicker LanguagePicker
	theme          Theme
	width          int
	height         int
}

func NewAppModel(cfg domain.TestConfig, provider engine.TextSource, cache *langcache.Cache, session *engine.Session, store storage.Store) AppModel {
	theme := ResolveTheme(cfg.Theme)
	return AppModel{
		cfg:       cfg,
		provider:  provider,
		langCache: cache,
		store:     store,
		theme:     theme,
		test:      NewTestModel(session, cfg, theme),
	}
}

func (m *AppModel) sizables() []sizable {
	return []sizable{&m.test, &m.settings, &m.languagePicker}
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

func (m AppModel) Init() tea.Cmd {
	return m.test.Init()
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		for _, s := range m.sizables() {
			s.setSize(msg.Width, msg.Height)
		}
		return m, nil
	case OpenLanguagePickerMsg:
		m.languagePicker = NewLanguagePicker(m.langCache, m.cfg.Language, m.theme)
		m.languagePicker.setSize(m.width, m.height)
		m.pushPhase(phaseLanguagePicker)
		return m, m.languagePicker.Init()
	case clipboardCopiedMsg:
		if msg.err != nil {
			m.notice = errorNotice("clipboard unavailable — run ttype doctor")
		} else {
			m.notice = infoNotice("Copied to clipboard")
		}
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			return m, m.popPhase()
		}

		if m.phase == phaseResult && m.inResultsGrace() {
			return m, nil
		}
		m.notice = statusNotice{}

		switch msg.String() {
		case "enter", "tab":
			if m.phase == phaseResult {
				return m, m.restartTest()
			}
		case "C":
			if m.phase == phaseResult {
				return m, copyResultCmd(m.result)
			}
		case "S":
			if m.phase == phaseResult {
				m.settings = NewSettingsPanel(m.cfg, m.theme)
				m.settings.setSize(m.width, m.height)
				m.pushPhase(phaseSettings)
				return m, nil
			}
		case "L":
			if m.phase == phaseResult {
				m.languagePicker = NewLanguagePicker(m.langCache, m.cfg.Language, m.theme)
				m.languagePicker.setSize(m.width, m.height)
				m.pushPhase(phaseLanguagePicker)
				return m, m.languagePicker.Init()
			}
		case "ctrl+s":
			m.settings = NewSettingsPanel(m.cfg, m.theme)
			m.settings.setSize(m.width, m.height)
			m.pushPhase(phaseSettings)
			return m, nil
		}
	}

	switch m.phase {
	case phaseTest:
		return m.updateTest(msg)
	case phaseResult:
		return m, nil
	case phaseSettings:
		return m.updateSettings(msg)
	case phaseLanguagePicker:
		return m.updateLanguagePicker(msg)
	}

	return m, nil
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

func (m AppModel) updateTest(msg tea.Msg) (tea.Model, tea.Cmd) {
	next, cmd := m.test.Update(msg)
	m.test = next.(TestModel)
	if m.test.session.State() == domain.SessionFinished {
		m.result, m.notice = m.finishResult()
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

func (m AppModel) finishResult() (domain.Result, statusNotice) {
	result, err := m.test.session.Result()
	if err != nil {
		return result, errorNotice(fmt.Sprintf("result not saved: %s", err))
	}
	if m.store == nil {
		return result, statusNotice{}
	}
	if err := m.store.SaveResult(result); err != nil {
		return result, errorNotice(fmt.Sprintf("result not saved: %s", err))
	}
	return result, statusNotice{}
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
	m.test = NewTestModel(session, m.cfg, m.theme)
	m.test.setSize(m.width, m.height)
	m.phase = phaseTest
	m.navStack = nil
	m.finishedAt = time.Time{}
	m.notice = statusNotice{}
	return tea.Batch(tea.ClearScreen, m.test.Init())
}

func (m AppModel) View() string {
	switch m.phase {
	case phaseSettings:
		return m.settings.View()
	case phaseLanguagePicker:
		return m.languagePicker.View()
	case phaseResult:
		return renderResult(m.result, m.theme, m.width, m.height, m.notice)
	default:
		return m.test.View()
	}
}

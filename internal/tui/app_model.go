package tui

import (
	"github.com/alirezaudev/ttype/internal/engine"
	tea "github.com/charmbracelet/bubbletea"
)

type appPhase int

const (
	phaseTest appPhase = iota
	phaseResult
	phaseSettings
)

type sizable interface {
	setSize(width, height int)
}

type AppModel struct {
	cfg       engine.Config
	newTarget func() (string, error)
	phase     appPhase
	navStack  []appPhase
	test      TestModel
	result    resultSnapshot
	settings  SettingsPanel
	theme     Theme
	width     int
	height    int
}

func NewAppModel(cfg engine.Config, newTarget func() (string, error), session *engine.Session) AppModel {
	theme := ResolveTheme(cfg.Theme)
	return AppModel{
		cfg:       cfg,
		newTarget: newTarget,
		theme:     theme,
		test:      NewTestModel(session, cfg, theme),
	}
}

func (m *AppModel) sizables() []sizable {
	return []sizable{&m.test, &m.settings}
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
	if prev == phaseTest && m.test.session.Finished() {
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
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			return m, m.popPhase()
		case "enter", "tab":
			if m.phase == phaseResult {
				newMsg := m.restartTest()
				return m, newMsg
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
	}

	return m, nil
}

func (m AppModel) updateTest(msg tea.Msg) (tea.Model, tea.Cmd) {
	next, cmd := m.test.Update(msg)
	m.test = next.(TestModel)
	if m.test.session.Finished() {
		m.result = snapshotResult(m.test.session, m.cfg)
		m.phase = phaseResult
		return m, tea.Batch(cmd, tea.ClearScreen)
	}
	return m, cmd
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
	return m, m.restartTest()
}

func (m *AppModel) restartTest() tea.Cmd {
	session, err := engine.NewSession(m.newTarget, m.cfg, nil)
	if err != nil {
		return nil
	}
	m.test = NewTestModel(session, m.cfg, m.theme)
	m.test.setSize(m.width, m.height)
	m.phase = phaseTest
	m.navStack = nil
	return tea.Batch(tea.ClearScreen, m.test.Init())

}

func (m AppModel) View() string {
	switch m.phase {
	case phaseSettings:
		return m.settings.View()
	case phaseResult:
		return renderResult(m.result, m.theme, m.width, m.height)
	default:
		return m.test.View()
	}
}

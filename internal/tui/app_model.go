package tui

import (
	"github.com/alirezaudev/ttype/internal/engine"
	tea "github.com/charmbracelet/bubbletea"
)

type appPhase int

const (
	phaseTest appPhase = iota
	phaseResults
)

type AppModel struct {
	cfg       engine.Config
	newTarget func() (string, error)
	phase     appPhase
	test      TestModel
	width     int
	height    int
}

func NewAppModel(cfg engine.Config, newTarget func() (string, error), session *engine.Session) AppModel {
	return AppModel{
		cfg:       cfg,
		newTarget: newTarget,
		test:      NewTestModel(session),
	}
}

func (m AppModel) Init() tea.Cmd {
	return m.test.Init()
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.test.setSize(msg.Width, msg.Height)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter", "tab":
			if m.phase == phaseResults {
				return m.restart()
			}
		}
	}

	switch m.phase {
	case phaseTest:
		return m.updateTest(msg)
	case phaseResults:
		return m, nil
	}

	return m, nil
}

func (m AppModel) updateTest(msg tea.Msg) (tea.Model, tea.Cmd) {
	next, cmd := m.test.Update(msg)
	m.test = next.(TestModel)
	if m.test.session.Finished() {
		m.phase = phaseResults
		return m, tea.Batch(cmd, tea.ClearScreen)
	}
	return m, cmd
}

func (m AppModel) restart() (tea.Model, tea.Cmd) {
	session, err := engine.NewSession(m.newTarget, m.cfg, nil)
	if err != nil {
		return m, nil
	}
	m.test = NewTestModel(session)
	m.test.setSize(m.width, m.height)
	m.phase = phaseTest
	return m, tea.Batch(tea.ClearScreen, m.test.Init())
}

func (m AppModel) View() string {
	return m.test.View()
}

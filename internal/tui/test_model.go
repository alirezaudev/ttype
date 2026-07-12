package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/alirezaudev/ttype/internal/engine"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	correctStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	incorrectStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	cursorStyle    = lipgloss.NewStyle().Reverse(true)
)

type TestModel struct {
	session *engine.Session
	width   int
	height  int
}

func NewTestModel(text string, duration time.Duration) TestModel {
	return TestModel{
		session: engine.NewSession(text, duration),
	}
}

func (m TestModel) Init() tea.Cmd {
	return tick()
}

func (m TestModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tickMsg:
		if m.session.Tick() {
			return m, nil
		}
		return m, tick()
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "backspace":
			m.session.Backspace()
		default:
			if len(msg.Runes) == 0 || m.session.Finished() {
				return m, nil
			}

			m.session.InputRune(msg.Runes[0])
		}
	}

	return m, nil
}

func (m TestModel) View() string {
	var out strings.Builder

	if m.session.Finished() {
		out.WriteString(fmt.Sprintf("WPM: %d", int(m.session.WPM())))
		return m.center(out.String())
	}

	remaining := (m.session.Remaining() + time.Second - 1) / time.Second
	out.WriteString(fmt.Sprintf("%d\n", int(remaining)))

	cursor := m.session.Cursor()
	input := m.session.Input()
	for i, r := range m.session.TargetRunes() {
		switch {
		case i == cursor:
			out.WriteString(cursorStyle.Render(string(r)))
		case i < cursor && input[i] == r:
			out.WriteString(correctStyle.Render(string(r)))
		case i < cursor:
			out.WriteString(incorrectStyle.Render(string(r)))
		default:
			out.WriteRune(r)
		}
	}

	return m.center(out.String())
}

func (m TestModel) center(content string) string {
	if m.width == 0 || m.height == 0 {
		return content
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

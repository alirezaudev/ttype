package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	correctStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	incorrectStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	cursorStyle    = lipgloss.NewStyle().Reverse(true)
)

type TestModel struct {
	Text       string
	Typed      []rune
	Index      int
	Keystrokes int
}

func (tm TestModel) Init() tea.Cmd {
	return nil
}

func (tm TestModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return tm, tea.Quit
		default:
			char := msg.String()[0]
			tm.Index++
			tm.Keystrokes++
			tm.Typed = append(tm.Typed, rune(char))
			if tm.Index+2 == len(tm.Text) {
				// TODO calculate the result
				return tm, tea.Quit
			}
		}
	}

	return tm, nil
}

func (tm TestModel) View() string {
	var out strings.Builder

	for i, r := range []rune(tm.Text) {
		switch {
		case i == tm.Index:
			out.WriteString(cursorStyle.Render(string(r)))
		case i < len(tm.Typed) && tm.Typed[i] == r:
			out.WriteString(correctStyle.Render(string(r)))
		case i < len(tm.Typed):
			out.WriteString(incorrectStyle.Render(string(r)))
		default:
			out.WriteRune(r)
		}
	}

	return out.String()
}

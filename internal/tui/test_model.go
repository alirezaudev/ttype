package tui

import (
	"fmt"
	"strings"
	"time"

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
	Keystrokes int
	StartedAt  time.Time
	EndedAt    time.Time
	Duration   time.Duration
	finished   bool
}

func (tm TestModel) Init() tea.Cmd {
	return tick()
}

func (tm TestModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	index := len(tm.Typed)
	switch msg := msg.(type) {
	case tickMsg:
		if !tm.StartedAt.IsZero() && time.Since(tm.StartedAt) >= tm.Duration {
			tm.finished = true
			tm.EndedAt = time.Now()
			return tm, nil
		}
		return tm, tick()
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return tm, tea.Quit
		case "backspace":
			if index == 0 || tm.finished {
				return tm, nil
			}
			index--
			tm.Typed = tm.Typed[:index]
		default:
			if tm.finished {
				return tm, nil
			}
			if tm.StartedAt.IsZero() {
				tm.StartedAt = time.Now()
			}
			char := msg.String()[0]
			index++
			tm.Keystrokes++
			tm.Typed = append(tm.Typed, rune(char))
			if index-1 == len(tm.Text) {
				tm.finished = true
				tm.EndedAt = time.Now()
			}
		}
	}

	return tm, nil
}

func (tm TestModel) View() string {
	var out strings.Builder

	if tm.finished {
		correct := 0
		for i := 0; i < len(tm.Typed); i++ {
			if tm.Typed[i] == rune(tm.Text[i]) {
				correct++
			}
		}
		wpm := (float64(correct) / 5.0) / tm.EndedAt.Sub(tm.StartedAt).Minutes()
		out.WriteString(fmt.Sprintf("WPM: %d", int(wpm)))
		return out.String()
	}
	left := tm.Duration
	if !tm.StartedAt.IsZero() {
		left -= time.Since(tm.StartedAt)
	}
	if left < 0 {
		left = 0
	}

	out.WriteString(fmt.Sprintf("%d\n", int((left+time.Second-1)/time.Second)))
	for i, r := range []rune(tm.Text) {
		switch {
		case i == len(tm.Typed):
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

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

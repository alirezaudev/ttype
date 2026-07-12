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
	Parts      []rune
	Typed      []rune
	Keystrokes int
	StartedAt  time.Time
	EndedAt    time.Time
	Duration   time.Duration
	width      int
	height     int
	finished   bool
}

func NewTestModel(text string, duration time.Duration) TestModel {
	return TestModel{
		Parts:    []rune(text),
		Duration: duration,
	}
}

func (tm TestModel) Init() tea.Cmd {
	return tick()
}

func (tm TestModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	index := len(tm.Typed)
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		tm.width = msg.Width
		tm.height = msg.Height
		return tm, nil
	case tickMsg:
		if tm.finished {
			return tm, nil
		}
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
			if len(msg.Runes) == 0 || tm.finished || index >= len(tm.Parts) {
				return tm, nil
			}
			if tm.StartedAt.IsZero() {
				tm.StartedAt = time.Now()
			}
			char := msg.Runes[0]
			index++
			tm.Keystrokes++
			tm.Typed = append(tm.Typed, char)
			if index == len(tm.Parts) {
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
		for i := 0; i < len(tm.Typed) && i < len(tm.Parts); i++ {
			if tm.Typed[i] == tm.Parts[i] {
				correct++
			}
		}
		wpm := (float64(correct) / 5.0) / tm.EndedAt.Sub(tm.StartedAt).Minutes()
		out.WriteString(fmt.Sprintf("WPM: %d", int(wpm)))
		return tm.center(out.String())
	}
	left := tm.Duration
	if !tm.StartedAt.IsZero() {
		left -= time.Since(tm.StartedAt)
	}
	if left < 0 {
		left = 0
	}

	out.WriteString(fmt.Sprintf("%d\n", int((left+time.Second-1)/time.Second)))
	for i, r := range tm.Parts {
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

	return tm.center(out.String())
}

func (tm TestModel) center(content string) string {
	if tm.width == 0 || tm.height == 0 {
		return content
	}
	return lipgloss.Place(tm.width, tm.height, lipgloss.Center, lipgloss.Center, content)
}

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

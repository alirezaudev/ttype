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
	cfg     engine.Config
	width   int
	height  int
}

func NewTestModel(session *engine.Session, cfg engine.Config) TestModel {
	return TestModel{session: session, cfg: cfg}
}

func (m *TestModel) setSize(width, height int) {
	m.width, m.height = width, height
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
		m.session.Tick()
		return m, tick()
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			_ = m.session.Restart()
			return m, nil
		case "backspace":
			m.session.Backspace()
		case "ctrl+w", "ctrl+h", "alt+ctrl+h", "alt+backspace":
			m.session.DeleteWord()
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

	status := fmt.Sprintf(
		"wpm %-3d · raw %-3d · acc %-3d%% · err %-3d",
		int(m.session.WPM()),
		int(m.session.RawWPM()),
		int(m.session.Accuracy()),
		m.session.Incorrect(),
	)

	if m.session.Finished() {
		out.WriteString(status)
		return m.center(out.String())
	}

	if m.session.Kind() == engine.TestKindTimed {
		out.WriteString(fmt.Sprintf("%s · %s\n\n", formatClock(m.session.Remaining()), status))
	} else {
		done, total := m.session.WordsProgress()
		totalStr := fmt.Sprintf("%d", total)
		out.WriteString(fmt.Sprintf("%*d/%s · %s\n\n", len(totalStr), done, totalStr, status))
	}

	cursor := m.session.Cursor()
	input := m.session.Input()
	target := m.session.TargetRunes()
	lines := wordWrapIndices(target, m.typingWidth())
	from, to := visibleLineWindow(lines, cursor, 3)
	for li, line := range lines[from:to] {
		for i := line.start; i < line.end; i++ {
			r := target[i]
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
		if li < len(lines)-1 {
			out.WriteByte('\n')
		}
	}
	if m.width == 0 {
		return out.String()
	}
	block := lipgloss.NewStyle().Width(m.typingWidth()).Render(out.String())
	return m.center(block)
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

func formatClock(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	secs := int(d.Round(time.Second).Seconds())
	return fmt.Sprintf("%d:%02d", secs/60, secs%60)
}

func (m TestModel) typingWidth() int {
	if m.width == 0 {
		return len(m.session.TargetRunes())
	}

	cap := 100
	if m.cfg.Width > 0 {
		cap = m.cfg.Width
	}
	w := min(m.width, cap)

	if w < 20 {
		w = 20
	}

	if w > m.width-4 {
		w = m.width - 4
	}

	if w < 10 {
		w = 10
	}

	return w
}

type lineSpan struct {
	start int
	end   int
}

func wordWrapIndices(text []rune, width int) []lineSpan {
	if len(text) == 0 {
		return nil
	}

	if width < 1 {
		width = 1
	}

	var lines []lineSpan
	start := 0
	for start < len(text) {
		if len(text)-start <= width {
			// Last line, or the only line if the text fits on a single line.
			lines = append(lines, lineSpan{start: start, end: len(text)})
			break
		}

		end := start + width
		breakAt := end
		space := false
		for i := end; i > start; i-- {
			if text[i-1] == ' ' {
				breakAt = i // Break at the last space so the next word stays together.
				space = true
				break
			}
		}

		if !space {
			breakAt = end // No space found, so hard-break at the width.
		}

		lines = append(lines, lineSpan{start: start, end: breakAt})
		start = breakAt

		for start < len(text) && text[start] == ' ' {
			start++ // Skip leading spaces on the next line.
		}
	}

	return lines
}

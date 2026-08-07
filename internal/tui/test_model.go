package tui

import (
	"strings"
	"time"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/engine"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TestModel struct {
	session   *engine.Session
	cfg       domain.TestConfig
	theme     Theme
	capsProbe func() bool
	width     int
	height    int
}

func NewTestModel(session *engine.Session, cfg domain.TestConfig, theme Theme) TestModel {
	return TestModel{
		session:   session,
		cfg:       cfg,
		theme:     theme,
		capsProbe: newCapsLockMonitor().on,
	}
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

	out.WriteString(renderHUD(m.session, m.theme, m.typingWidth(), m.cfg))
	out.WriteString("\n\n")

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
				out.WriteString(m.theme.Cursor.Render(string(r)))
			case i < cursor && input[i] == r:
				out.WriteString(m.theme.Correct.Render(string(r)))
			case i < cursor:
				out.WriteString(m.theme.Incorrect.Render(string(r)))
			default:
				out.WriteString(m.theme.Pending.Render(string(r)))
			}
		}
		if li < len(lines)-1 {
			out.WriteByte('\n')
		}
	}
	if m.width == 0 {
		return out.String()
	}
	out.WriteByte('\n')
	out.WriteString(m.hintLine())
	block := lipgloss.NewStyle().Width(m.typingWidth()).Render(out.String())
	return m.center(block)
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

func (m TestModel) hintLine() string {
	if m.capsWarnActive() {
		return renderCapsWarn(m.theme)
	}
	if !m.session.Started() {
		return m.theme.Help.Render("start typing to begin")
	}
	return ""
}

func (m TestModel) capsWarnActive() bool {
	if m.capsProbe != nil && m.capsProbe() {
		return true
	}
	return m.session.CapsLockSuspected()
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

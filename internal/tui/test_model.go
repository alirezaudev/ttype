package tui

import (
	"strings"
	"time"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/engine"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TestModel struct {
	session   *engine.Session
	cfg       domain.TestConfig
	theme     Theme
	version   domain.VersionInfo
	capsProbe func() bool
	hideLive  bool
	width     int
	height    int
}

func NewTestModel(session *engine.Session, cfg domain.TestConfig, theme Theme, ver domain.VersionInfo) TestModel {
	return TestModel{
		session:   session,
		cfg:       cfg,
		theme:     theme,
		version:   ver,
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
		switch {
		case key.Matches(msg, testKeys.Restart):
			_ = m.session.Restart()
		case key.Matches(msg, testKeys.ToggleLive):
			m.hideLive = !m.hideLive
		case key.Matches(msg, testKeys.Backspace):
			m.session.Backspace()
		case key.Matches(msg, testKeys.DeleteWord):
			m.session.DeleteWord()
		default:
			if len(msg.Runes) == 0 || msg.Alt || m.session.State() == domain.SessionFinished {
				return m, nil
			}
			m.session.InputRune(msg.Runes[0])
		}
	}

	return m, nil
}

func (m TestModel) View() string {
	var out strings.Builder

	if !m.cfg.Zen {
		out.WriteString(renderHUD(m.session, m.theme, m.typingWidth(), m.cfg, m.hideLive))
		out.WriteString("\n\n")
	}
	out.WriteString(m.renderWords())

	if m.width == 0 {
		return out.String()
	}
	out.WriteByte('\n')
	out.WriteString(m.hintLine())
	if !m.cfg.Zen {
		out.WriteString("\n\n")
		out.WriteString(m.theme.Help.Render(testHelpLine()))
	}
	block := lipgloss.NewStyle().Width(m.typingWidth()).Render(out.String())
	return composeWithBottomFooter(
		block,
		renderFooterSection(m.theme, m.cfg, m.version, m.width),
		m.width, m.height,
	)
}

func (m TestModel) renderWords() string {
	cursor := m.session.Cursor()
	input := m.session.Input()
	target := m.session.TargetRunes()
	revealed := cursor
	if m.cfg.Blind {
		revealed = blindRevealEnd(target, cursor)
	}

	lines := wordWrapIndices(target, m.typingWidth())
	from, to := visibleLineWindow(lines, cursor, 3)

	var out strings.Builder
	for i := from; i < to; i++ {
		line := lines[i]
		for j := line.start; j < line.end; j++ {
			r := target[j]
			switch {
			case j == cursor && m.cfg.Blind:
				out.WriteString(m.theme.Cursor.Render("·"))
			case j == cursor:
				out.WriteString(m.theme.Cursor.Render(string(r)))
			case j < cursor && j >= revealed:
				out.WriteString(m.theme.Pending.Render("·"))
			case j < cursor && input[j] == r:
				out.WriteString(m.theme.Correct.Render(string(r)))
			case j < cursor:
				out.WriteString(m.theme.Incorrect.Render(string(r)))
			default:
				out.WriteString(m.theme.Pending.Render(string(r)))
			}
		}
		if i < to-1 {
			out.WriteByte('\n')
		}
	}
	return out.String()
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
	if !m.cfg.Zen && m.session.State() == domain.SessionReady {
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

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

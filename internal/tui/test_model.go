package tui

import (
	"strings"
	"time"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/engine"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// wrapCache memoizes the word wrap, which only changes on restart or resize.
// It hangs off the model by pointer so Bubble Tea's value copies share it.
type wrapCache struct {
	target string
	width  int
	lines  []lineSpan
}

func (c *wrapCache) wrap(target []rune, text string, width int) []lineSpan {
	if c == nil {
		return wordWrapIndices(target, width)
	}
	if c.target != text || c.width != width || c.lines == nil {
		c.target = text
		c.width = width
		c.lines = wordWrapIndices(target, width)
	}
	return c.lines
}

type TestModel struct {
	session   *engine.Session
	cfg       domain.TestConfig
	theme     Theme
	version   domain.VersionInfo
	capsProbe func() bool
	wrap      *wrapCache
	// reorderRTL sends right-to-left text in display order.
	reorderRTL bool
	hideLive   bool
	width      int
	height     int
}

func NewTestModel(session *engine.Session, cfg domain.TestConfig, theme Theme, ver domain.VersionInfo) TestModel {
	return TestModel{
		session:    session,
		cfg:        cfg,
		theme:      theme,
		version:    ver,
		capsProbe:  newCapsLockMonitor().on,
		wrap:       &wrapCache{},
		reorderRTL: reorderRTL,
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
			// A paste is not typing.
			if msg.Paste || len(msg.Runes) == 0 || msg.Alt || m.session.State() == domain.SessionFinished {
				return m, nil
			}
			// Keys that arrive in one read (SSH, an input method) share a message.
			for _, r := range msg.Runes {
				m.session.InputRune(r)
			}
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

// cellKind groups characters by the style they render with, so a run of them
// costs one Render call instead of one per character.
type cellKind int

const (
	cellPending cellKind = iota
	cellCorrect
	cellIncorrect
)

func (m TestModel) styleFor(kind cellKind) lipgloss.Style {
	switch kind {
	case cellCorrect:
		return m.theme.Correct
	case cellIncorrect:
		return m.theme.Incorrect
	default:
		return m.theme.Pending
	}
}

func (m TestModel) renderWords() string {
	cursor := m.session.Cursor()
	target := m.session.TargetRunes()
	revealed := cursor
	if m.cfg.Blind {
		revealed = blindRevealEnd(target, cursor)
	}

	width := m.typingWidth()
	lines := m.wrap.wrap(target, m.session.Target(), width)
	from, to := visibleLineWindow(lines, cursor, 3)
	rtl := m.reorderRTL && rightToLeft(target)

	var out, run strings.Builder
	runKind := cellPending
	flush := func() {
		if run.Len() == 0 {
			return
		}
		out.WriteString(m.styleFor(runKind).Render(run.String()))
		run.Reset()
	}

	for i := from; i < to; i++ {
		line := lines[i]
		var order []int
		if rtl {
			order = visualOrder(target, line.start, line.end)
			// Right-aligned, so each line starts where the reader looks.
			if m.width > 0 {
				pad := width - runewidth.StringWidth(string(target[line.start:line.end]))
				out.WriteString(strings.Repeat(" ", max(pad, 0)))
			}
		}

		for k := 0; k < line.end-line.start; k++ {
			j := line.start + k
			r := target[j]
			if rtl {
				j = order[k]
				r = mirrorRune(target[j])
			}

			// The cursor keeps its own call: it is a single cell and its style
			// never matches the run around it.
			if j == cursor {
				flush()
				text := string(r)
				// A word's first letter is never typed yet, and before the
				// first keystroke nothing has shown it, so it stays visible.
				if m.cfg.Blind && cursor > revealed {
					text = "·"
				}
				out.WriteString(m.theme.Cursor.Render(text))
				continue
			}

			kind := cellPending
			text := string(r)
			switch {
			case j < cursor && j >= revealed:
				text = "·"
			case j < cursor:
				switch m.session.StatusAt(j) {
				case engine.KeystrokeCorrect:
					kind = cellCorrect
				case engine.KeystrokeIncorrect:
					kind = cellIncorrect
				}
			}

			if kind != runKind {
				flush()
				runKind = kind
			}
			run.WriteString(text)
		}
		flush()
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

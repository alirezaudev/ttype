package tui

import (
	"fmt"
	"strings"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type welcomeField int

const (
	welcomeMode welcomeField = iota
	welcomeDuration
	welcomeTheme
)

var welcomeDurations = []domain.Duration{
	domain.Duration15, domain.Duration30, domain.Duration60, domain.Duration120,
}

// Welcome runs once, on the very first launch, so the defaults are a choice
// rather than something to discover later in the settings panel.
type Welcome struct {
	cfg    domain.TestConfig
	field  welcomeField
	theme  Theme
	width  int
	height int
}

func NewWelcome(cfg domain.TestConfig, theme Theme) Welcome {
	return Welcome{cfg: cfg, theme: theme}
}

func (m *Welcome) setSize(width, height int) { m.width, m.height = width, height }

func (m Welcome) Config() domain.TestConfig { return m.cfg }

func (m Welcome) Update(msg tea.Msg) (Welcome, tea.Cmd, bool) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil, false
	}

	switch {
	case key.Matches(keyMsg, pickerKeys.Confirm), key.Matches(keyMsg, pickerKeys.Cancel):
		return m, nil, true
	case key.Matches(keyMsg, pickerKeys.Up):
		if m.field > welcomeMode {
			m.field--
		}
	case key.Matches(keyMsg, pickerKeys.Down):
		if m.field < welcomeTheme {
			m.field++
		}
	case key.Matches(keyMsg, pickerKeys.Left):
		m.adjust(-1)
	case key.Matches(keyMsg, pickerKeys.Right):
		m.adjust(1)
	}
	return m, nil, false
}

func (m *Welcome) adjust(dir int) {
	switch m.field {
	case welcomeMode:
		modes := domain.AllTextModes()
		m.cfg.TextMode = modes[(indexOfMode(modes, m.cfg.TextMode)+len(modes)+dir)%len(modes)]
	case welcomeDuration:
		i := 0
		for idx, d := range welcomeDurations {
			if d == m.cfg.Duration {
				i = idx
			}
		}
		m.cfg.Kind = domain.TestKindTimed
		m.cfg.WordCount = 0
		m.cfg.Duration = welcomeDurations[(i+len(welcomeDurations)+dir)%len(welcomeDurations)]
	case welcomeTheme:
		names := ThemeNames()
		i := 0
		for idx, name := range names {
			if name == m.cfg.Theme {
				i = idx
			}
		}
		m.cfg.Theme = names[(i+len(names)+dir)%len(names)]
	}
}

func indexOfMode(modes []domain.TextMode, mode domain.TextMode) int {
	for i, candidate := range modes {
		if candidate == mode {
			return i
		}
	}
	return 0
}

func (m Welcome) row(label, value string, active bool) string {
	prefix := "  "
	if active {
		prefix = "> "
	}
	return prefix + m.theme.Help.Render(fmt.Sprintf("%-10s", label+":")) + " " + value
}

func (m Welcome) View() string {
	theme := m.cfg.Theme
	if theme == "" {
		theme = ThemeDefault
	}
	mode := m.cfg.TextMode
	if mode == "" {
		mode = domain.TextModeWords
	}

	lines := []string{
		m.theme.Finished.Render("Welcome to ttype"),
		"",
		m.theme.Help.Render("Pick your defaults. You can change them any time with S."),
		"",
		m.row("mode", string(mode), m.field == welcomeMode),
		m.row("duration", fmt.Sprintf("%ds", m.cfg.Duration.Seconds()), m.field == welcomeDuration),
		m.row("theme", theme, m.field == welcomeTheme),
		"",
		m.theme.Help.Render(helpLine(pickerKeys.Up, pickerKeys.Left, pickerKeys.Confirm)),
	}

	content := strings.Join(lines, "\n")
	if m.width == 0 || m.height == 0 {
		return content
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

package tui

import (
	"fmt"
	"strings"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/text/langcache"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type settingsField int

const (
	settingsTestKind settingsField = iota
	settingsLength
	settingsMode
	settingsLanguage
	settingsWidth
	settingsTheme
	settingsPunctuation
	settingsNumbers
)

const (
	settingsLastField  = settingsNumbers
	settingsLabelWidth = 10
	settingsValueWidth = 10
)

type SettingsPanel struct {
	cfg    domain.TestConfig
	theme  Theme
	field  settingsField
	width  int
	height int
}

func NewSettingsPanel(cfg domain.TestConfig, theme Theme) SettingsPanel {
	return SettingsPanel{cfg: cfg, theme: theme}
}

func (m *SettingsPanel) setSize(width, height int) { m.width, m.height = width, height }

func (m SettingsPanel) Init() tea.Cmd { return nil }

func (m SettingsPanel) Update(msg tea.Msg) (SettingsPanel, tea.Cmd, bool, bool) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil, false, false
	}
	switch key.String() {
	case "esc", "q":
		return m, nil, true, false
	case "enter":
		if m.field == settingsLanguage {
			return m, func() tea.Msg { return OpenLanguagePickerMsg{} }, false, false
		}
		return m, nil, true, true
	case "up", "k":
		if m.field > 0 {
			m.field--
		}
	case "down", "j":
		if m.field < settingsLastField {
			m.field++
		}
	case "left", "h":
		m.adjust(-1)
	case "right", "l":
		m.adjust(1)
	}
	return m, nil, false, false
}

func (m *SettingsPanel) adjust(dir int) {
	switch m.field {
	case settingsTestKind:
		if m.cfg.Kind == domain.TestKindWords {
			m.cfg.Kind = domain.TestKindTimed
			if m.cfg.Duration <= 0 {
				m.cfg.Duration = domain.Duration60
			}
		} else {
			m.cfg.Kind = domain.TestKindWords
			if m.cfg.WordCount <= 0 {
				m.cfg.WordCount = 25
			}
		}
	case settingsLength:
		if m.cfg.Kind == domain.TestKindWords {
			m.cfg.WordCount += dir * 5
			if m.cfg.WordCount < 10 {
				m.cfg.WordCount = 10
			}
			if m.cfg.WordCount > 1000 {
				m.cfg.WordCount = 1000
			}
		} else {
			next := int(m.cfg.Duration) + dir*15
			if next < 15 {
				next = 15
			}
			m.cfg.Duration = domain.Duration(next)
		}
	case settingsMode:
		modes := domain.AllTextModes()
		idx := 0
		for i, mode := range modes {
			if mode == m.cfg.TextMode {
				idx = i
				break
			}
		}
		m.cfg.TextMode = modes[(idx+len(modes)+dir)%len(modes)]
	case settingsLanguage:
		if dir < 0 {
			m.cfg.Language = ""
		}
	case settingsWidth:
		if m.cfg.Width <= 0 && dir > 0 {
			m.cfg.Width = 60
		} else {
			m.cfg.Width += dir * 10
			if m.cfg.Width < 0 {
				m.cfg.Width = 0
			}
			if m.cfg.Width > 100 {
				m.cfg.Width = 100
			}
		}
	case settingsTheme:
		names := ThemeNames()
		current := m.cfg.Theme
		if current == "" {
			current = ThemeDefault
		}
		idx := 0
		for i, n := range names {
			if n == current {
				idx = i
				break
			}
		}
		m.cfg.Theme = names[(idx+len(names)+dir)%len(names)]
	case settingsPunctuation:
		m.cfg.Punctuation = !m.cfg.Punctuation
	case settingsNumbers:
		m.cfg.Numbers = !m.cfg.Numbers
	}
}

func toggleLabel(on bool) string {
	if on {
		return "on"
	}
	return "off"
}

func (m SettingsPanel) kindLabel() string {
	if m.cfg.Kind == domain.TestKindWords {
		return "words"
	}
	return "timed"
}

func (m SettingsPanel) lengthFieldLabel() string {
	if m.cfg.Kind == domain.TestKindWords {
		return "length"
	}
	return "duration"
}

func (m SettingsPanel) lengthLabel() string {
	if m.cfg.Kind == domain.TestKindWords {
		return fmt.Sprintf("%d words", m.cfg.WordCount)
	}
	return fmt.Sprintf("%ds", m.cfg.Duration.Seconds())
}

func (m SettingsPanel) modeLabel() string {
	if m.cfg.TextMode == "" {
		return string(domain.TextModeWords)
	}
	return string(m.cfg.TextMode)
}

func (m SettingsPanel) widthLabel() string {
	if m.cfg.Width <= 0 {
		return "auto"
	}
	return fmt.Sprintf("%d", m.cfg.Width)
}

func (m SettingsPanel) themeLabel() string {
	if m.cfg.Theme == "" {
		return ThemeDefault
	}
	return m.cfg.Theme
}

func (m SettingsPanel) row(label, value string, active bool) string {
	prefix := "  "
	if active {
		prefix = "> "
	}
	lbl := m.theme.Help.Render(fmt.Sprintf("%-*s", settingsLabelWidth, label+":"))
	val := fmt.Sprintf("%-*s", settingsValueWidth, value)
	return prefix + lbl + " " + val
}

func (m SettingsPanel) View() string {
	lines := []string{
		m.theme.Finished.Render("Settings"),
		"",
		m.row("test", m.kindLabel(), m.field == settingsTestKind),
		m.row(m.lengthFieldLabel(), m.lengthLabel(), m.field == settingsLength),
		m.row("mode", m.modeLabel(), m.field == settingsMode),
		m.row("language", langcache.DisplayName(m.cfg.Language), m.field == settingsLanguage),
		m.row("width", m.widthLabel(), m.field == settingsWidth),
		m.row("theme", m.themeLabel(), m.field == settingsTheme),
		m.row("punct", toggleLabel(m.cfg.Punctuation), m.field == settingsPunctuation),
		m.row("numbers", toggleLabel(m.cfg.Numbers), m.field == settingsNumbers),
		"",
		m.theme.Help.Render("↑/↓ field  ←/→ value  enter apply  language: enter picks  esc back"),
	}
	content := strings.Join(lines, "\n")
	if m.width == 0 || m.height == 0 {
		return content
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

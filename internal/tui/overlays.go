package tui

import (
	"fmt"
	"strings"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/text"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type settingsField int

const (
	settingsTestKind settingsField = iota
	settingsLength
	settingsLanguage
	settingsWidth
	settingsTheme
)

const (
	settingsLastField  = settingsTheme
	settingsLabelWidth = 10
	settingsValueWidth = 10
)

type SettingsPanel struct {
	cfg      domain.TestConfig
	provider *text.Provider
	theme    Theme
	field    settingsField
	picking  bool
	langs    []string
	langIdx  int
	langErr  error
	width    int
	height   int
}

func NewSettingsPanel(cfg domain.TestConfig, provider *text.Provider, theme Theme) SettingsPanel {
	return SettingsPanel{cfg: cfg, provider: provider, theme: theme}
}

func (m *SettingsPanel) setSize(width, height int) { m.width, m.height = width, height }

func (m SettingsPanel) Init() tea.Cmd { return nil }

func (m SettingsPanel) Update(msg tea.Msg) (SettingsPanel, tea.Cmd, bool, bool) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil, false, false
	}

	if m.picking {
		switch key.String() {
		case "esc", "q":
			m.picking = false
		case "up", "k":
			if m.langIdx > 0 {
				m.langIdx--
			}
		case "down", "j":
			if m.langIdx < len(m.langs)-1 {
				m.langIdx++
			}
		case "enter":
			id := ""
			if m.langIdx < len(m.langs) {
				id = m.langs[m.langIdx]
			}
			m.langErr = m.provider.UseLanguage(id)
			if m.langErr == nil {
				m.cfg.Language = id
				m.picking = false
			}
		}
		return m, nil, false, false
	}

	switch key.String() {
	case "esc", "q":
		return m, nil, true, false
	case "enter":
		if m.field == settingsLanguage {
			m.langs, m.langErr = m.provider.Languages()
			m.langs = append([]string{""}, m.langs...)
			m.langIdx = 0
			for i, id := range m.langs {
				if id == m.cfg.Language {
					m.langIdx = i
					break
				}
			}
			m.picking = true
			return m, nil, false, false
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
	case settingsLanguage:
		if dir < 0 && m.provider != nil {
			m.cfg.Language = ""
			m.langErr = m.provider.UseLanguage("")
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
	}
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
	if m.picking {
		lines := []string{m.theme.Finished.Render("Pick a language"), ""}
		if m.langErr != nil {
			lines = append(lines, m.theme.Incorrect.Render(m.langErr.Error()), "")
		}

		start := m.langIdx - 5
		if start < 0 {
			start = 0
		}
		end := start + 12
		if end > len(m.langs) {
			end = len(m.langs)
		}
		for i := start; i < end; i++ {
			prefix := "  "
			if i == m.langIdx {
				prefix = "> "
			}
			lines = append(lines, prefix+m.theme.HUDValue.Render(text.DisplayName(m.langs[i])))
		}

		lines = append(lines, "", m.theme.Help.Render("↑/↓ select  enter confirm  esc back"))
		content := strings.Join(lines, "\n")
		if m.width == 0 || m.height == 0 {
			return content
		}
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}

	lines := []string{
		m.theme.Finished.Render("Settings"),
		"",
		m.row("test", m.kindLabel(), m.field == settingsTestKind),
		m.row(m.lengthFieldLabel(), m.lengthLabel(), m.field == settingsLength),
		m.row("language", text.DisplayName(m.cfg.Language), m.field == settingsLanguage),
		m.row("width", m.widthLabel(), m.field == settingsWidth),
		m.row("theme", m.themeLabel(), m.field == settingsTheme),
		"",
		m.theme.Help.Render("↑/↓ field  ←/→ value  enter apply  language: enter picks  esc back"),
	}
	content := strings.Join(lines, "\n")
	if m.width == 0 || m.height == 0 {
		return content
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

package tui

import (
	"strings"

	"github.com/alirezaudev/ttype/internal/text"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type OpenLanguagePickerMsg struct{}

type LanguagesLoadedMsg struct {
	IDs []string
	Err error
}

func loadLanguagesCmd(provider *text.Provider) tea.Cmd {
	return func() tea.Msg {
		ids, err := provider.Languages()
		return LanguagesLoadedMsg{IDs: ids, Err: err}
	}
}

type LanguagePicker struct {
	provider *text.Provider
	current  string
	ids      []string
	filtered []string
	filter   string
	idx      int
	loading  bool
	err      error
	theme    Theme
	width    int
	height   int
}

func NewLanguagePicker(provider *text.Provider, current string, theme Theme) LanguagePicker {
	return LanguagePicker{
		provider: provider,
		current:  current,
		theme:    theme,
		loading:  provider != nil,
	}
}

func (m *LanguagePicker) setSize(width, height int) { m.width, m.height = width, height }

func (m *LanguagePicker) setSelection(id string) {
	m.filtered = m.buildFiltered()
	m.idx = 0
	for i, candidate := range m.filtered {
		if candidate == id {
			m.idx = i
			return
		}
	}
}

func (m LanguagePicker) Init() tea.Cmd {
	if m.provider == nil {
		return nil
	}
	return loadLanguagesCmd(m.provider)
}

func (m LanguagePicker) Update(msg tea.Msg) (LanguagePicker, tea.Cmd, bool, bool) {
	switch msg := msg.(type) {
	case LanguagesLoadedMsg:
		m.loading = false
		m.err = msg.Err
		m.ids = prependBuiltIn(msg.IDs)
		m.setSelection(m.current)
		return m, nil, false, false
	case tea.KeyMsg:
		if m.loading {
			if msg.String() == "esc" {
				return m, nil, true, false
			}
			return m, nil, false, false
		}
		switch msg.String() {
		case "esc":
			return m, nil, true, false
		case "enter":
			return m, nil, true, true
		case "up":
			if m.idx > 0 {
				m.idx--
			}
		case "down":
			if m.idx < len(m.filtered)-1 {
				m.idx++
			}
		case "backspace", "ctrl+h":
			if len(m.filter) > 0 {
				m.filter = m.filter[:len(m.filter)-1]
				m.setSelection(m.Selected())
			}
		default:
			if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] >= ' ' {
				m.filter += string(msg.Runes[0])
				m.idx = 0
				m.filtered = m.buildFiltered()
			}
		}
	}
	return m, nil, false, false
}

func prependBuiltIn(ids []string) []string {
	return append([]string{""}, ids...)
}

func (m LanguagePicker) buildFiltered() []string {
	base := m.ids
	if len(base) == 0 {
		base = prependBuiltIn(nil)
	}

	filter := strings.ToLower(strings.TrimSpace(m.filter))
	if filter == "" {
		return append([]string(nil), base...)
	}

	var out []string
	for _, id := range base {
		if strings.Contains(strings.ToLower(text.DisplayName(id)), filter) {
			out = append(out, id)
		}
	}
	return out
}

func (m LanguagePicker) Selected() string {
	if len(m.filtered) == 0 {
		return m.current
	}

	idx := m.idx
	if idx < 0 {
		idx = 0
	}
	if idx >= len(m.filtered) {
		idx = len(m.filtered) - 1
	}
	return m.filtered[idx]
}

func (m LanguagePicker) View() string {
	lines := []string{m.theme.Finished.Render("Pick a language"), ""}

	if m.loading {
		lines = append(lines, m.theme.Help.Render("loading languages..."))
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, strings.Join(lines, "\n"))
	}

	if m.err != nil {
		lines = append(lines, m.theme.Incorrect.Render(m.err.Error()), "")
	}

	filterLine := "filter: " + m.filter
	if m.filter == "" {
		filterLine = "filter: (type to search)"
	}
	lines = append(lines, m.theme.HUD.Render(filterLine), "")

	lines = append(lines, languageWindow(m.filtered, m.idx, m.height, m.theme)...)
	if len(m.filtered) == 0 {
		lines = append(lines, m.theme.Help.Render("no matches"))
	}

	lines = append(lines, "", m.theme.Help.Render("type filter  ↑/↓ select  enter confirm  esc back"))
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, strings.Join(lines, "\n"))
}

func languageWindow(ids []string, idx, height int, theme Theme) []string {
	if len(ids) == 0 {
		return nil
	}

	maxRows := height - 12
	if maxRows < 8 {
		maxRows = 8
	}
	if maxRows > 20 {
		maxRows = 20
	}

	start := idx - maxRows/2
	if start < 0 {
		start = 0
	}
	end := start + maxRows
	if end > len(ids) {
		end = len(ids)
		start = end - maxRows
		if start < 0 {
			start = 0
		}
	}

	var lines []string
	for i := start; i < end; i++ {
		prefix := "  "
		if i == idx {
			prefix = "> "
		}
		lines = append(lines, prefix+theme.HUDValue.Render(text.DisplayName(ids[i])))
	}
	return lines
}

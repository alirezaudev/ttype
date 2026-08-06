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
	m.idx = 0
	for i, candidate := range m.ids {
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
			if m.idx < len(m.ids)-1 {
				m.idx++
			}
		}
	}
	return m, nil, false, false
}

func prependBuiltIn(ids []string) []string {
	return append([]string{""}, ids...)
}

func (m LanguagePicker) Selected() string {
	if len(m.ids) == 0 {
		return m.current
	}

	idx := m.idx
	if idx < 0 {
		idx = 0
	}
	if idx >= len(m.ids) {
		idx = len(m.ids) - 1
	}
	return m.ids[idx]
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

	lines = append(lines, languageWindow(m.ids, m.idx, m.height, m.theme)...)
	lines = append(lines, "", m.theme.Help.Render("↑/↓ select  enter confirm  esc back"))
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

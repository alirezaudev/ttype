package tui

import (
	"strings"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type OpenModePickerMsg struct{}

type ModePicker struct {
	modes  []domain.TextMode
	idx    int
	theme  Theme
	width  int
	height int
}

func NewModePicker(current domain.TextMode, theme Theme) ModePicker {
	m := ModePicker{modes: domain.AllTextModes(), theme: theme}
	for i, mode := range m.modes {
		if mode == current {
			m.idx = i
			break
		}
	}
	return m
}

func (m *ModePicker) setSize(width, height int) { m.width, m.height = width, height }

func (m ModePicker) Selected() domain.TextMode { return m.modes[m.idx] }

// Returns the picker, a command, whether it is done and whether to apply.
func (m ModePicker) Update(msg tea.Msg) (ModePicker, tea.Cmd, bool, bool) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil, false, false
	}

	switch {
	case key.Matches(keyMsg, pickerKeys.Cancel), isBackKey(keyMsg):
		return m, nil, true, false
	case key.Matches(keyMsg, pickerKeys.Confirm):
		return m, nil, true, true
	case key.Matches(keyMsg, pickerKeys.Up), key.Matches(keyMsg, pickerKeys.Left):
		m.idx = (m.idx - 1 + len(m.modes)) % len(m.modes)
	case key.Matches(keyMsg, pickerKeys.Down), key.Matches(keyMsg, pickerKeys.Right):
		m.idx = (m.idx + 1) % len(m.modes)
	}
	return m, nil, false, false
}

func (m ModePicker) View() string {
	lines := []string{m.theme.Finished.Render("Mode"), ""}
	for i, mode := range m.modes {
		row := "  " + string(mode)
		if i == m.idx {
			lines = append(lines, m.theme.SelectedItem.Render("> "+string(mode)))
			continue
		}
		lines = append(lines, m.theme.Help.Render(row))
	}
	lines = append(lines, "", m.theme.Help.Render(helpLine(pickerKeys.Up, pickerKeys.Confirm, appKeys.Back)))

	content := strings.Join(lines, "\n")
	if m.width == 0 || m.height == 0 {
		return content
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

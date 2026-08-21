package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const helpOverlayWidth = 52

type helpOverlayKeymap struct {
	Settings key.Binding
	Mode     key.Binding
}

var helpOverlayKeys = helpOverlayKeymap{
	Settings: key.NewBinding(key.WithKeys("S"), key.WithHelp("S", "settings")),
	Mode:     key.NewBinding(key.WithKeys("M"), key.WithHelp("M", "mode picker")),
}

type HelpOverlay struct {
	theme  Theme
	width  int
	height int
}

func NewHelpOverlay(theme Theme) HelpOverlay {
	return HelpOverlay{theme: theme}
}

func (m *HelpOverlay) setSize(width, height int) { m.width, m.height = width, height }

// Anything that is not a shortcut closes the overlay, so it never traps you.
func (m HelpOverlay) Update(msg tea.Msg) (HelpOverlay, tea.Cmd, bool) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil, false
	}
	switch {
	case key.Matches(keyMsg, helpOverlayKeys.Settings):
		return m, func() tea.Msg { return OpenSettingsMsg{} }, true
	case key.Matches(keyMsg, helpOverlayKeys.Mode):
		return m, func() tea.Msg { return OpenModePickerMsg{} }, true
	}
	return m, nil, true
}

func (m HelpOverlay) View() string {
	width := helpOverlayWidth
	if m.width > 0 && m.width-4 < width {
		width = m.width - 4
	}

	lines := []string{m.theme.Finished.Render("ttype")}
	lines = append(lines, "", m.theme.HUDTitle.Render("During a test"))
	for _, b := range []key.Binding{
		testKeys.Backspace, testKeys.DeleteWord, testKeys.Skip,
		testKeys.ToggleLive, testKeys.Restart, appKeys.Exit,
	} {
		lines = append(lines, m.theme.Help.Render(overlayRow(b.Help().Key, b.Help().Desc)))
	}

	lines = append(lines, "", m.theme.HUDTitle.Render("Results"))
	for _, b := range []key.Binding{
		resultsKeys.Restart, resultsKeys.Copy, resultsKeys.Settings,
		resultsKeys.Mode, resultsKeys.Language,
	} {
		lines = append(lines, m.theme.Help.Render(overlayRow(b.Help().Key, b.Help().Desc)))
	}

	lines = append(lines, "", m.theme.HUDTitle.Render("Flags"))
	for _, example := range []struct{ flag, desc string }{
		{"--time 30", "30 second test"},
		{"--words 25", "finish 25 words"},
		{"--mode go", "type Go snippets"},
		{"--blind", "no feedback while typing"},
		{"--zen", "words only"},
	} {
		lines = append(lines, m.theme.Help.Render(overlayRow(example.flag, example.desc)))
	}

	lines = append(lines, "", m.theme.Help.Render(
		helpLine(helpOverlayKeys.Settings, helpOverlayKeys.Mode, appKeys.Back)))

	box := m.theme.Border.
		BorderStyle(lipgloss.RoundedBorder()).
		Padding(0, 1).
		Width(width).
		Render(strings.Join(lines, "\n"))

	if m.width == 0 || m.height == 0 {
		return box
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}

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
	if m.width > 0 {
		width = min(width, m.width-4)
	}
	if width < 8 {
		width = max(m.width-2, 4)
	}

	// The border and padding cost two rows and four columns; whatever is left
	// is what the sections get to fill. Zooming the terminal font can leave
	// very few rows, so the content drops a section at a time.
	budget := m.height
	if budget > 0 {
		budget -= 2
	}

	box := m.theme.Border.
		BorderStyle(lipgloss.RoundedBorder()).
		Padding(0, 1).
		Width(width)

	// Truncate rather than wrap: a wrapped row would silently cost a line the
	// budget above did not account for.
	inner := lipgloss.NewStyle().MaxWidth(max(width-2, 1))
	lines := m.lines(budget)
	for i, line := range lines {
		lines[i] = inner.Render(line)
	}

	rendered := box.Render(strings.Join(lines, "\n"))
	if m.width == 0 || m.height == 0 {
		return rendered
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, rendered)
}

// lines builds the tallest variant that fits the row budget, and hard-clips if
// even the shortest one does not.
func (m HelpOverlay) lines(budget int) []string {
	variants := [][]string{
		m.full(),
		m.withoutFlags(),
		m.duringTestOnly(),
		m.essentials(),
	}

	for _, lines := range variants {
		if budget <= 0 || len(lines) <= budget {
			return lines
		}
	}

	shortest := variants[len(variants)-1]
	if budget < 1 {
		budget = 1
	}
	return shortest[:min(len(shortest), budget)]
}

func (m HelpOverlay) full() []string {
	lines := m.withoutFlags()
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
	return append(lines, "", m.footerHint())
}

func (m HelpOverlay) withoutFlags() []string {
	lines := m.duringTestOnly()
	lines = append(lines, "", m.theme.HUDTitle.Render("Results"))
	for _, b := range []key.Binding{
		resultsKeys.Restart, resultsKeys.Copy, resultsKeys.Settings,
		resultsKeys.Mode, resultsKeys.Language,
	} {
		lines = append(lines, m.theme.Help.Render(overlayRow(b.Help().Key, b.Help().Desc)))
	}
	return append(lines, "", m.footerHint())
}

func (m HelpOverlay) duringTestOnly() []string {
	lines := []string{m.theme.Finished.Render("ttype"), "", m.theme.HUDTitle.Render("During a test")}
	for _, b := range []key.Binding{
		testKeys.Backspace, testKeys.DeleteWord, testKeys.Skip,
		testKeys.ToggleLive, appKeys.Settings, testKeys.Restart, appKeys.Exit,
	} {
		lines = append(lines, m.theme.Help.Render(overlayRow(b.Help().Key, b.Help().Desc)))
	}
	return lines
}

func (m HelpOverlay) essentials() []string {
	return []string{
		m.theme.Help.Render(overlayRow(testKeys.Restart.Help().Key, "restart")),
		m.theme.Help.Render(overlayRow(appKeys.Exit.Help().Key, "quit")),
		m.theme.Help.Render("  man ttype"),
	}
}

func (m HelpOverlay) footerHint() string {
	return m.theme.Help.Render(helpLine(helpOverlayKeys.Settings, helpOverlayKeys.Mode, appKeys.Back))
}

package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/alirezaudev/ttype/internal/text/langcache"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

type OpenLanguagePickerMsg struct{}

type LanguagesLoadedMsg struct {
	IDs []string
	Err error
}

func loadLanguagesCmd(cache *langcache.Cache) tea.Cmd {
	return func() tea.Msg {
		ids, err := cache.List()
		return LanguagesLoadedMsg{IDs: ids, Err: err}
	}
}

type LanguagePicker struct {
	cache    *langcache.Cache
	current  string
	ids      []string
	filtered []string
	filter   string
	idx      int
	// refreshing is set while the full list is fetched in the background.
	refreshing bool
	// partial is set when only the downloaded languages are known.
	partial bool
	// moved is set once the user picks something, so a late list keeps it.
	moved  bool
	err    error
	theme  Theme
	width  int
	height int
}

// NewLanguagePicker opens on what is already on disk, so it shows at once
// even offline; an old list is refreshed in the background.
func NewLanguagePicker(cache *langcache.Cache, current string, theme Theme) LanguagePicker {
	m := LanguagePicker{cache: cache, current: current, theme: theme}
	if cache != nil {
		saved := cache.SavedList()
		m.partial = len(saved) == 0
		m.refreshing = cache.ListStale()
		m.ids = m.order(saved)
	} else {
		m.ids = prependBuiltIn(nil)
	}
	m.setSelection(current)
	return m
}

// order puts the downloaded languages right after the built-in one, since
// those are the ones people switch between.
func (m LanguagePicker) order(ids []string) []string {
	var installed []string
	if m.cache != nil {
		installed = m.cache.Installed()
	}
	sort.Strings(installed)

	seen := make(map[string]bool, len(installed))
	out := prependBuiltIn(installed)
	for _, id := range installed {
		seen[id] = true
	}
	for _, id := range ids {
		if !seen[id] {
			out = append(out, id)
		}
	}
	return out
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
	if m.cache == nil || !m.refreshing {
		return nil
	}
	return loadLanguagesCmd(m.cache)
}

func (m LanguagePicker) Update(msg tea.Msg) (LanguagePicker, tea.Cmd, bool, bool) {
	switch msg := msg.(type) {
	case LanguagesLoadedMsg:
		m.refreshing = false
		if len(msg.IDs) == 0 {
			// Offline with no saved list: the downloaded ones still work.
			if m.partial && msg.Err != nil {
				m.err = fmt.Errorf("could not list more languages: %w", msg.Err)
			}
			return m, nil, false, false
		}
		m.partial = false
		m.err = nil
		selected := m.current
		if m.moved {
			selected = m.Selected()
		}
		m.ids = m.order(msg.IDs)
		m.setSelection(selected)
		return m, nil, false, false
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, nil, true, false
		case "enter":
			return m, nil, true, true
		case "up":
			m.moved = true
			if m.idx > 0 {
				m.idx--
			}
		case "down":
			m.moved = true
			if m.idx < len(m.filtered)-1 {
				m.idx++
			}
		case "backspace", "ctrl+h":
			m.moved = true
			if len(m.filter) > 0 {
				m.filter = m.filter[:len(m.filter)-1]
				m.setSelection(m.Selected())
			}
		default:
			if msg.Type == tea.KeyRunes {
				m.moved = true
				for _, r := range msg.Runes {
					if r >= ' ' {
						m.filter += string(r)
					}
				}
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
		if strings.Contains(strings.ToLower(langcache.DisplayName(id)), filter) {
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

	switch {
	case m.err != nil:
		lines = append(lines, m.theme.Incorrect.Render(m.err.Error()), "")
	case m.partial && m.refreshing:
		lines = append(lines, m.theme.Help.Render("looking for more languages..."), "")
	}

	filterLine := "filter: " + m.filter
	if m.filter == "" {
		filterLine = "filter: (type to search)"
	}
	lines = append(lines, m.theme.HUD.Render(filterLine), "")

	lines = append(lines, languageWindow(m.filtered, m.idx, m.height, m.theme, m.cache)...)
	if len(m.filtered) == 0 {
		lines = append(lines, m.theme.Help.Render("no matches"))
	}

	lines = append(lines, "", m.theme.Help.Render("type filter  ↑/↓ select  enter confirm  esc back"))
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, strings.Join(lines, "\n"))
}

func languageWindow(ids []string, idx, height int, theme Theme, cache *langcache.Cache) []string {
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

	// The rows are centered as a block, so the downloaded mark has to sit in
	// its own column: pad every name to the widest one first, otherwise the
	// ticks scatter across the list.
	nameWidth := 0
	for i := start; i < end; i++ {
		if w := runewidth.StringWidth(langcache.DisplayName(ids[i])); w > nameWidth {
			nameWidth = w
		}
	}

	var lines []string
	for i := start; i < end; i++ {
		prefix := "  "
		if i == idx {
			prefix = "> "
		}

		name := langcache.DisplayName(ids[i])
		mark := " "
		if cache != nil && ids[i] != "" && cache.Cached(ids[i]) {
			mark = "✓"
		}
		label := name + strings.Repeat(" ", nameWidth-runewidth.StringWidth(name)) + "  " + mark
		lines = append(lines, prefix+theme.HUDValue.Render(label))
	}
	return lines
}

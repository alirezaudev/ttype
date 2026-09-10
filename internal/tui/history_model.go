package tui

import (
	"fmt"
	"strings"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/storage"
	"github.com/alirezaudev/ttype/internal/text/langcache"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type historyKeymap struct {
	Watch key.Binding
}

var historyKeys = historyKeymap{
	Watch: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "watch replay")),
}

type HistoryModel struct {
	store    storage.Store
	results  []domain.Result
	table    table.Model
	replay   *ReplayModel
	theme    Theme
	notice   statusNotice
	width    int
	height   int
	quitting bool
}

func NewHistoryModel(store storage.Store, results []domain.Result, theme Theme) HistoryModel {
	columns := []table.Column{
		{Title: "Date", Width: 13},
		{Title: "Mode", Width: 9},
		{Title: "Lang", Width: 10},
		{Title: "Test", Width: 6},
		{Title: "WPM", Width: 5},
		{Title: "Raw", Width: 5},
		{Title: "Acc", Width: 5},
		{Title: "Err", Width: 4},
	}

	rows := make([]table.Row, 0, len(results))
	for _, r := range results {
		rows = append(rows, table.Row{
			r.Timestamp.Local().Format("Jan 2 15:04"),
			string(r.Config.TextMode),
			languageColumn(r.Config.Language),
			testLengthLabel(r.Config),
			fmt.Sprintf("%.0f", r.WPM),
			fmt.Sprintf("%.0f", r.RawWPM),
			fmt.Sprintf("%.0f%%", r.Accuracy),
			fmt.Sprintf("%d", keystrokeErrors(r)),
		})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(min(max(len(rows), 1), 15)+historyHeaderHeight),
	)
	styles := table.DefaultStyles()
	styles.Header = theme.HUDTitle.BorderStyle(lipgloss.NormalBorder()).BorderBottom(true)
	styles.Selected = theme.SelectedItem
	styles.Cell = theme.HUDValue
	t.SetStyles(styles)

	return HistoryModel{store: store, results: results, table: t, theme: theme}
}

// The table's height counts its header and the rule under it, so asking for
// the row count alone leaves the oldest run off the bottom.
const historyHeaderHeight = 2

const languageColumnWidth = 10

func languageColumn(id string) string {
	name := []rune(langcache.DisplayName(id))
	if len(name) <= languageColumnWidth {
		return string(name)
	}
	return string(name[:languageColumnWidth-1]) + "…"
}

func testLengthLabel(cfg domain.TestConfig) string {
	if cfg.IsWordsMode() {
		return fmt.Sprintf("%dw", cfg.WordCount)
	}
	return fmt.Sprintf("%ds", cfg.Duration.Seconds())
}

func (m HistoryModel) Init() tea.Cmd { return nil }

func (m HistoryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		if m.replay != nil {
			m.replay.setSize(msg.Width, msg.Height)
		}
		return m, nil
	case tea.KeyMsg:
		if isQuitKey(msg) {
			m.quitting = true
			return m, tea.Quit
		}
		if m.replay != nil {
			next, cmd, done := m.replay.Update(msg)
			m.replay = &next
			if done {
				m.replay = nil
			}
			return m, cmd
		}
		if isBackKey(msg) {
			m.quitting = true
			return m, tea.Quit
		}
		if key.Matches(msg, historyKeys.Watch) {
			return m, m.startReplay()
		}
	default:
		if m.replay != nil {
			next, cmd, done := m.replay.Update(msg)
			m.replay = &next
			if done {
				m.replay = nil
			}
			return m, cmd
		}
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m *HistoryModel) startReplay() tea.Cmd {
	if len(m.results) == 0 {
		return nil
	}
	result := m.results[m.table.Cursor()]

	replays, ok := m.store.(storage.ReplayStore)
	if !ok {
		m.notice = errorNotice("this store keeps no replays")
		return nil
	}
	recording, err := replays.LoadReplay(result.ID)
	if err != nil {
		m.notice = errorNotice("no replay saved for this result")
		return nil
	}

	replay, err := NewReplayModel(result, recording, m.theme)
	if err != nil {
		m.notice = errorNotice(err.Error())
		return nil
	}
	replay.setSize(m.width, m.height)
	m.replay = &replay
	m.notice = statusNotice{}
	return replay.Init()
}

func (m HistoryModel) View() string {
	if m.quitting {
		return ""
	}
	if m.replay != nil {
		return m.replay.View()
	}

	if len(m.results) == 0 {
		return "No test history yet.\n"
	}

	lines := []string{
		m.theme.Finished.Render("History"),
		"",
		m.table.View(),
		"",
	}
	if !m.notice.empty() {
		lines = append(lines, m.notice.render(m.theme), "")
	}
	lines = append(lines, m.theme.Help.Render(helpLine(historyKeys.Watch, appKeys.Back)))

	content := strings.Join(lines, "\n")
	if m.width == 0 || m.height == 0 {
		return content
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

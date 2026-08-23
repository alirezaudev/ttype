package app

import (
	"fmt"
	"io"
	"os"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/storage"
	"github.com/alirezaudev/ttype/internal/text/langcache"
	"github.com/alirezaudev/ttype/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

func RunHistory(store storage.Store, limit int, plain bool) error {
	results, err := store.ListResults(limit)
	if err != nil {
		return fmt.Errorf("load history: %w", err)
	}

	if plain || !term.IsTerminal(int(os.Stdout.Fd())) {
		printHistoryTable(os.Stdout, results)
		return nil
	}

	settings, err := store.LoadSettings()
	if err != nil {
		return fmt.Errorf("load settings: %w", err)
	}

	model := tui.NewHistoryModel(store, results, tui.ResolveTheme(settings.Theme))
	if _, err := tea.NewProgram(model, tea.WithAltScreen()).Run(); err != nil {
		return fmt.Errorf("tui: %w", err)
	}
	return nil
}

const languageColumnWidth = 10

func historyLanguage(id string) string {
	if id == "" {
		return "english"
	}

	name := []rune(langcache.DisplayName(id))
	if len(name) <= languageColumnWidth {
		return string(name)
	}
	return string(name[:languageColumnWidth-1]) + "…"
}

func printHistoryTable(out io.Writer, results []domain.Result) {
	if len(results) == 0 {
		fmt.Fprintln(out, "No test history yet.")
		return
	}

	fmt.Fprintf(out, "%-16s  %-10s  %-6s  %6s  %6s  %5s  %4s\n",
		"Date", "Lang", "Time", "WPM", "Raw", "Acc", "Err")

	for _, result := range results {
		length := fmt.Sprintf("%ds", result.Config.Duration.Seconds())
		if result.Config.IsWordsMode() {
			length = fmt.Sprintf("%dw", result.Config.WordCount)
		}

		fmt.Fprintf(out, "%-16s  %-10s  %-6s  %6.0f  %6.0f  %4.0f%%  %4d\n",
			result.Timestamp.Local().Format("Jan 2 15:04"),
			historyLanguage(result.Config.Language), length, result.WPM, result.RawWPM, result.Accuracy, result.Incorrect)
	}
}

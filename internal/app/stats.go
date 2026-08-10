package app

import (
	"fmt"
	"io"
	"os"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/storage"
)

func RunStats(store storage.Store, filter domain.StatsFilter) error {
	summary, err := store.Summary(filter)
	if err != nil {
		return fmt.Errorf("load stats: %w", err)
	}

	printStats(os.Stdout, summary)
	return nil
}

func printStats(out io.Writer, summary domain.StatsSummary) {
	if summary.TotalTests == 0 {
		if summary.FilterMode != "" {
			fmt.Fprintf(out, "No tests recorded for mode %q.\n", summary.FilterMode)
			return
		}
		fmt.Fprintln(out, "No test history yet.")
		return
	}

	title := "All tests"
	if summary.FilterMode != "" {
		title = fmt.Sprintf("Mode: %s", summary.FilterMode)
	}
	fmt.Fprintf(out, "%s\n\n", title)

	rows := [][2]string{
		{"tests", fmt.Sprintf("%d", summary.TotalTests)},
		{"avg wpm", fmt.Sprintf("%.2f", summary.AverageWPM)},
		{"avg acc", fmt.Sprintf("%.2f%%", summary.AverageAccuracy)},
		{"best wpm", fmt.Sprintf("%.2f", summary.BestWPM)},
		{"last 10", fmt.Sprintf("%.2f", summary.RecentAverageWPM)},
	}
	for _, row := range rows {
		fmt.Fprintf(out, "  %-9s %s\n", row[0], row[1])
	}

	if summary.FilterMode == "" && summary.PersonalBest.BestWPM > 0 {
		fmt.Fprintf(out, "\n  %-9s %.2f wpm · %.2f%% accuracy\n",
			"all-time", summary.PersonalBest.BestWPM, summary.PersonalBest.BestAccuracy)
	}
}

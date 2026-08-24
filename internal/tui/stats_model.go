package tui

import (
	"fmt"
	"strings"

	"github.com/alirezaudev/ttype/internal/domain"
)

// RenderStats draws the stats page. It is a plain render rather than a Bubble
// Tea model: the page has nothing to interact with.
func RenderStats(summary domain.StatsSummary, trend []float64, theme Theme, width int) string {
	if summary.TotalTests == 0 {
		if summary.FilterMode != "" {
			return fmt.Sprintf("No tests recorded for mode %q.\n", summary.FilterMode)
		}
		return "No test history yet.\n"
	}

	title := "All tests"
	if summary.FilterMode != "" {
		title = "Mode: " + summary.FilterMode
	}

	rows := [][2]string{
		{"tests", fmt.Sprintf("%d", summary.TotalTests)},
		{"avg wpm", fmt.Sprintf("%.2f", summary.AverageWPM)},
		{"avg acc", fmt.Sprintf("%.2f%%", summary.AverageAccuracy)},
		{"best wpm", fmt.Sprintf("%.2f", summary.BestWPM)},
		{"last 10", fmt.Sprintf("%.2f", summary.RecentAverageWPM)},
	}

	lines := []string{theme.Finished.Render(title), ""}
	for _, row := range rows {
		lines = append(lines, "  "+
			theme.HUDStatLabel(row[0]).Render(fmt.Sprintf("%-9s", row[0]))+
			theme.HUDValue.Render(row[1]))
	}

	if summary.FilterMode == "" && summary.PersonalBest.BestWPM > 0 {
		lines = append(lines, "", "  "+theme.Help.Render("all-time  ")+
			theme.HUDValue.Render(fmt.Sprintf("%.2f wpm · %.2f%% accuracy",
				summary.PersonalBest.BestWPM, summary.PersonalBest.BestAccuracy)))
	}

	if len(trend) > 1 {
		sample := downsampleSeries(trend, max(min(width, 72)-4, 8))
		lines = append(lines, "",
			"  "+theme.Help.Render("trend"),
			"  "+theme.HUDWPM.Render(renderSparkline(sample)))
	}

	return strings.Join(lines, "\n") + "\n"
}

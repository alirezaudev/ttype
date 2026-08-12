package tui

import (
	"fmt"
	"strings"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/storage"
	"github.com/charmbracelet/lipgloss"
)

func resultSubtitle(cfg domain.TestConfig) string {
	if cfg.Kind == domain.TestKindWords {
		return fmt.Sprintf("%d words · words", cfg.WordCount)
	}
	return fmt.Sprintf("%ds · timed", cfg.Duration.Seconds())
}

const (
	resultColWidth = 9

	resultChartMaxWidth  = 64
	resultChartMaxHeight = 12
)

func pbNotice(pb storage.PBUpdate) string {
	if pb.PrevWPM <= 0 {
		return fmt.Sprintf("New personal best for %s!", pb.Label)
	}
	return fmt.Sprintf("New personal best for %s! (was %.2f)", pb.Label, pb.PrevWPM)
}

func renderResult(result domain.Result, pb storage.PBUpdate, theme Theme, width, height int, notice statusNotice) string {
	title := theme.Finished.Render("Test Complete")
	subtitle := theme.Help.Render(resultSubtitle(result.Config))

	var header strings.Builder
	for i, label := range []string{"wpm", "raw", "acc", "con", "err"} {
		if i > 0 {
			header.WriteString(" ")
		}
		header.WriteString(theme.HUDStatLabel(label).Render(fmt.Sprintf("%-*s", resultColWidth, label)))
	}
	headers := header.String()
	values := fmt.Sprintf("%-*s %-*s %-*s %-*s %-*d",
		resultColWidth, fmt.Sprintf("%.2f", result.WPM),
		resultColWidth, fmt.Sprintf("%.2f", result.RawWPM),
		resultColWidth, fmt.Sprintf("%.2f%%", result.Accuracy),
		resultColWidth, fmt.Sprintf("%.2f%%", result.Consistency),
		resultColWidth, result.Incorrect,
	)

	elapsed := theme.Help.Render(formatClock(result.Duration) + " elapsed")
	help := theme.Help.Render("tab/enter restart  C copy  S settings  L language")

	head := []string{
		title,
		"",
		subtitle,
		"",
		headers,
		values,
		"",
		elapsed,
	}

	var tail []string
	if pb.IsNew {
		tail = append(tail, "", theme.Finished.Render(pbNotice(pb)))
	}
	if !notice.empty() {
		tail = append(tail, "", notice.render(theme))
	}
	tail = append(tail, "", help)

	lines := head
	chart := renderResultChart(
		result.WPMHistory,
		result.RawWPMHistory,
		result.ErrorHistory,
		theme,
		min(width-4, resultChartMaxWidth),
		min(height-len(head)-len(tail)-1, resultChartMaxHeight),
	)
	if chart != "" {
		lines = append(lines, "", chart)
	}
	lines = append(lines, tail...)

	content := strings.Join(lines, "\n")

	if width == 0 || height == 0 {
		return content
	}
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}

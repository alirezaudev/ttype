package tui

import (
	"fmt"
	"strings"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/charmbracelet/lipgloss"
)

func resultSubtitle(cfg domain.TestConfig) string {
	if cfg.Kind == domain.TestKindWords {
		return fmt.Sprintf("%d words · words", cfg.WordCount)
	}
	return fmt.Sprintf("%ds · timed", cfg.Duration.Seconds())
}

const resultColWidth = 9

func renderResult(result domain.Result, theme Theme, width, height int, notice statusNotice) string {
	title := theme.Finished.Render("Test Complete")
	subtitle := theme.Help.Render(resultSubtitle(result.Config))

	var header strings.Builder
	for i, label := range []string{"wpm", "raw", "acc", "err"} {
		if i > 0 {
			header.WriteString(" ")
		}
		header.WriteString(theme.HUDStatLabel(label).Render(fmt.Sprintf("%-*s", resultColWidth, label)))
	}
	headers := header.String()
	values := fmt.Sprintf("%-*s %-*s %-*s %-*d",
		resultColWidth, fmt.Sprintf("%.2f", result.WPM),
		resultColWidth, fmt.Sprintf("%.2f", result.RawWPM),
		resultColWidth, fmt.Sprintf("%.2f%%", result.Accuracy),
		resultColWidth, result.Incorrect,
	)

	elapsed := theme.Help.Render(formatClock(result.Duration) + " elapsed")
	help := theme.Help.Render("tab/enter restart  C copy  S settings  L language")

	lines := []string{
		title,
		"",
		subtitle,
		"",
		headers,
		values,
		"",
		elapsed,
	}
	if !notice.empty() {
		lines = append(lines, "", notice.render(theme))
	}
	lines = append(lines, "", help)

	content := strings.Join(lines, "\n")

	if width == 0 || height == 0 {
		return content
	}
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}

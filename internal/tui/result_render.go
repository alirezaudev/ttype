package tui

import (
	"fmt"
	"strings"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/stats"
	"github.com/alirezaudev/ttype/internal/storage"
	"github.com/charmbracelet/lipgloss"
)

const (
	// Below this the hero stacks above the chart instead of sitting beside it.
	resultSideBySideMinWidth = 70

	resultChartMaxWidth  = 64
	resultChartMaxHeight = 12
)

func resultSubtitle(cfg domain.TestConfig) string {
	if cfg.IsWordsMode() {
		return fmt.Sprintf("%d words · %s", cfg.WordCount, cfg.TextMode)
	}
	return fmt.Sprintf("%ds · %s", cfg.Duration.Seconds(), cfg.TextMode)
}

func pbNotice(pb storage.PBUpdate) string {
	if pb.PrevWPM <= 0 {
		return fmt.Sprintf("New personal best for %s!", pb.Label)
	}
	return fmt.Sprintf("New personal best for %s! (was %.2f)", pb.Label, pb.PrevWPM)
}

func renderResult(result domain.Result, pb storage.PBUpdate, theme Theme, width, height int, ver domain.VersionInfo, notice statusNotice) string {
	titleText := "Test Complete"
	title := theme.Finished.Render(titleText)
	if result.Failed {
		title = theme.Incorrect.Render("Test Failed")
	}

	head := []string{title, theme.Help.Render(resultSubtitle(result.Config))}
	if result.Failed && result.FailureReason != "" {
		head = append(head, theme.Incorrect.Render(result.FailureReason))
	}
	if pb.IsNew {
		head = append(head, theme.HUDValue.Render(pbNotice(pb)))
	}
	if !notice.empty() {
		head = append(head, notice.render(theme))
	}

	tail := []string{"", renderStatStrip(resultStats(result, theme), theme, max(width-4, 20))}
	if heatmap := renderCharHeatmap(result.CharErrors, theme, 8); heatmap != "" {
		tail = append(tail, "", heatmap)
	}
	tail = append(tail, "", theme.Help.Render(resultsHelpLine()))

	footer := renderFooterSection(theme, result.Config, ver, width)
	budget := height - len(head) - len(tail) - 2
	if footer != "" {
		budget -= lipgloss.Height(footer)
	}

	parts := append([]string{}, head...)
	parts = append(parts, "", renderResultCenterpiece(result, theme, width, budget))
	parts = append(parts, tail...)

	return composeWithBottomFooter(lipgloss.JoinVertical(lipgloss.Center, parts...), footer, width, height)
}

func renderResultCenterpiece(result domain.Result, theme Theme, width, budget int) string {
	sideBySide := width >= resultSideBySideMinWidth
	hero := renderResultHero(result, theme, sideBySide)

	if sideBySide {
		chart := renderResultChart(
			result.WPMHistory, result.RawWPMHistory, result.ErrorHistory, theme,
			min(width-lipgloss.Width(hero)-8, 72),
			max(min(budget, 16), lipgloss.Height(hero)),
		)
		if chart == "" {
			return hero
		}
		return lipgloss.JoinHorizontal(lipgloss.Center, hero, "    ", chart)
	}

	chart := renderResultChart(
		result.WPMHistory, result.RawWPMHistory, result.ErrorHistory, theme,
		min(width-4, resultChartMaxWidth),
		min(budget-lipgloss.Height(hero)-1, resultChartMaxHeight),
	)
	if chart == "" {
		return hero
	}
	return lipgloss.JoinVertical(lipgloss.Center, hero, "", chart)
}

// The two numbers that matter, stacked when they sit next to the chart.
func renderResultHero(result domain.Result, theme Theme, stacked bool) string {
	wpmValue := fmt.Sprintf("%.0f", result.WPM)
	accValue := fmt.Sprintf("%.0f%%", result.Accuracy)

	if !unicodeCapable() {
		return theme.Help.Render("wpm ") + theme.HUDWPM.Render(wpmValue) +
			theme.Help.Render("   acc ") + theme.HUDAcc.Render(accValue)
	}

	wpm := lipgloss.JoinVertical(lipgloss.Left,
		theme.Help.Render("wpm"), renderBigDigits(wpmValue, theme.HUDWPM))
	acc := lipgloss.JoinVertical(lipgloss.Left,
		theme.Help.Render("acc"), renderBigDigits(accValue, theme.HUDAcc))

	if stacked {
		return lipgloss.JoinVertical(lipgloss.Left, wpm, "", acc)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, wpm, "   ", acc)
}

func resultStats(result domain.Result, theme Theme) []string {
	seg := func(label, value string) string {
		return theme.Help.Render(label+" ") + theme.HUDValue.Render(value)
	}

	extra := result.TotalChars - result.Correct - result.Incorrect
	if extra < 0 {
		extra = 0
	}

	segments := []string{
		seg("raw", fmt.Sprintf("%.0f", result.RawWPM)),
		seg("consistency", fmt.Sprintf("%.0f%%", result.Consistency)),
		seg("errors", fmt.Sprintf("%d", keystrokeErrors(result))),
		seg("chars", fmt.Sprintf("%d/%d/%d/%d", result.Correct, result.Incorrect, extra, result.Skipped)),
		seg("time", formatClock(result.Duration)),
	}
	if result.Seed != 0 {
		segments = append(segments, seg("seed", fmt.Sprintf("%d", result.Seed)))
	}
	return segments
}

// renderStatStrip joins the segments with dots, wrapping when one would push
// the line past maxWidth.
func renderStatStrip(segments []string, theme Theme, maxWidth int) string {
	sep := theme.Help.Render(" · ")

	var lines []string
	line := ""
	for _, segment := range segments {
		if line == "" {
			line = segment
			continue
		}
		if candidate := line + sep + segment; lipgloss.Width(candidate) <= maxWidth {
			line = candidate
			continue
		}
		lines = append(lines, line)
		line = segment
	}
	if line != "" {
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

// renderCharHeatmap lists the characters that tripped you up most.
func renderCharHeatmap(charErrors map[string]int, theme Theme, limit int) string {
	top := stats.TopCharErrors(charErrors, limit)
	if len(top) == 0 {
		return ""
	}

	parts := make([]string, 0, len(top))
	for _, e := range top {
		char := e.Char
		if char == " " {
			char = "space"
		}
		parts = append(parts, fmt.Sprintf("%s×%d", char, e.Count))
	}
	return theme.HUD.Render("missed ") + theme.Help.Render(strings.Join(parts, "  "))
}

// Results saved before keystroke counting carry zeros, so fall back to the
// buffer count for those.
func keystrokeErrors(result domain.Result) int {
	if result.KeystrokesCorrect+result.KeystrokesIncorrect > 0 {
		return result.KeystrokesIncorrect
	}
	return result.Incorrect
}

package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/alirezaudev/ttype/internal/engine"
	"github.com/charmbracelet/lipgloss"
)

var (
	resultTitleStyle = lipgloss.NewStyle().Bold(true)
	resultLabelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	resultHelpStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

type resultSnapshot struct {
	wpm      float64
	rawWpm   float64
	accuracy float64
	errors   int
	elapsed  time.Duration
	cfg      engine.Config
}

func snapshotResult(s *engine.Session, cfg engine.Config) resultSnapshot {
	return resultSnapshot{
		wpm:      s.WPM(),
		rawWpm:   s.RawWPM(),
		accuracy: s.Accuracy(),
		errors:   s.Incorrect(),
		elapsed:  s.Elapsed(),
		cfg:      cfg,
	}
}

func (s resultSnapshot) subtitle() string {
	if s.cfg.Kind == engine.TestKindWords {
		return fmt.Sprintf("%d words · words", s.cfg.WordCount)
	}
	return fmt.Sprintf("%ds · timed", int(s.cfg.Duration/time.Second))
}

const resultColWidth = 9

func renderResult(snap resultSnapshot, width, height int) string {
	title := resultTitleStyle.Render("Test Complete")
	subtitle := resultLabelStyle.Render(snap.subtitle())

	headers := resultLabelStyle.Render(fmt.Sprintf("%-*s %-*s %-*s %-*s",
		resultColWidth, "wpm",
		resultColWidth, "raw",
		resultColWidth, "acc",
		resultColWidth, "err",
	))
	values := fmt.Sprintf("%-*s %-*s %-*s %-*d",
		resultColWidth, fmt.Sprintf("%.2f", snap.wpm),
		resultColWidth, fmt.Sprintf("%.2f", snap.rawWpm),
		resultColWidth, fmt.Sprintf("%.2f%%", snap.accuracy),
		resultColWidth, snap.errors,
	)

	elapsed := resultLabelStyle.Render(formatClock(snap.elapsed) + " elapsed")
	help := resultHelpStyle.Render("tab/enter restart  ctrl+s settings")

	content := strings.Join([]string{
		title,
		"",
		subtitle,
		"",
		headers,
		values,
		"",
		elapsed,
		"",
		help,
	}, "\n")

	if width == 0 || height == 0 {
		return content
	}
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}

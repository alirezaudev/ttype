package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/engine"
	"github.com/charmbracelet/lipgloss"
)

type resultSnapshot struct {
	wpm      float64
	rawWpm   float64
	accuracy float64
	errors   int
	elapsed  time.Duration
	cfg      domain.TestConfig
}

func snapshotResult(s *engine.Session, cfg domain.TestConfig) resultSnapshot {
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
	if s.cfg.Kind == domain.TestKindWords {
		return fmt.Sprintf("%d words · words", s.cfg.WordCount)
	}
	return fmt.Sprintf("%ds · timed", s.cfg.Duration.Seconds())
}

const resultColWidth = 9

func renderResult(snap resultSnapshot, theme Theme, width, height int) string {
	title := theme.Finished.Render("Test Complete")
	subtitle := theme.Help.Render(snap.subtitle())

	headers := theme.Help.Render(fmt.Sprintf("%-*s %-*s %-*s %-*s",
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

	elapsed := theme.Help.Render(formatClock(snap.elapsed) + " elapsed")
	help := theme.Help.Render("tab/enter restart  ctrl+s settings")

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

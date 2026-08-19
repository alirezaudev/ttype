package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/engine"
	"github.com/charmbracelet/lipgloss"
)

func renderHUD(session *engine.Session, theme Theme, width int, cfg domain.TestConfig, hideLive bool) string {
	if cfg.Zen {
		return ""
	}

	live := session.LiveStats()
	row := theme.HUDTime.Render(hudTimer(session, cfg, live.Elapsed))

	if !hideLive {
		_, errKeystrokes := session.Keystrokes()
		line := hudLiveLine(live.WPM, live.RawWPM, live.Accuracy, errKeystrokes, cfg.Blind)
		if lipgloss.Width(row)+2+len(line) <= width {
			row += "  " + theme.Help.Render(line)
		}
	}

	brand := theme.HUDTitle.Render("ttype") +
		theme.Help.Render(" · ") +
		theme.HUDMode.Render(string(cfg.TextMode))

	if gap := width - lipgloss.Width(row+brand); gap >= 2 {
		row += strings.Repeat(" ", gap) + brand
	}

	return lipgloss.NewStyle().Width(width).Render(row)
}

// Blind mode shows raw speed only: wpm, accuracy and the error count would
// leak exactly what blind mode hides.
func hudLiveLine(wpm, raw, acc float64, errs int, blind bool) string {
	if blind {
		return fmt.Sprintf("raw %-3s", hudNum(raw))
	}
	return fmt.Sprintf("wpm %-3s · raw %-3s · acc %-4s · err %-3s",
		hudNum(wpm), hudNum(raw), fmt.Sprintf("%.0f%%", acc), hudNum(float64(errs)))
}

func hudNum(v float64) string {
	if v < 0 {
		v = 0
	}
	if v > 999 {
		v = 999
	}
	return fmt.Sprintf("%.0f", v)
}

func hudTimer(session *engine.Session, cfg domain.TestConfig, elapsed time.Duration) string {
	if cfg.IsWordsMode() {
		done, total := session.WordsProgress()
		totalStr := fmt.Sprintf("%d", total)
		return fmt.Sprintf("%*d/%s · %s", len(totalStr), done, totalStr, formatClock(elapsed))
	}

	full := formatClock(time.Duration(cfg.Duration.Seconds()) * time.Second)
	var label string
	switch session.State() {
	case domain.SessionReady:
		label = full
	case domain.SessionActive:
		label = formatClock(session.Remaining())
	default:
		label = formatClock(elapsed)
	}
	return fmt.Sprintf("%-*s", len(full), label)
}

// blindRevealEnd is the start of the word under the cursor: everything before
// it is revealed, the word in progress stays hidden.
func blindRevealEnd(target []rune, cursor int) int {
	if cursor <= 0 {
		return 0
	}
	start := cursor
	for start > 0 && target[start-1] != ' ' {
		start--
	}
	return start
}

func renderCapsWarn(theme Theme) string {
	label := "⚠ Caps Lock?"
	if !unicodeCapable() {
		label = "! Caps Lock?"
	}
	return theme.CapsWarn.Render(" " + label + " ")
}

func unicodeCapable() bool {
	for _, key := range []string{"LC_ALL", "LC_CTYPE", "LANG"} {
		v := os.Getenv(key)
		if v == "" {
			continue
		}
		v = strings.ToUpper(v)
		return strings.Contains(v, "UTF-8") || strings.Contains(v, "UTF8")
	}
	return true
}

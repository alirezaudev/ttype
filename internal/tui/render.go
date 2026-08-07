package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/engine"
)

func renderHUD(session *engine.Session, theme Theme) string {
	live := session.LiveStats()

	status := theme.Help.Render(fmt.Sprintf(
		"wpm %-3d · raw %-3d · acc %-3d%% · err %-3d",
		int(live.WPM),
		int(live.RawWPM),
		int(live.Accuracy),
		live.Incorrect,
	))

	if session.Kind() == domain.TestKindTimed {
		return theme.HUDTime.Render(formatClock(session.Remaining())) + " · " + status + "\n\n"
	}

	done, total := session.WordsProgress()
	totalStr := fmt.Sprintf("%d", total)
	progress := theme.HUD.Render(fmt.Sprintf("%*d/%s", len(totalStr), done, totalStr))
	return progress + " · " + status + "\n\n"
}

func renderCapsWarn(theme Theme) string {
	label := "⚠ Caps Lock?"
	if !chartUnicodeCapable() {
		label = "! Caps Lock?"
	}
	return theme.CapsWarn.Render(" " + label + " ")
}

func chartUnicodeCapable() bool {
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

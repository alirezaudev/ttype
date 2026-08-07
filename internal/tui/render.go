package tui

import (
	"fmt"

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

package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/alirezaudev/ttype/internal/domain"
)

// The table sizes itself from the number of rows, and it used to forget that
// the header takes rows too, which pushed the oldest run out of view.
func TestHistoryShowsEveryRun(t *testing.T) {
	t.Parallel()

	var results []domain.Result
	for i := 0; i < 4; i++ {
		results = append(results, domain.Result{
			ID:        fmt.Sprintf("r%d", i),
			Timestamp: time.Date(2026, 9, 10, 20, i, 0, 0, time.UTC),
			Config:    domain.TestConfig{Kind: domain.TestKindWords, WordCount: 8, TextMode: domain.TextModeWords},
			WPM:       float64(40 + i*11),
		})
	}

	m := NewHistoryModel(nil, results, defaultTheme())
	m.width, m.height = 100, 30
	view := stripANSI(m.View())

	for _, r := range results {
		if want := fmt.Sprintf("%.0f", r.WPM); !strings.Contains(view, want) {
			t.Fatalf("history view is missing the %s wpm run:\n%s", want, view)
		}
	}
}

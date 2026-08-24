package app

import (
	"strings"
	"testing"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/tui"
)

func TestRenderStatsEmpty(t *testing.T) {
	t.Parallel()

	got := tui.RenderStats(domain.StatsSummary{}, nil, tui.DefaultTheme(), 72)
	if got != "No test history yet.\n" {
		t.Fatalf("output = %q", got)
	}
}

func TestRenderStatsEmptyForAMode(t *testing.T) {
	t.Parallel()

	got := tui.RenderStats(domain.StatsSummary{FilterMode: "sql"}, nil, tui.DefaultTheme(), 72)
	if !strings.Contains(got, `mode "sql"`) {
		t.Fatalf("output should name the filtered mode: %q", got)
	}
}

func TestRenderStatsSummary(t *testing.T) {
	t.Parallel()

	out := tui.RenderStats(domain.StatsSummary{
		TotalTests:       3,
		AverageWPM:       61.5,
		AverageAccuracy:  94.25,
		BestWPM:          70,
		RecentAverageWPM: 62,
		PersonalBest:     domain.PersonalBests{BestWPM: 70, BestAccuracy: 99},
	}, []float64{40, 55, 70}, tui.DefaultTheme(), 72)

	for _, want := range []string{"tests", "61.50", "94.25%", "70.00", "all-time", "trend"} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
}

func TestRenderStatsHidesAllTimeWhenFiltered(t *testing.T) {
	t.Parallel()

	out := tui.RenderStats(domain.StatsSummary{
		TotalTests:   2,
		FilterMode:   "sql",
		AverageWPM:   40,
		PersonalBest: domain.PersonalBests{BestWPM: 120},
	}, nil, tui.DefaultTheme(), 72)

	if strings.Contains(out, "all-time") {
		t.Fatalf("filtered stats should not show all-time bests:\n%s", out)
	}
}

package tui

import (
	"strings"
	"testing"

	"github.com/alirezaudev/ttype/internal/domain"
)

func TestStatsTitleNamesTheFilters(t *testing.T) {
	t.Parallel()

	summary := domain.StatsSummary{TotalTests: 1, FilterMode: "custom", FilterTag: "english"}
	if got := stripANSI(RenderStats(summary, nil, defaultTheme(), 72)); !strings.HasPrefix(got, "Mode: custom · Tag: english") {
		t.Fatalf("stats = %q", got)
	}

	summary.TotalTests = 0
	if got := RenderStats(summary, nil, defaultTheme(), 72); got != "No tests recorded for mode \"custom\" and tag \"english\".\n" {
		t.Fatalf("empty stats = %q", got)
	}
}

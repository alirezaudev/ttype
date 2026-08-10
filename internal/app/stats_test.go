package app

import (
	"bytes"
	"strings"
	"testing"

	"github.com/alirezaudev/ttype/internal/domain"
)

func TestPrintStatsEmpty(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	printStats(&buf, domain.StatsSummary{})

	if got := buf.String(); got != "No test history yet.\n" {
		t.Fatalf("output = %q", got)
	}
}

func TestPrintStatsEmptyForAMode(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	printStats(&buf, domain.StatsSummary{FilterMode: "sql"})

	if !strings.Contains(buf.String(), `mode "sql"`) {
		t.Fatalf("output should name the filtered mode: %q", buf.String())
	}
}

func TestPrintStatsSummary(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	printStats(&buf, domain.StatsSummary{
		TotalTests:       3,
		AverageWPM:       61.5,
		AverageAccuracy:  94.25,
		BestWPM:          70,
		RecentAverageWPM: 62,
		PersonalBest:     domain.PersonalBests{BestWPM: 70, BestAccuracy: 99},
	})

	out := buf.String()
	for _, want := range []string{"tests", "61.50", "94.25%", "70.00", "all-time"} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
}

func TestPrintStatsHidesAllTimeWhenFiltered(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	printStats(&buf, domain.StatsSummary{
		TotalTests:   2,
		FilterMode:   "sql",
		AverageWPM:   40,
		PersonalBest: domain.PersonalBests{BestWPM: 120},
	})

	if strings.Contains(buf.String(), "all-time") {
		t.Fatalf("filtered stats should not show all-time bests:\n%s", buf.String())
	}
}

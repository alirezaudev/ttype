package storage

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/alirezaudev/ttype/internal/domain"
)

func saveRuns(t *testing.T, s *JSONStore, runs ...domain.Result) {
	t.Helper()
	for _, run := range runs {
		if _, err := s.SaveResult(run); err != nil {
			t.Fatalf("SaveResult: %v", err)
		}
	}
}

func run(id string, wpm float64, mode domain.TextMode, failed bool) domain.Result {
	return domain.Result{
		ID:                  id,
		Timestamp:           time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC),
		Config:              domain.TestConfig{Kind: domain.TestKindTimed, Duration: 60, TextMode: mode},
		WPM:                 wpm,
		RawWPM:              wpm + 10,
		Accuracy:            95,
		KeystrokesIncorrect: 3,
		Failed:              failed,
	}
}

func TestExcludeFailedDropsFailedRuns(t *testing.T) {
	t.Parallel()

	s := testStore(t)
	saveRuns(t, s,
		run("a", 60, domain.TextModeWords, false),
		run("b", 20, domain.TextModeWords, true),
	)

	summary, err := s.Summary(domain.StatsFilter{ExcludeFailed: true})
	if err != nil {
		t.Fatalf("Summary: %v", err)
	}
	if summary.TotalTests != 1 || summary.AverageWPM != 60 {
		t.Fatalf("summary = %+v, want only the passing run", summary)
	}
}

func TestTrendFollowsTheFilter(t *testing.T) {
	t.Parallel()

	s := testStore(t)
	saveRuns(t, s,
		run("a", 50, domain.TextModeWords, false),
		run("b", 70, domain.TextModeGo, false),
		run("c", 60, domain.TextModeWords, false),
	)

	mode := domain.TextModeWords
	trend, err := s.Trend(domain.StatsFilter{TextMode: &mode})
	if err != nil {
		t.Fatalf("Trend: %v", err)
	}
	if len(trend) != 2 || trend[0] != 50 || trend[1] != 60 {
		t.Fatalf("trend = %v, want [50 60]", trend)
	}
}

func TestExportCSVWritesEveryRun(t *testing.T) {
	t.Parallel()

	s := testStore(t)
	saveRuns(t, s, run("a", 50, domain.TextModeWords, false), run("b", 70, domain.TextModeGo, true))

	var buf bytes.Buffer
	if err := s.ExportCSV(&buf, domain.StatsFilter{}); err != nil {
		t.Fatalf("ExportCSV: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("csv has %d lines, want header + 2 rows:\n%s", len(lines), buf.String())
	}
	if !strings.HasPrefix(lines[0], "timestamp,mode,language") {
		t.Fatalf("header = %q", lines[0])
	}
	if !strings.Contains(lines[1], "50.00") || !strings.Contains(lines[2], "true") {
		t.Fatalf("rows = %q / %q", lines[1], lines[2])
	}
}

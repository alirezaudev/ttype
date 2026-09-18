package storage

import (
	"testing"

	"github.com/alirezaudev/ttype/internal/domain"
)

func TestStatsFilterByTag(t *testing.T) {
	t.Parallel()

	s := testStore(t)
	for i, tag := range []string{"english", "", "english"} {
		r := domain.Result{
			ID:     string(rune('a' + i)),
			Config: domain.TestConfig{Kind: domain.TestKindTimed, Duration: domain.Duration60, Tag: tag},
			WPM:    float64(40 + 10*i),
		}
		if _, err := s.SaveResult(r); err != nil {
			t.Fatalf("SaveResult: %v", err)
		}
	}

	summary, err := s.Summary(domain.StatsFilter{Tag: "english"})
	if err != nil {
		t.Fatalf("Summary: %v", err)
	}
	if summary.TotalTests != 2 || summary.AverageWPM != 50 || summary.FilterTag != "english" {
		t.Fatalf("summary = %+v, want 2 english runs averaging 50", summary)
	}

	results, err := s.ListResults(0)
	if err != nil {
		t.Fatalf("ListResults: %v", err)
	}
	if results[0].Config.Tag != "english" || results[1].Config.Tag != "" {
		t.Fatalf("tags = %q, %q", results[0].Config.Tag, results[1].Config.Tag)
	}
}

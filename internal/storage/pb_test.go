package storage

import (
	"testing"
	"time"

	"github.com/alirezaudev/ttype/internal/domain"
)

func TestConfigLabel(t *testing.T) {
	t.Parallel()

	timed := domain.TestConfig{
		Kind:     domain.TestKindTimed,
		Duration: domain.Duration60,
		TextMode: domain.TextModeWords,
	}

	cases := []struct {
		name string
		cfg  domain.TestConfig
		want string
	}{
		{
			name: "plain timed run keeps the bare label",
			cfg:  timed,
			want: "words/60s",
		},
		{
			name: "words mode counts words, not seconds",
			cfg: domain.TestConfig{
				Kind:      domain.TestKindWords,
				WordCount: 25,
				TextMode:  domain.TextModeWords,
			},
			want: "words/25w",
		},
		{
			name: "mode is part of the bucket",
			cfg: domain.TestConfig{
				Kind:     domain.TestKindTimed,
				Duration: domain.Duration60,
				TextMode: domain.TextModeSQL,
			},
			want: "sql/60s",
		},
		{
			name: "dimensions append in a fixed order",
			cfg: domain.TestConfig{
				Kind:        domain.TestKindTimed,
				Duration:    domain.Duration60,
				TextMode:    domain.TextModeWords,
				Language:    "spanish",
				Punctuation: true,
				Numbers:     true,
			},
			want: "words/60s/spanish/punct/num",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := configLabel(tc.cfg); got != tc.want {
				t.Fatalf("configLabel = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestPersonalBestsSplitByConfig(t *testing.T) {
	t.Parallel()

	s := testStore(t)

	plain := domain.TestConfig{
		Kind:     domain.TestKindTimed,
		Duration: domain.Duration60,
		TextMode: domain.TextModeWords,
	}
	punctuated := plain
	punctuated.Punctuation = true

	if _, err := s.SaveResult(domain.Result{ID: "a", Timestamp: time.Now().UTC(), Config: plain, WPM: 90}); err != nil {
		t.Fatalf("SaveResult: %v", err)
	}

	pb, err := s.SaveResult(domain.Result{ID: "b", Timestamp: time.Now().UTC(), Config: punctuated, WPM: 60})
	if err != nil {
		t.Fatalf("SaveResult: %v", err)
	}
	if !pb.IsNew {
		t.Fatal("a punctuation run must not compete against plain runs")
	}
	if pb.Label != "words/60s/punct" {
		t.Fatalf("label = %q, want %q", pb.Label, "words/60s/punct")
	}
	if pb.PrevWPM != 0 {
		t.Fatalf("prev = %.2f, want 0 for an empty bucket", pb.PrevWPM)
	}

	pb, err = s.SaveResult(domain.Result{ID: "c", Timestamp: time.Now().UTC(), Config: punctuated, WPM: 55})
	if err != nil {
		t.Fatalf("SaveResult: %v", err)
	}
	if pb.IsNew {
		t.Fatal("55 wpm should not beat 60 wpm in the same bucket")
	}
	if pb.PrevWPM != 60 {
		t.Fatalf("prev = %.2f, want 60", pb.PrevWPM)
	}
}

func TestBestsFileTracksAllTimeRecords(t *testing.T) {
	t.Parallel()

	s := testStore(t)

	bests, err := s.LoadBests()
	if err != nil {
		t.Fatalf("LoadBests: %v", err)
	}
	if bests.BestWPM != 0 {
		t.Fatal("a missing bests file should read as zero, not error")
	}

	cfg := domain.TestConfig{Kind: domain.TestKindTimed, Duration: domain.Duration60}
	if _, err := s.SaveResult(domain.Result{ID: "a", Config: cfg, WPM: 70, Accuracy: 91}); err != nil {
		t.Fatalf("SaveResult: %v", err)
	}
	if _, err := s.SaveResult(domain.Result{ID: "b", Config: cfg, WPM: 65, Accuracy: 99}); err != nil {
		t.Fatalf("SaveResult: %v", err)
	}

	bests, err = s.LoadBests()
	if err != nil {
		t.Fatalf("LoadBests: %v", err)
	}
	if bests.BestWPM != 70 {
		t.Fatalf("best wpm = %.2f, want 70", bests.BestWPM)
	}
	if bests.BestAccuracy != 99 {
		t.Fatalf("best accuracy = %.2f, want 99", bests.BestAccuracy)
	}
}

// Custom text can be anything, so it can't set a record for real tests.
func TestCustomRunsAreNotPersonalBests(t *testing.T) {
	t.Parallel()

	s := testStore(t)
	custom := domain.TestConfig{Kind: domain.TestKindWords, WordCount: 3, TextMode: domain.TextModeCustom}
	words := domain.TestConfig{Kind: domain.TestKindWords, WordCount: 10, TextMode: domain.TextModeWords}

	pb, err := s.SaveResult(domain.Result{ID: "a", Timestamp: time.Now().UTC(), Config: custom, WPM: 150, Accuracy: 100})
	if err != nil {
		t.Fatalf("SaveResult: %v", err)
	}
	if pb.IsNew {
		t.Fatal("a custom run was announced as a personal best")
	}
	if _, err := s.SaveResult(domain.Result{ID: "b", Timestamp: time.Now().UTC(), Config: words, WPM: 60, Accuracy: 95}); err != nil {
		t.Fatalf("SaveResult: %v", err)
	}

	bests, err := s.LoadBests()
	if err != nil {
		t.Fatalf("LoadBests: %v", err)
	}
	if bests.BestWPM != 60 {
		t.Fatalf("all-time best = %.2f, want 60 from the word test", bests.BestWPM)
	}

	all, err := s.Summary(domain.StatsFilter{})
	if err != nil {
		t.Fatalf("Summary: %v", err)
	}
	if all.TotalTests != 2 || all.BestWPM != 60 {
		t.Fatalf("summary = %+v, want both runs counted and a best of 60", all)
	}
	mode := domain.TextModeCustom
	onlyCustom, err := s.Summary(domain.StatsFilter{TextMode: &mode})
	if err != nil {
		t.Fatalf("Summary: %v", err)
	}
	if onlyCustom.TotalTests != 1 || onlyCustom.BestWPM != 150 {
		t.Fatalf("custom summary = %+v, want the custom run and its 150", onlyCustom)
	}
}

// A run that ended because it dropped under --min-wpm is a failure, and
// failures were being announced as personal bests on the result screen.
func TestFailedRunsAreNotPersonalBests(t *testing.T) {
	t.Parallel()

	s := testStore(t)
	cfg := domain.TestConfig{Kind: domain.TestKindWords, WordCount: 12, TextMode: domain.TextModeWords, MinWPM: 60}

	pb, err := s.SaveResult(domain.Result{
		ID: "a", Timestamp: time.Now().UTC(), Config: cfg, WPM: 55, Accuracy: 100,
		Failed: true, FailureReason: "WPM below minimum (60)",
	})
	if err != nil {
		t.Fatalf("SaveResult: %v", err)
	}
	if pb.IsNew {
		t.Fatal("a failed run was announced as a personal best")
	}

	bests, err := s.LoadBests()
	if err != nil {
		t.Fatalf("LoadBests: %v", err)
	}
	if bests.BestWPM != 0 {
		t.Fatalf("all-time best = %.2f, want 0 — the only run failed", bests.BestWPM)
	}

	pb, err = s.SaveResult(domain.Result{ID: "b", Timestamp: time.Now().UTC(), Config: cfg, WPM: 40})
	if err != nil {
		t.Fatalf("SaveResult: %v", err)
	}
	if !pb.IsNew || pb.PrevWPM != 0 {
		t.Fatalf("pb = %+v, want a first best — the 55 wpm run failed", pb)
	}
}

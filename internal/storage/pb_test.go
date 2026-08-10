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

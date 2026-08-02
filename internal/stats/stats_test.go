package stats_test

import (
	"testing"
	"time"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/stats"
)

func TestWPM(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		correct int
		elapsed time.Duration
		want    float64
	}{
		{
			name:    "standard 60 wpm for one minute",
			correct: 300,
			elapsed: time.Minute,
			want:    60,
		},
		{
			name:    "half minute doubles wpm",
			correct: 300,
			elapsed: 30 * time.Second,
			want:    120,
		},
		{
			name:    "zero elapsed returns zero",
			correct: 100,
			elapsed: 0,
			want:    0,
		},
		{
			name:    "partial word rounds to two decimals",
			correct: 37,
			elapsed: 10 * time.Second,
			want:    44.4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := stats.WPM(tt.correct, tt.elapsed)
			if got != tt.want {
				t.Fatalf("WPM() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRawWPM(t *testing.T) {
	t.Parallel()

	counts := domain.CharCounts{Correct: 250, Incorrect: 50, Extra: 0}
	got := stats.RawWPM(counts, time.Minute)
	if got != 60 {
		t.Fatalf("RawWPM() = %v, want 60", got)
	}
}

func TestAccuracy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		correct   int
		incorrect int
		want      float64
	}{
		{name: "perfect", correct: 100, incorrect: 0, want: 100},
		{name: "half", correct: 50, incorrect: 50, want: 50},
		{name: "empty input", correct: 0, incorrect: 0, want: 100},
		{name: "mostly correct", correct: 97, incorrect: 3, want: 97},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := stats.Accuracy(tt.correct, tt.incorrect)
			if got != tt.want {
				t.Fatalf("Accuracy() = %v, want %v", got, tt.want)
			}
		})
	}
}

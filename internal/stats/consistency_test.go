package stats_test

import (
	"math"
	"testing"

	"github.com/alirezaudev/ttype/internal/stats"
)

func curveExpected(cov float64) float64 {
	c2 := cov * cov
	raw := 100 * (1 - math.Tanh(cov*(1+c2*(1.0/3+c2/5))))
	return math.Round(raw*100) / 100
}

func TestConsistency(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		samples []float64
		want    float64
	}{
		{name: "empty", samples: nil, want: 100},
		{name: "zero mean", samples: []float64{0, 0, 0}, want: 100},
		{name: "flat", samples: []float64{60, 60, 60, 60}, want: 100},
		{name: "single sample", samples: []float64{72}, want: 100},
		{name: "cov 0.25", samples: []float64{75, 45}, want: curveExpected(0.25)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := stats.Consistency(tt.samples); got != tt.want {
				t.Fatalf("Consistency(%v) = %v, want %v", tt.samples, got, tt.want)
			}
		})
	}
}

func TestConsistencyFallsAsSpreadGrows(t *testing.T) {
	t.Parallel()

	prev := 101.0
	for _, spread := range []float64{0, 5, 10, 20, 40} {
		got := stats.Consistency([]float64{60 - spread, 60 + spread})
		if got > prev {
			t.Fatalf("spread %v scored %v, higher than the previous %v", spread, got, prev)
		}
		if got < 0 || got > 100 {
			t.Fatalf("score %v out of range", got)
		}
		prev = got
	}
}

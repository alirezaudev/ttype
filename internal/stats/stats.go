package stats

import (
	"math"
	"sort"
	"time"

	"github.com/alirezaudev/ttype/internal/domain"
)

const charsPerWord = 5.0

func WPM(correctChars int, elapsed time.Duration) float64 {
	if elapsed <= 0 {
		return 0
	}
	seconds := elapsed.Seconds()
	return round2((float64(correctChars) / charsPerWord) * (60.0 / seconds))
}

func RawWPM(counts domain.CharCounts, elapsed time.Duration) float64 {
	if elapsed <= 0 {
		return 0
	}
	seconds := elapsed.Seconds()
	typed := float64(counts.Correct + counts.Incorrect + counts.Extra)
	return round2((typed / charsPerWord) * (60.0 / seconds))
}

func Accuracy(correct, incorrect int) float64 {
	total := correct + incorrect
	if total == 0 {
		return 100
	}
	return round2(float64(correct) / float64(total) * 100)
}

func Live(counts domain.CharCounts, elapsed time.Duration) domain.LiveStats {
	return domain.LiveStats{
		WPM:       WPM(counts.Correct, elapsed),
		RawWPM:    RawWPM(counts, elapsed),
		Accuracy:  Accuracy(counts.Correct, counts.Incorrect),
		Correct:   counts.Correct,
		Incorrect: counts.Incorrect,
		Elapsed:   elapsed,
	}
}

func Consistency(rawSamples []float64) float64 {
	if len(rawSamples) == 0 {
		return 100
	}

	var sum float64
	for _, v := range rawSamples {
		sum += v
	}
	mean := sum / float64(len(rawSamples))
	if mean <= 0 {
		return 100
	}

	var variance float64
	for _, v := range rawSamples {
		d := v - mean
		variance += d * d
	}
	stddev := math.Sqrt(variance / float64(len(rawSamples)))

	score := consistencyCurve(stddev / mean)
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return round2(score)
}

func consistencyCurve(cov float64) float64 {
	c2 := cov * cov
	s := cov * (1 + c2*(1.0/3+c2/5))
	return 100 * (1 - math.Tanh(s))
}

type CharError struct {
	Char  string
	Count int
}

// TopCharErrors ranks the most-missed expected characters, ties alphabetical.
func TopCharErrors(counts map[string]int, n int) []CharError {
	if len(counts) == 0 || n <= 0 {
		return nil
	}

	out := make([]CharError, 0, len(counts))
	for char, count := range counts {
		out = append(out, CharError{Char: char, Count: count})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Char < out[j].Char
	})

	if n > len(out) {
		n = len(out)
	}
	return out[:n]
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

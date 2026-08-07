package stats

import (
	"math"
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

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

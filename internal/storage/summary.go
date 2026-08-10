package storage

import (
	"math"

	"github.com/alirezaudev/ttype/internal/domain"
)

const recentTests = 10

func (s *JSONStore) Summary(filter domain.StatsFilter) (domain.StatsSummary, error) {
	history, err := s.loadHistory()
	if err != nil {
		return domain.StatsSummary{}, err
	}
	history = applyFilter(history, filter)

	bests, err := s.LoadBests()
	if err != nil {
		return domain.StatsSummary{}, err
	}

	summary := domain.StatsSummary{
		TotalTests:   len(history),
		PersonalBest: bests,
	}
	if filter.TextMode != nil {
		summary.FilterMode = string(*filter.TextMode)
	}
	if len(history) == 0 {
		return summary, nil
	}

	var wpmSum, accSum float64
	for _, item := range history {
		wpmSum += item.WPM
		accSum += item.Accuracy
		if item.WPM > summary.BestWPM {
			summary.BestWPM = item.WPM
		}
	}
	n := float64(len(history))
	summary.AverageWPM = round2(wpmSum / n)
	summary.AverageAccuracy = round2(accSum / n)

	recent := history
	if len(recent) > recentTests {
		recent = recent[len(recent)-recentTests:]
	}
	var recentSum float64
	for _, item := range recent {
		recentSum += item.WPM
	}
	summary.RecentAverageWPM = round2(recentSum / float64(len(recent)))

	return summary, nil
}

func applyFilter(history []storedResult, filter domain.StatsFilter) []storedResult {
	if filter.TextMode == nil {
		return history
	}

	out := make([]storedResult, 0, len(history))
	for _, item := range history {
		if domain.TextMode(item.TextMode) == *filter.TextMode {
			out = append(out, item)
		}
	}
	return out
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

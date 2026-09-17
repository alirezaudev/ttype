package storage

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"strconv"

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

	onlyCustom := filter.TextMode != nil && *filter.TextMode == domain.TextModeCustom
	var wpmSum, accSum float64
	for _, item := range history {
		wpmSum += item.WPM
		accSum += item.Accuracy
		if domain.TextMode(item.TextMode) == domain.TextModeCustom && !onlyCustom {
			continue
		}
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
	if filter.TextMode == nil && !filter.ExcludeFailed {
		return history
	}

	out := make([]storedResult, 0, len(history))
	for _, item := range history {
		if filter.TextMode != nil && domain.TextMode(item.TextMode) != *filter.TextMode {
			continue
		}
		if filter.ExcludeFailed && item.Failed {
			continue
		}
		out = append(out, item)
	}
	return out
}

// Trend returns the wpm of each matching run, oldest first, for a sparkline.
func (s *JSONStore) Trend(filter domain.StatsFilter) ([]float64, error) {
	history, err := s.loadHistory()
	if err != nil {
		return nil, err
	}

	history = applyFilter(history, filter)
	out := make([]float64, 0, len(history))
	for _, item := range history {
		out = append(out, item.WPM)
	}
	return out, nil
}

// ExportCSV writes the matching history as CSV, oldest first.
func (s *JSONStore) ExportCSV(w io.Writer, filter domain.StatsFilter) error {
	history, err := s.loadHistory()
	if err != nil {
		return err
	}
	history = applyFilter(history, filter)

	out := csv.NewWriter(w)
	if err := out.Write([]string{
		"timestamp", "mode", "language", "test", "wpm", "raw_wpm",
		"accuracy", "consistency", "errors", "failed",
	}); err != nil {
		return err
	}

	for _, item := range history {
		test := fmt.Sprintf("%ds", item.Duration)
		if item.WordCount > 0 {
			test = fmt.Sprintf("%dw", item.WordCount)
		}
		if err := out.Write([]string{
			item.Timestamp,
			item.TextMode,
			item.Language,
			test,
			strconv.FormatFloat(item.WPM, 'f', 2, 64),
			strconv.FormatFloat(item.RawWPM, 'f', 2, 64),
			strconv.FormatFloat(item.Accuracy, 'f', 2, 64),
			strconv.FormatFloat(item.Consistency, 'f', 2, 64),
			strconv.Itoa(item.KeystrokesIncorrect),
			strconv.FormatBool(item.Failed),
		}); err != nil {
			return err
		}
	}

	out.Flush()
	return out.Error()
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

package storage

import (
	"fmt"

	"github.com/alirezaudev/ttype/internal/domain"
)

type PBUpdate struct {
	IsNew   bool
	PrevWPM float64
	NewWPM  float64
	Label   string
}

func configLabel(cfg domain.TestConfig) string {
	var label string
	if cfg.IsWordsMode() {
		label = fmt.Sprintf("%s/%dw", cfg.TextMode, cfg.WordCount)
	} else {
		label = fmt.Sprintf("%s/%ds", cfg.TextMode, cfg.Duration.Seconds())
	}

	if cfg.Language != "" {
		label += "/" + cfg.Language
	}
	if cfg.Punctuation {
		label += "/punct"
	}
	if cfg.Numbers {
		label += "/num"
	}
	return label
}

// Files from before per-configuration bests get them from history once.
func bestsFromHistory(history []storedResult) map[string]domain.ConfigBest {
	out := make(map[string]domain.ConfigBest)
	for _, item := range history {
		recordConfigBest(out, unmarshalResult(item))
	}
	return out
}

func recordConfigBest(bests map[string]domain.ConfigBest, result domain.Result) bool {
	if !countsForBests(result) {
		return false
	}
	label := configLabel(result.Config)
	if result.WPM <= bests[label].WPM {
		return false
	}
	bests[label] = domain.ConfigBest{WPM: result.WPM, Accuracy: result.Accuracy, Date: result.Timestamp.UTC(), ResultID: result.ID}
	return true
}

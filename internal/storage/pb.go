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

func bestWPMForConfig(history []storedResult, cfg domain.TestConfig) float64 {
	label := configLabel(cfg)

	var best float64
	for _, item := range history {
		if configLabel(unmarshalResult(item).Config) != label {
			continue
		}
		if item.WPM > best {
			best = item.WPM
		}
	}
	return best
}

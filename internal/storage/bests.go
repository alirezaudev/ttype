package storage

import (
	"encoding/json"
	"path/filepath"
	"time"

	"github.com/alirezaudev/ttype/internal/domain"
)

const bestsFilename = "bests.json"

func (s *JSONStore) bestsPath() string {
	return filepath.Join(s.dirs.Data, bestsFilename)
}

func (s *JSONStore) LoadBests() (domain.PersonalBests, error) {
	var bests domain.PersonalBests
	if ok, err := s.readJSON(s.bestsPath(), &bests); !ok {
		return domain.PersonalBests{}, err
	}
	return bests, nil
}

// A run that stopped early because it dropped under --min-wpm is kept in
// history but never counts as a best.
// Custom text can be anything, so a run on it is not a record to beat.
func countsForBests(result domain.Result) bool {
	return !result.Failed && result.Config.TextMode != domain.TextModeCustom
}

// seeded means bests were just filled from history and need writing.
func (s *JSONStore) updateBests(bests domain.PersonalBests, result domain.Result, seeded bool) error {
	changed := seeded
	if countsForBests(result) {
		if recordConfigBest(bests.ByConfig, result) {
			changed = true
		}
		if result.WPM > bests.BestWPM {
			bests.BestWPM = result.WPM
			changed = true
		}
		if result.Accuracy > bests.BestAccuracy {
			bests.BestAccuracy = result.Accuracy
			changed = true
		}
	}
	if !changed {
		return nil
	}

	bests.UpdatedAt = time.Now().UTC()
	data, err := json.MarshalIndent(bests, "", "  ")
	if err != nil {
		return err
	}
	return writeFile(s.bestsPath(), data)
}

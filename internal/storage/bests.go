package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/alirezaudev/ttype/internal/domain"
)

func (s *JSONStore) bestsPath() string {
	return filepath.Join(s.dirs.Data, "bests.json")
}

func (s *JSONStore) LoadBests() (domain.PersonalBests, error) {
	data, err := os.ReadFile(s.bestsPath())
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return domain.PersonalBests{}, nil
		}
		return domain.PersonalBests{}, err
	}

	var bests domain.PersonalBests
	if err := json.Unmarshal(data, &bests); err != nil {
		return domain.PersonalBests{}, fmt.Errorf("parse bests: %w", err)
	}
	return bests, nil
}

func (s *JSONStore) updateBests(result domain.Result) error {
	bests, err := s.LoadBests()
	if err != nil {
		return err
	}

	changed := false
	if result.WPM > bests.BestWPM {
		bests.BestWPM = result.WPM
		changed = true
	}
	if result.Accuracy > bests.BestAccuracy {
		bests.BestAccuracy = result.Accuracy
		changed = true
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

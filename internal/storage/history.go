package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/alirezaudev/ttype/internal/domain"
)

const historyFilename = "history.json"

type storedResult struct {
	ID                  string  `json:"id"`
	Timestamp           string  `json:"timestamp"`
	TestKind            string  `json:"test_kind,omitempty"`
	Duration            int     `json:"duration_sec"`
	WordCount           int     `json:"word_count,omitempty"`
	TextMode            string  `json:"text_mode"`
	Language            string  `json:"language,omitempty"`
	Theme               string  `json:"theme"`
	Punctuation         bool    `json:"punctuation,omitempty"`
	Numbers             bool    `json:"numbers,omitempty"`
	Width               int     `json:"width,omitempty"`
	Seed                int64   `json:"seed,omitempty"`
	WPM                 float64 `json:"wpm"`
	RawWPM              float64 `json:"raw_wpm"`
	Accuracy            float64 `json:"accuracy"`
	Consistency         float64 `json:"consistency,omitempty"`
	Correct             int     `json:"correct"`
	Incorrect           int     `json:"incorrect"`
	KeystrokesCorrect   int     `json:"keystrokes_correct,omitempty"`
	KeystrokesIncorrect int     `json:"keystrokes_incorrect,omitempty"`
	Skipped             int     `json:"skipped,omitempty"`
	TotalChars          int     `json:"total_chars"`

	WPMHistory    []float64      `json:"wpm_history,omitempty"`
	RawWPMHistory []float64      `json:"raw_wpm_history,omitempty"`
	ErrorHistory  []int          `json:"error_history,omitempty"`
	CharErrors    map[string]int `json:"char_errors,omitempty"`
}

func marshalResult(r domain.Result) storedResult {
	durSec := r.Config.Duration.Seconds()
	if r.Config.IsWordsMode() {
		durSec = int(r.Duration.Seconds())
	}

	return storedResult{
		ID:                  r.ID,
		Timestamp:           r.Timestamp.UTC().Format(time.RFC3339),
		TestKind:            string(r.Config.Kind),
		Duration:            durSec,
		WordCount:           r.Config.WordCount,
		TextMode:            string(r.Config.TextMode),
		Language:            r.Config.Language,
		Theme:               r.Config.Theme,
		Punctuation:         r.Config.Punctuation,
		Numbers:             r.Config.Numbers,
		Width:               r.Config.Width,
		Seed:                r.Seed,
		WPM:                 r.WPM,
		RawWPM:              r.RawWPM,
		Accuracy:            r.Accuracy,
		Consistency:         r.Consistency,
		Correct:             r.Correct,
		Incorrect:           r.Incorrect,
		KeystrokesCorrect:   r.KeystrokesCorrect,
		KeystrokesIncorrect: r.KeystrokesIncorrect,
		Skipped:             r.Skipped,
		TotalChars:          r.TotalChars,
		WPMHistory:          r.WPMHistory,
		RawWPMHistory:       r.RawWPMHistory,
		ErrorHistory:        r.ErrorHistory,
		CharErrors:          r.CharErrors,
	}
}

func unmarshalResult(s storedResult) domain.Result {
	timestamp, _ := time.Parse(time.RFC3339, s.Timestamp)

	cfg := domain.TestConfig{
		Kind:        domain.TestKind(s.TestKind),
		Duration:    domain.Duration(s.Duration),
		WordCount:   s.WordCount,
		Width:       s.Width,
		Theme:       s.Theme,
		Language:    s.Language,
		TextMode:    domain.TextMode(s.TextMode),
		Punctuation: s.Punctuation,
		Numbers:     s.Numbers,
	}

	elapsed := time.Duration(s.Duration) * time.Second
	if cfg.Kind == "" {
		cfg.Kind = domain.TestKindTimed
	}

	return domain.Result{
		ID:                  s.ID,
		Timestamp:           timestamp,
		Config:              cfg,
		WPM:                 s.WPM,
		RawWPM:              s.RawWPM,
		Accuracy:            s.Accuracy,
		Consistency:         s.Consistency,
		Correct:             s.Correct,
		Incorrect:           s.Incorrect,
		KeystrokesCorrect:   s.KeystrokesCorrect,
		KeystrokesIncorrect: s.KeystrokesIncorrect,
		Skipped:             s.Skipped,
		TotalChars:          s.TotalChars,
		Duration:            elapsed,
		Seed:                s.Seed,
		WPMHistory:          s.WPMHistory,
		RawWPMHistory:       s.RawWPMHistory,
		ErrorHistory:        s.ErrorHistory,
		CharErrors:          s.CharErrors,
	}
}

func (s *JSONStore) SaveResult(result domain.Result) (PBUpdate, error) {
	lock, err := acquireLock(s.dirs.Data)
	if err != nil {
		return PBUpdate{}, fmt.Errorf("acquire lock: %w", err)
	}
	defer lock.release()

	history, err := s.loadHistory()
	if err != nil {
		return PBUpdate{}, err
	}

	update := PBUpdate{
		Label:   configLabel(result.Config),
		NewWPM:  result.WPM,
		PrevWPM: bestWPMForConfig(history, result.Config),
	}
	update.IsNew = result.WPM > update.PrevWPM

	history = append(history, marshalResult(result))
	if limit := s.historyLimit(); len(history) > limit {
		history = history[len(history)-limit:]
	}
	if err := s.saveHistory(history); err != nil {
		return PBUpdate{}, err
	}
	if err := s.updateBests(result); err != nil {
		return PBUpdate{}, err
	}
	return update, nil
}

// ListResults returns the newest results first, at most limit of them.
func (s *JSONStore) ListResults(limit int) ([]domain.Result, error) {
	history, err := s.loadHistory()
	if err != nil {
		return nil, err
	}

	if limit > 0 && len(history) > limit {
		history = history[len(history)-limit:]
	}

	out := make([]domain.Result, 0, len(history))
	for i := len(history) - 1; i >= 0; i-- {
		out = append(out, unmarshalResult(history[i]))
	}
	return out, nil
}

func (s *JSONStore) historyLimit() int {
	if s.historyCap > 0 {
		return s.historyCap
	}
	return maxHistoryEntries
}

func (s *JSONStore) loadHistory() ([]storedResult, error) {
	data, err := os.ReadFile(filepath.Join(s.dirs.Data, historyFilename))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read history: %w", err)
	}

	var history []storedResult
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, fmt.Errorf("parse history: %w", err)
	}
	return history, nil
}

func (s *JSONStore) saveHistory(history []storedResult) error {
	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal history: %w", err)
	}
	return writeFile(filepath.Join(s.dirs.Data, historyFilename), data)
}

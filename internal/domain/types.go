package domain

import (
	"fmt"
	"time"
)

type CharCounts struct {
	Correct   int
	Incorrect int
	Extra     int
}

func (c CharCounts) TotalTyped() int {
	return c.Correct + c.Incorrect + c.Extra
}

type Duration int

const (
	Duration15  Duration = 15
	Duration30  Duration = 30
	Duration60  Duration = 60
	Duration120 Duration = 120
)

func ParseDuration(seconds int) (Duration, error) {
	if seconds <= 0 {
		return 0, fmt.Errorf("duration must be positive")
	}
	d := Duration(seconds)
	switch d {
	case Duration15, Duration30, Duration60, Duration120:
		return d, nil
	default:
		return d, nil
	}
}

func (d Duration) Seconds() int {
	return int(d)
}

type LiveStats struct {
	WPM       float64
	RawWPM    float64
	Accuracy  float64
	Correct   int
	Incorrect int
	Elapsed   time.Duration
}

type Result struct {
	ID                  string        `json:"id"`
	Timestamp           time.Time     `json:"timestamp"`
	Config              TestConfig    `json:"config"`
	WPM                 float64       `json:"wpm"`
	RawWPM              float64       `json:"raw_wpm"`
	Accuracy            float64       `json:"accuracy"`
	Consistency         float64       `json:"consistency,omitempty"`
	Correct             int           `json:"correct"`
	Incorrect           int           `json:"incorrect"`
	KeystrokesCorrect   int           `json:"keystrokes_correct,omitempty"`
	KeystrokesIncorrect int           `json:"keystrokes_incorrect,omitempty"`
	Skipped             int           `json:"skipped,omitempty"`
	TotalChars          int           `json:"total_chars"`
	Duration            time.Duration `json:"duration"`
	Seed                int64         `json:"seed,omitempty"`

	WPMHistory    []float64      `json:"wpm_history,omitempty"`
	RawWPMHistory []float64      `json:"raw_wpm_history,omitempty"`
	ErrorHistory  []int          `json:"error_history,omitempty"`
	CharErrors    map[string]int `json:"char_errors,omitempty"`
	Failed        bool           `json:"failed,omitempty"`
	FailureReason string         `json:"failure_reason,omitempty"`
	// History keeps only the missed words.
	Words []WordResult `json:"words,omitempty"`
}

type WordResult struct {
	Expected string `json:"expected"`
	// Set only for missed words.
	Typed string `json:"typed,omitempty"`
	// A fixed mistake still counts as missed.
	Missed    bool `json:"missed,omitempty"`
	Corrected bool `json:"corrected,omitempty"`
}

type PersonalBests struct {
	BestWPM      float64   `json:"best_wpm"`
	BestAccuracy float64   `json:"best_accuracy"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type StatsFilter struct {
	TextMode      *TextMode
	ExcludeFailed bool
}

type StatsSummary struct {
	TotalTests       int
	FilterMode       string
	AverageWPM       float64
	AverageAccuracy  float64
	BestWPM          float64
	RecentAverageWPM float64
	PersonalBest     PersonalBests
}

type Settings struct {
	DefaultDuration  Duration   `json:"default_duration"`
	DefaultWordCount int        `json:"default_word_count,omitempty"`
	DefaultWidth     int        `json:"default_width,omitempty"`
	Theme            string     `json:"theme"`
	Language         string     `json:"language,omitempty"`
	DefaultMode      TextMode   `json:"default_mode,omitempty"`
	Punctuation      bool       `json:"default_punctuation,omitempty"`
	Numbers          bool       `json:"default_numbers,omitempty"`
	Blind            bool       `json:"default_blind,omitempty"`
	Zen              bool       `json:"default_zen,omitempty"`
	DefaultMinWPM    int        `json:"default_min_wpm,omitempty"`
	Onboarded        bool       `json:"onboarded,omitempty"`
	Update           UpdateMode `json:"update,omitempty"`
}

func DefaultSettings() Settings {
	return Settings{
		DefaultDuration: Duration60,
		Theme:           "default",
	}
}

type SessionState int

const (
	SessionReady SessionState = iota
	SessionActive
	SessionFinished
)

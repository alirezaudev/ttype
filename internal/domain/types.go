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
	Correct             int           `json:"correct"`
	Incorrect           int           `json:"incorrect"`
	KeystrokesCorrect   int           `json:"keystrokes_correct,omitempty"`
	KeystrokesIncorrect int           `json:"keystrokes_incorrect,omitempty"`
	TotalChars          int           `json:"total_chars"`
	Duration            time.Duration `json:"duration"`
	Seed                int64         `json:"seed,omitempty"`
}

type Settings struct {
	DefaultDuration  Duration `json:"default_duration"`
	DefaultWordCount int      `json:"default_word_count,omitempty"`
	DefaultWidth     int      `json:"default_width,omitempty"`
	Theme            string   `json:"theme"`
	Language         string   `json:"language,omitempty"`
	DefaultMode      TextMode `json:"default_mode,omitempty"`
	Punctuation      bool     `json:"default_punctuation,omitempty"`
	Numbers          bool     `json:"default_numbers,omitempty"`
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

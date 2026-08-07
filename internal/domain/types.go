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

type Settings struct {
	DefaultDuration  Duration `json:"default_duration"`
	DefaultWordCount int      `json:"default_word_count,omitempty"`
	DefaultWidth     int      `json:"default_width,omitempty"`
	Theme            string   `json:"theme"`
	Language         string   `json:"language,omitempty"`
}

func DefaultSettings() Settings {
	return Settings{
		DefaultDuration: Duration60,
		Theme:           "default",
	}
}

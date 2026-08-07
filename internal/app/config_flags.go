package app

import (
	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/tui"
)

type TestFlags struct {
	TimeSec   int
	WordCount int
	Language  string
	Theme     string
	Width     int
}

func ConfigFromFlags(f TestFlags) (domain.TestConfig, error) {
	theme := f.Theme
	if theme == "" {
		theme = tui.ThemeDefault
	}

	themeName, err := tui.ParseThemeName(theme)
	if err != nil {
		return domain.TestConfig{}, err
	}

	cfg := domain.TestConfig{
		Language: f.Language,
		Theme:    themeName,
		Width:    f.Width,
	}

	if f.WordCount > 0 {
		cfg.Kind = domain.TestKindWords
		cfg.WordCount = f.WordCount
		return cfg, nil
	}

	duration, err := domain.ParseDuration(f.TimeSec)
	if err != nil {
		return domain.TestConfig{}, err
	}

	cfg.Kind = domain.TestKindTimed
	cfg.Duration = duration
	return cfg, nil
}

package app

import (
	"fmt"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/tui"
)

type TestFlags struct {
	TimeSec     int
	WordCount   int
	Language    string
	Theme       string
	Width       int
	Mode        string
	Punctuation bool
	Numbers     bool
	Blind       bool
	Zen         bool
	MinWPM      int
	Seed        int64
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

	mode := f.Mode
	if mode == "" {
		mode = string(domain.TextModeWords)
	}

	textMode, err := domain.ParseTextMode(mode)
	if err != nil {
		return domain.TestConfig{}, err
	}
	if textMode == domain.TextModeCustom {
		return domain.TestConfig{}, fmt.Errorf("custom text is piped in or given with --file or --text, not --mode")
	}

	cfg := domain.TestConfig{
		Language:    f.Language,
		Theme:       themeName,
		Width:       f.Width,
		TextMode:    textMode,
		Punctuation: f.Punctuation,
		Numbers:     f.Numbers,
		Blind:       f.Blind,
		Zen:         f.Zen,
		MinWPM:      f.MinWPM,
		Seed:        f.Seed,
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

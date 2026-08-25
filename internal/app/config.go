package app

import (
	"fmt"
	"io"
	"os"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/storage"
	"github.com/alirezaudev/ttype/internal/tui"
)

// ConfigChanges carries only the fields the user actually passed, so an unset
// flag never overwrites a saved default with a zero value.
type ConfigChanges struct {
	Duration    *int
	WordCount   *int
	Mode        *string
	Language    *string
	Theme       *string
	Width       *int
	MinWPM      *int
	Punctuation *bool
	Numbers     *bool
	Blind       *bool
	Zen         *bool
}

func (c ConfigChanges) empty() bool {
	return c == ConfigChanges{}
}

func RunConfig(store storage.Store, changes ConfigChanges) error {
	settings, err := store.LoadSettings()
	if err != nil {
		return fmt.Errorf("load settings: %w", err)
	}

	if changes.empty() {
		printSettings(os.Stdout, settings, store.Paths())
		return nil
	}

	if settings, err = applyConfigChanges(settings, changes); err != nil {
		return err
	}
	if err := store.SaveSettings(settings); err != nil {
		return fmt.Errorf("save settings: %w", err)
	}

	printSettings(os.Stdout, settings, store.Paths())
	return nil
}

func applyConfigChanges(settings domain.Settings, c ConfigChanges) (domain.Settings, error) {
	if c.Duration != nil {
		duration, err := domain.ParseDuration(*c.Duration)
		if err != nil {
			return settings, err
		}
		settings.DefaultDuration = duration
		settings.DefaultWordCount = 0
	}
	if c.WordCount != nil {
		if *c.WordCount < 1 {
			return settings, fmt.Errorf("word count must be at least 1")
		}
		settings.DefaultWordCount = *c.WordCount
	}
	if c.Mode != nil {
		mode, err := domain.ParseTextMode(*c.Mode)
		if err != nil {
			return settings, err
		}
		settings.DefaultMode = mode
	}
	if c.Theme != nil {
		theme, err := tui.ParseThemeName(*c.Theme)
		if err != nil {
			return settings, err
		}
		settings.Theme = theme
	}
	if c.Language != nil {
		settings.Language = *c.Language
	}
	if c.Width != nil {
		settings.DefaultWidth = *c.Width
	}
	if c.MinWPM != nil {
		settings.DefaultMinWPM = *c.MinWPM
	}
	if c.Punctuation != nil {
		settings.Punctuation = *c.Punctuation
	}
	if c.Numbers != nil {
		settings.Numbers = *c.Numbers
	}
	if c.Blind != nil {
		settings.Blind = *c.Blind
	}
	if c.Zen != nil {
		settings.Zen = *c.Zen
	}
	return settings, nil
}

func printSettings(out io.Writer, settings domain.Settings, dirs storage.Dirs) {
	test := fmt.Sprintf("%ds", settings.DefaultDuration.Seconds())
	if settings.DefaultWordCount > 0 {
		test = fmt.Sprintf("%d words", settings.DefaultWordCount)
	}

	mode := settings.DefaultMode
	if mode == "" {
		mode = domain.TextModeWords
	}
	language := settings.Language
	if language == "" {
		language = "english (built-in)"
	}
	width := "auto"
	if settings.DefaultWidth > 0 {
		width = fmt.Sprintf("%d", settings.DefaultWidth)
	}
	minWPM := "off"
	if settings.DefaultMinWPM > 0 {
		minWPM = fmt.Sprintf("%d", settings.DefaultMinWPM)
	}

	rows := [][2]string{
		{"test", test},
		{"mode", string(mode)},
		{"language", language},
		{"theme", settings.Theme},
		{"width", width},
		{"min wpm", minWPM},
		{"punctuation", onOff(settings.Punctuation)},
		{"numbers", onOff(settings.Numbers)},
		{"blind", onOff(settings.Blind)},
		{"zen", onOff(settings.Zen)},
	}
	for _, row := range rows {
		fmt.Fprintf(out, "  %-12s %s\n", row[0], row[1])
	}

	fmt.Fprintf(out, "\n  %-12s %s\n", "config", dirs.Config)
	fmt.Fprintf(out, "  %-12s %s\n", "data", dirs.Data)
}

func onOff(v bool) string {
	if v {
		return "on"
	}
	return "off"
}

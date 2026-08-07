package main

import (
	"flag"
	"log"

	"github.com/alirezaudev/ttype/internal/app"
	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/engine"
	"github.com/alirezaudev/ttype/internal/storage"
	"github.com/alirezaudev/ttype/internal/text"
	"github.com/alirezaudev/ttype/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	wordCount := flag.Int("words", 0, "run a words test of this many words instead of a timed test")
	language := flag.String("language", "", "use a downloadable language word list instead of the built-in one")
	flag.Parse()

	dirs, err := storage.DefaultDirs()
	if err != nil {
		log.Fatalln(err)
	}

	store, err := storage.NewJSONStore(dirs.Config)
	if err != nil {
		log.Fatalln(err)
	}

	provider, err := text.NewProvider(dirs.Data)
	if err != nil {
		log.Fatalln(err)
	}

	cfg, err := resolveTestConfig(store, *wordCount, *language)
	if err != nil {
		log.Fatalln(err)
	}

	session, err := engine.NewSession(cfg, provider, nil)
	if err != nil {
		log.Fatalln(err)
	}

	p := tea.NewProgram(tui.NewAppModel(cfg, provider, provider.LanguageCache(), session, store), tea.WithAltScreen())

	_, err = p.Run()
	if err != nil {
		log.Fatalln(err)
	}
}

func resolveTestConfig(store storage.Store, wordCount int, language string) (domain.TestConfig, error) {
	settings, err := store.LoadSettings()
	if err != nil {
		return domain.TestConfig{}, err
	}

	f := app.TestFlags{
		TimeSec:   domain.Duration60.Seconds(),
		WordCount: wordCount,
		Language:  language,
		Theme:     settings.Theme,
		Width:     settings.DefaultWidth,
	}

	if settings.DefaultDuration > 0 {
		f.TimeSec = settings.DefaultDuration.Seconds()
	}
	if wordCount == 0 {
		f.WordCount = settings.DefaultWordCount
	}
	if language == "" {
		f.Language = settings.Language
	}

	return app.ConfigFromFlags(f)
}

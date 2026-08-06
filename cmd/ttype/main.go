package main

import (
	"flag"
	"log"

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

	cfg := domain.TestConfig{Kind: domain.TestKindTimed, Duration: domain.Duration60}
	count := 100
	if settings, err := store.LoadSettings(); err == nil {
		cfg.Theme = settings.Theme
		cfg.Width = settings.DefaultWidth
		cfg.Language = settings.Language
		if settings.DefaultWordCount > 0 {
			cfg.Kind = domain.TestKindWords
			count = settings.DefaultWordCount
		} else if settings.DefaultDuration > 0 {
			cfg.Duration = settings.DefaultDuration
		}
	}
	if *wordCount > 0 {
		cfg.Kind = domain.TestKindWords
		count = *wordCount
	}
	if *language != "" {
		cfg.Language = *language
	}
	cfg.WordCount = count

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

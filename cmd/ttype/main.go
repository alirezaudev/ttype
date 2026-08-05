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
	flag.Parse()

	store, err := storage.NewDefaultStore()
	if err != nil {
		log.Fatalln(err)
	}

	provider, err := text.NewProvider()
	if err != nil {
		log.Fatalln(err)
	}

	cfg := domain.TestConfig{Kind: domain.TestKindTimed, Duration: domain.Duration60}
	count := 100
	if settings, err := store.LoadSettings(); err == nil {
		cfg.Theme = settings.Theme
		cfg.Width = settings.DefaultWidth
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
	cfg.WordCount = count

	session, err := engine.NewSession(cfg, provider, nil)
	if err != nil {
		log.Fatalln(err)
	}

	p := tea.NewProgram(tui.NewAppModel(cfg, provider, provider, session, store), tea.WithAltScreen())

	_, err = p.Run()
	if err != nil {
		log.Fatalln(err)
	}
}

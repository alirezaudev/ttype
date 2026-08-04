package main

import (
	"flag"
	"log"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/alirezaudev/ttype/assets"
	"github.com/alirezaudev/ttype/internal/engine"
	"github.com/alirezaudev/ttype/internal/storage"
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

	words, err := assets.LoadWords()
	if err != nil {
		log.Fatalln(err)
	}

	cfg := engine.Config{Kind: engine.TestKindTimed, Duration: 60 * time.Second}
	count := 100
	if settings, err := store.LoadSettings(); err == nil {
		cfg.Theme = settings.Theme
		cfg.Width = settings.DefaultWidth
		if settings.DefaultWordCount > 0 {
			cfg.Kind = engine.TestKindWords
			count = settings.DefaultWordCount
		} else if settings.DefaultDuration > 0 {
			cfg.Duration = time.Duration(settings.DefaultDuration) * time.Second
		}
	}
	if *wordCount > 0 {
		cfg.Kind = engine.TestKindWords
		count = *wordCount
	}
	cfg.WordCount = count

	newTarget := func() (string, error) {
		sampled := make([]string, count)
		for i := range sampled {
			sampled[i] = words[rand.IntN(len(words))]
		}
		return strings.Join(sampled, " "), nil
	}

	session, err := engine.NewSession(newTarget, cfg, nil)
	if err != nil {
		log.Fatalln(err)
	}

	p := tea.NewProgram(tui.NewAppModel(cfg, newTarget, session, store), tea.WithAltScreen())

	_, err = p.Run()
	if err != nil {
		log.Fatalln(err)
	}
}

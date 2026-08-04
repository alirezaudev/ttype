package main

import (
	"flag"
	"log"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/alirezaudev/ttype/assets"
	"github.com/alirezaudev/ttype/internal/engine"
	"github.com/alirezaudev/ttype/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	wordCount := flag.Int("words", 0, "run a words test of this many words instead of a timed test")
	flag.Parse()

	words, err := assets.LoadWords()
	if err != nil {
		log.Fatalln(err)
	}

	cfg := engine.Config{Kind: engine.TestKindTimed, Duration: 10 * time.Second}
	count := 100
	if *wordCount > 0 {
		cfg = engine.Config{Kind: engine.TestKindWords}
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

	p := tea.NewProgram(tui.NewTestModel(session), tea.WithAltScreen())

	_, err = p.Run()
	if err != nil {
		log.Fatalln(err)
	}
}

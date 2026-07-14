package main

import (
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
	words, err := assets.LoadWords()
	if err != nil {
		log.Fatalln(err)
	}

	newTarget := func() (string, error) {
		sampled := make([]string, 100)
		for i := range sampled {
			sampled[i] = words[rand.IntN(len(words))]
		}
		return strings.Join(sampled, " "), nil
	}

	session, err := engine.NewSession(newTarget, 10*time.Second, nil)
	if err != nil {
		log.Fatalln(err)
	}

	p := tea.NewProgram(tui.NewTestModel(session), tea.WithAltScreen())

	_, err = p.Run()
	if err != nil {
		log.Fatalln(err)
	}
}

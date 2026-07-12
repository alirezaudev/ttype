package main

import (
	"log"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/alirezaudev/ttype/assets"
	"github.com/alirezaudev/ttype/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	words, err := assets.LoadWords()
	if err != nil {
		log.Fatalln(err)
	}

	sampled := make([]string, 100)
	for i := range sampled {
		sampled[i] = words[rand.IntN(len(words))]
	}

	p := tea.NewProgram(tui.NewTestModel(
		strings.Join(sampled, " "),
		10*time.Second,
	), tea.WithAltScreen())

	_, err = p.Run()
	if err != nil {
		log.Fatalln(err)
	}
}

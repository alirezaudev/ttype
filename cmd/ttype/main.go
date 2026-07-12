package main

import (
	"log"
	"time"

	"github.com/alirezaudev/ttype/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {

	p := tea.NewProgram(tui.NewTestModel(
		"Yes, man has a hard life. I think the best definition of man is this: a creature who can grow accustomed to anything.",
		10*time.Second,
	), tea.WithAltScreen())

	_, err := p.Run()
	if err != nil {
		log.Fatalln(err)
	}
}

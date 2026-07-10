package main

import (
	"log"
	"os"
	"time"

	"github.com/alirezaudev/ttype/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {

	p := tea.NewProgram(tui.TestModel{
		Text:     "Yes, man has a hard life. I think the best definition of man is this: a creature who can grow accustomed to anything.",
		Duration: 10 * time.Second,
	})

	_, err := p.Run()
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}
}

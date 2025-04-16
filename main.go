package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
)

type mainModel struct{}

func main() {
	p := tea.NewProgram(&mainModel{})
	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"log"

	"agent/models"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	m := &models.MainModel{}
	p := tea.NewProgram(m)
	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}

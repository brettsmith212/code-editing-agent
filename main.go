package main

import (
	"log"

	"agent/agent"
	"agent/models"
	"agent/tools"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/anthropics/anthropic-sdk-go"
)

func main() {
	client := anthropic.NewClient()
	toolDefs := []tools.ToolDefinition{ /* Add tool definitions here if needed */ }
	myAgent := agent.NewAgent(&client, nil, toolDefs)

	m := &models.MainModel{
		Agent: myAgent,
	}
	p := tea.NewProgram(m)
	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}

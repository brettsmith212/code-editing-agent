package main

import (
	"log"
	"path/filepath"

	"agent/agent"
	"agent/logger"
	"agent/models"
	"agent/tools"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/anthropics/anthropic-sdk-go"
)

func main() {
	// Initialize logger
	logDir := filepath.Join(".", "logs")
	if err := logger.Initialize(logDir); err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logger.Close()

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

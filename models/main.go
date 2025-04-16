package models

import (
	tea "github.com/charmbracelet/bubbletea"
)

type MainModel struct {
	chat        *chatModel
	codeview    *codeviewModel
	sidebar     *sidebarModel
	agent       interface{} // Placeholder for *agent.Agent
	conversation []interface{} // Placeholder for []anthropic.MessageParam
	state       string
}

func (m *MainModel) Init() tea.Cmd {
	return nil
}

func (m *MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m *MainModel) View() string {
	return "Main UI (placeholder)"
}

// Placeholder structs for compilation

type codeviewModel struct{}
type sidebarModel struct{}

package models

import (
	tea "github.com/charmbracelet/bubbletea"
)

type MainModel struct {
	chat        *chatModel
	codeview    *codeviewModel
	sidebar     *sidebarModel
	agent       interface{} // Placeholder for *agent.Agent
	conversation []string // Conversation history as plain strings for now
	state       string
	quitting    bool
}

func (m *MainModel) Init() tea.Cmd {
	m.chat = newChatModel()
	m.conversation = []string{}
	return nil
}

func (m *MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle Ctrl+C (tea.KeyMsg with Ctrl+C)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			m.quitting = true
			return m, tea.Quit
		}
		// On enter, append textarea content to conversation
		if m.chat != nil && msg.Type == tea.KeyEnter {
			input := m.chat.textarea.Value()
			if input != "" {
				m.conversation = append(m.conversation, input)
				m.chat.textarea.Reset()
				// Update viewport content
				m.chat.viewport.SetContent(joinConversation(m.conversation))
			}
		}
	}
	// Forward updates to chat sub-model for now
	if m.chat != nil {
		updatedModel, cmd := m.chat.Update(msg)
		m.chat = updatedModel.(*chatModel)
		return m, cmd
	}
	return m, nil
}

func (m *MainModel) View() string {
	if m.quitting {
		return "Goodbye!"
	}
	if m.chat != nil {
		return m.chat.View()
	}
	return "Chat"
}

func joinConversation(conv []string) string {
	result := ""
	for _, line := range conv {
		result += line + "\n"
	}
	return result
}

// Placeholder structs for compilation

type codeviewModel struct{}
type sidebarModel struct{}

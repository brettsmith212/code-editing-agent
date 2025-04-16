package models

import (
	tea "github.com/charmbracelet/bubbletea"
	"context"
	"agent/agent"
)

type claudeResponseMsg struct{
	Text string
	Err error
}

type MainModel struct {
	chat        *chatModel
	codeview    *codeviewModel
	sidebar     *sidebarModel
	Agent       *agent.Agent
	conversation []string // Conversation history as plain strings for now
	state       string
	quitting    bool
	waitingForClaude bool
}

func (m *MainModel) Init() tea.Cmd {
	m.chat = newChatModel()
	m.conversation = []string{}
	m.waitingForClaude = false
	
	cmds := []tea.Cmd{
		tea.EnterAltScreen,
	}
	
	// Get initial window size
	return tea.Batch(cmds...)
}

func (m *MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// First pass the size message to the chat model
		if m.chat != nil {
			updatedModel, _ := m.chat.Update(msg)
			m.chat = updatedModel.(*chatModel)
		}
		return m, nil
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			m.quitting = true
			return m, tea.Quit
		}
		if m.chat != nil && msg.Type == tea.KeyEnter && !m.waitingForClaude {
			input := m.chat.textarea.Value()
			if input != "" {
				m.conversation = append(m.conversation, "You: "+input)
				m.chat.textarea.Reset()
				m.chat.AddMessage("User", input)
				m.waitingForClaude = true
				return m, m.sendToClaude(input)
			}
		}
	case claudeResponseMsg:
		m.waitingForClaude = false
		if msg.Err != nil {
			m.conversation = append(m.conversation, "Claude (error): "+msg.Err.Error())
			m.chat.AddMessage("Claude (error)", msg.Err.Error())
		} else {
			m.conversation = append(m.conversation, "Claude: "+msg.Text)
			m.chat.AddMessage("Claude", msg.Text)
		}
	}
	if m.chat != nil {
		updatedModel, cmd := m.chat.Update(msg)
		m.chat = updatedModel.(*chatModel)
		return m, cmd
	}
	return m, nil
}

func (m *MainModel) sendToClaude(input string) tea.Cmd {
	return func() tea.Msg {
		if m.Agent == nil {
			return claudeResponseMsg{Err: context.DeadlineExceeded}
		}
		ctx := context.Background()
		resp, err := m.Agent.RunInference(ctx, input)
		if err != nil {
			return claudeResponseMsg{Err: err}
		}
		return claudeResponseMsg{Text: resp}
	}
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

type codeviewModel struct{}
type sidebarModel struct{}

package models

import (
	"agent/logger"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

var (
	chatBorderStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(1, 2).BorderForeground(lipgloss.Color("63"))
	viewportStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(1, 2).BorderForeground(lipgloss.Color("99")).Height(12)
)

type chatModel struct {
	textarea textarea.Model
	viewport viewport.Model
	messages []string
}

func newChatModel() *chatModel {
	ta := textarea.New()
	ta.Placeholder = "Type your message..."
	ta.Prompt = "> "
	ta.Focus()

	ta.SetWidth(60)
	ta.SetHeight(3)

	vp := viewport.New(60, 12)
	vp.Style = viewportStyle

	return &chatModel{
		textarea: ta,
		viewport: vp,
		messages: make([]string, 0),
	}
}

func (c *chatModel) Init() tea.Cmd {
	return nil
}

func (c *chatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyEnter && !msg.Alt {
			userMsg := c.textarea.Value()
			if userMsg != "" {
				// Log user message
				logger.LogMessage("User", userMsg)
				c.messages = append(c.messages, "User: "+userMsg)
				c.viewport.SetContent(c.formatMessages())
				c.textarea.Reset()
				return c, nil
			}
		}
	}

	c.textarea, cmd = c.textarea.Update(msg)
	c.viewport, _ = c.viewport.Update(msg)
	return c, cmd
}

func (c *chatModel) View() string {
	return chatBorderStyle.Render(
		c.viewport.View()+"\n"+c.textarea.View(),
	)
}

func (c *chatModel) formatMessages() string {
	var content string
	for _, msg := range c.messages {
		content += msg + "\n"
	}
	return content
}

// AddAIMessage adds an AI response to the chat and logs it
func (c *chatModel) AddAIMessage(content string) {
	logger.LogMessage("AI", content)
	c.messages = append(c.messages, "AI: "+content)
	c.viewport.SetContent(c.formatMessages())
}

package models

import (
	"agent/logger"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"fmt"
)

var (
	chatBorderStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2).BorderForeground(lipgloss.Color("69"))
	viewportStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2).BorderForeground(lipgloss.Color("69"))
)

type chatModel struct {
	textarea textarea.Model
	viewport viewport.Model
	messages []string
	width    int
	height   int
}

func newChatModel() *chatModel {
	ta := textarea.New()
	ta.Placeholder = "Type your message..."
	ta.Prompt = ""
	ta.ShowLineNumbers = false
	ta.Focus()

	// Initial default size, will be updated on WindowSizeMsg
	initialWidth := 100
	initialHeight := 20

	ta.SetWidth(initialWidth)
	ta.SetHeight(1)

	// Remove highlighting and make it plain
	ta.FocusedStyle.Base = lipgloss.NewStyle().
		BorderForeground(lipgloss.Color("69")).
		BorderStyle(lipgloss.RoundedBorder())
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.FocusedStyle.Text = lipgloss.NewStyle()
	ta.BlurredStyle.Base = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder())
	ta.BlurredStyle.Text = lipgloss.NewStyle()

	vp := viewport.New(initialWidth, initialHeight-3) // Leave space for the textarea
	vp.Style = viewportStyle

	return &chatModel{
		textarea: ta,
		viewport: vp,
		messages: make([]string, 0),
		width:    initialWidth,
		height:   initialHeight,
	}
}

// updateSize updates the component sizes based on the terminal dimensions
func (c *chatModel) updateSize(width, height int) {
	c.width = width
	c.height = height

	// Adjust width for padding and borders
	contentWidth := width - 6 // Account for padding and borders
	if contentWidth < 20 {    // Minimum reasonable width
		contentWidth = 20
	}

	// Adjust viewport height to leave space for input
	viewportHeight := height - 4 // Space for textarea and some padding
	if viewportHeight < 5 {      // Minimum reasonable height
		viewportHeight = 5
	}

	c.textarea.SetWidth(contentWidth)
	c.viewport.Width = contentWidth
	c.viewport.Height = viewportHeight

	// Update the content to fit the new size
	c.viewport.SetContent(c.formatMessages())
}

func (c *chatModel) Init() tea.Cmd {
	// We'll let the main program handle initial window size
	return nil
}

func (c *chatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Handle window resize
		c.updateSize(msg.Width, msg.Height)
		return c, nil
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
	// Create a layout that takes the full available space
	return fmt.Sprintf("%s\n%s", 
		c.viewport.View(),
		c.textarea.View(),
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

package models

import (
	"agent/logger"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	userPrefix         = "User: "
	aiPrefix           = "AI: "
	claudePrefix       = "Claude: "
	claudeErrorPrefix  = "Claude (error): "
	paddingWidth       = 6
	minContentWidth    = 20
	minViewportHeight  = 5
	textareaHeight     = 1
	defaultPlaceholder = "Type your message..."
)

var (
	// Use normal borders instead of rounded for consistency
	viewportStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(1, 2).BorderForeground(lipgloss.Color("69"))
)

// chatModel holds state for the chat panel.
type chatModel struct {
	textarea textarea.Model
	viewport viewport.Model
	messages []string
	width    int
	height   int
}

// newChatModel creates and initializes a new chatModel.
func newChatModel() *chatModel {
	ta := textarea.New()
	ta.Placeholder = defaultPlaceholder
	ta.Prompt = ""
	ta.ShowLineNumbers = false
	ta.Focus()

	// Initial default size, will be updated on WindowSizeMsg
	initialWidth := 100
	initialHeight := 20

	ta.SetWidth(initialWidth)
	ta.SetHeight(textareaHeight)

	// Remove highlighting and make it plain
	ta.FocusedStyle.Base = lipgloss.NewStyle().
		BorderForeground(lipgloss.Color("69")).
		BorderStyle(lipgloss.NormalBorder())
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.FocusedStyle.Text = lipgloss.NewStyle()
	ta.BlurredStyle.Base = lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder())
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

// updateSize updates the component sizes based on the terminal dimensions.
func (m *chatModel) updateSize(width, height int) {
	m.width = width
	m.height = height

	// Adjust width for padding and borders
	contentWidth := width - paddingWidth // Account for padding and borders
	if contentWidth < minContentWidth {  // Minimum reasonable width
		contentWidth = minContentWidth
	}

	// Set viewport height to fill the available space
	// Adjust to leave just enough room for textarea (1 line + border)
	viewportHeight := height - 3            // Space for textarea and some padding
	if viewportHeight < minViewportHeight { // Minimum reasonable height
		viewportHeight = minViewportHeight
	}

	m.textarea.SetWidth(contentWidth)
	m.viewport.Width = contentWidth
	m.viewport.Height = viewportHeight

	// Set viewport options for better rendering
	m.viewport.SetContent(m.formatMessages())
	m.viewport.YPosition = 0
}

// Init is required by Bubbletea but not used for chatModel.
func (m *chatModel) Init() tea.Cmd {
	// We'll let the main program handle initial window size
	return nil
}

// Update handles Bubbletea messages for the chat panel.
func (m *chatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Handle window resize
		m.updateSize(msg.Width, msg.Height)
		return m, nil
	case tea.KeyMsg:
		if msg.Type == tea.KeyEnter && !msg.Alt {
			userMsg := m.textarea.Value()
			if userMsg != "" {
				// Log user message
				logger.LogMessage("User", userMsg)
				m.messages = append(m.messages, userPrefix+userMsg)
				m.viewport.SetContent(m.formatMessages())
				m.textarea.Reset()
				return m, nil
			}
		}
	}

	m.textarea, cmd = m.textarea.Update(msg)
	m.viewport, _ = m.viewport.Update(msg)
	return m, cmd
}

// View renders the chat panel (viewport + textarea).
func (m *chatModel) View() string {
	// Create a layout that takes the full available space
	// Add a newline at the beginning to ensure top border is visible
	return fmt.Sprintf("\n%s\n%s",
		m.viewport.View(),
		m.textarea.View(),
	)
}

// formatMessages prepares the chat messages for display in the viewport.
func (m *chatModel) formatMessages() string {
	var formattedContent strings.Builder

	// Calculate the content width (accounting for padding and borders)
	// Subtract more to ensure text doesn't get cut off
	contentWidth := m.viewport.Width - paddingWidth
	if contentWidth < minContentWidth {
		contentWidth = minContentWidth // Minimum reasonable width
	}

	for _, msg := range m.messages {
		// Split the message into parts (e.g., "User: Hello")
		parts := strings.SplitN(msg, ": ", 2)
		if len(parts) != 2 {
			// If not in expected format, just add the message
			formattedContent.WriteString(wrapText(msg, contentWidth))
			formattedContent.WriteString("\n\n")
			continue
		}

		// Get the sender and content
		sender, content := parts[0], parts[1]

		// Add the sender with formatting
		if sender == "User" {
			formattedContent.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("69")).Render(sender) + ": ")
		} else {
			formattedContent.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("176")).Render(sender) + ": ")
		}

		// Account for the prefix in our width calculation
		// We have to subtract the length of the prefix from our available width
		prefixLen := len(sender) + 2 // +2 for ": "
		adjustedWidth := contentWidth
		if prefixLen < contentWidth {
			adjustedWidth = contentWidth - prefixLen
		}

		// Wrap the content text and add it
		wrappedContent := wrapText(content, adjustedWidth)

		// For the first line, we'll use it as is, but for subsequent lines
		// we need to add proper indentation to align with the first line's content
		lines := strings.Split(wrappedContent, "\n")
		if len(lines) > 0 {
			formattedContent.WriteString(lines[0])

			// Add proper indentation for subsequent lines
			if len(lines) > 1 {
				indent := strings.Repeat(" ", prefixLen)
				for _, line := range lines[1:] {
					formattedContent.WriteString("\n" + indent + line)
				}
			}
			formattedContent.WriteString("\n\n")
		}
	}

	return formattedContent.String()
}

// boolToInt converts a boolean to an integer (1 for true, 0 for false).
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// AddAIMessage adds an AI response to the chat and logs it.
func (m *chatModel) AddAIMessage(content string) {
	logger.LogMessage("AI", content)
	m.messages = append(m.messages, aiPrefix+content)
	m.viewport.SetContent(m.formatMessages())
}

// AddMessage adds a message to the chat with the given sender and content.
func (m *chatModel) AddMessage(sender, content string) {
	prefix := userPrefix
	switch sender {
	case "AI":
		prefix = aiPrefix
	case "Claude":
		prefix = claudePrefix
	case "Claude (error)":
		prefix = claudeErrorPrefix
	}
	m.messages = append(m.messages, prefix+content)
	m.viewport.SetContent(m.formatMessages())
}

package models

import (
	"agent/logger"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"fmt"
	"strings"
)

var (
	// Use normal borders instead of rounded for consistency
	chatBorderStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(1, 2).BorderForeground(lipgloss.Color("69"))
	viewportStyle   = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(1, 2).BorderForeground(lipgloss.Color("69"))
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

// updateSize updates the component sizes based on the terminal dimensions
func (c *chatModel) updateSize(width, height int) {
	c.width = width
	c.height = height
	
	// Adjust width for padding and borders
	contentWidth := width - 6 // Account for padding and borders
	if contentWidth < 20 {    // Minimum reasonable width
		contentWidth = 20
	}
	
	// Set viewport height to fill the available space
	// Adjust to leave just enough room for textarea (1 line + border)
	viewportHeight := height - 3 // Space for textarea and some padding
	if viewportHeight < 5 {      // Minimum reasonable height
		viewportHeight = 5
	}
	
	c.textarea.SetWidth(contentWidth)
	c.viewport.Width = contentWidth
	c.viewport.Height = viewportHeight
	
	// Set viewport options for better rendering
	c.viewport.SetContent(c.formatMessages())
	c.viewport.YPosition = 0
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
	// Add a newline at the beginning to ensure top border is visible
	return fmt.Sprintf("\n%s\n%s", 
		c.viewport.View(),
		c.textarea.View(),
	)
}

func (c *chatModel) formatMessages() string {
	var formattedContent strings.Builder
	
	// Calculate the content width (accounting for padding and borders)
	// Subtract more to ensure text doesn't get cut off
	contentWidth := c.viewport.Width - 6
	if contentWidth < 10 {
		contentWidth = 10 // Minimum reasonable width
	}
	
	for _, msg := range c.messages {
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
		}
		
		formattedContent.WriteString("\n\n") // Add space between messages
	}
	
	return formattedContent.String()
}

// wrapText wraps text at the given width, breaking long words if necessary
func wrapText(text string, width int) string {
	// Handle empty text
	if text == "" {
		return ""
	}
	
	var result strings.Builder
	
	// Split the text into words
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}
	
	lineLength := 0
	isFirstWord := true
	
	// Process each word
	for _, word := range words {
		wordLen := len(word)
		
		// If this word is too long for a line by itself, we need to break it
		if wordLen > width {
			// If not the first word on the line, start a new line
			if !isFirstWord {
				result.WriteString("\n")
				lineLength = 0
			}
			
			// Break the long word into chunks
			for i := 0; i < wordLen; i += width {
				end := i + width
				if end > wordLen {
					end = wordLen
				}
				
				// Add the chunk
				result.WriteString(word[i:end])
				
				// Add a newline if there's more of this word to come
				if end < wordLen {
					result.WriteString("-\n")
				}
			}
			lineLength = wordLen % width
			if lineLength == 0 && wordLen > 0 {
				lineLength = width
			}
		} else {
			// Normal word that fits on a line
			if lineLength+wordLen+(1-boolToInt(isFirstWord)) > width {
				// Word won't fit on current line, start a new one
				result.WriteString("\n")
				result.WriteString(word)
				lineLength = wordLen
			} else {
				// Word fits on current line
				if !isFirstWord {
					result.WriteString(" ")
					lineLength++
				}
				result.WriteString(word)
				lineLength += wordLen
			}
		}
		
		isFirstWord = false
	}
	
	return result.String()
}

// boolToInt converts a boolean to an integer (1 for true, 0 for false)
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// AddAIMessage adds an AI response to the chat and logs it
func (c *chatModel) AddAIMessage(content string) {
	logger.LogMessage("AI", content)
	c.messages = append(c.messages, "AI: "+content)
	c.viewport.SetContent(c.formatMessages())
}

// AddMessage adds a message to the chat with the given sender and content
func (c *chatModel) AddMessage(sender, content string) {
	// Log the message
	logger.LogMessage(sender, content)
	
	// Add to messages
	c.messages = append(c.messages, sender+": "+content)
	
	// Update viewport content with formatted messages
	c.viewport.SetContent(c.formatMessages())
	
	// Make sure we scroll to the bottom
	c.viewport.GotoBottom()
}

package models

import (
	"agent/agent"
	"context"
	"io/ioutil"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// claudeResponseMsg is used to deliver responses from Claude (or errors) asynchronously.
type claudeResponseMsg struct {
	Text string
	Err  error
}

// openFileMsg is used to deliver file open requests.
type openFileMsg struct {
	FileName string
}

// MainModel is the root model for the Bubbletea application.
type MainModel struct {
	chat         *chatModel
	codeview     *codeviewModel
	sidebar      *sidebarModel
	Agent        *agent.Agent
	conversation []string // Conversation history as plain strings for now
	state        string
	quitting     bool
	waitingForClaude bool
	width        int    // Terminal width
	height       int    // Terminal height
	focusedPane  string // "sidebar" or "chat"
	sidebarShowingFile bool
}

// Init sets up the initial state for the main model.
func (m *MainModel) Init() tea.Cmd {
	m.chat = newChatModel()
	m.conversation = []string{}
	m.waitingForClaude = false
	m.sidebar = newSidebarModelFromDir(".")
	// Initialize codeview with default width and height (will be updated on WindowSizeMsg)
	m.codeview = NewCodeViewModel(80, 20)
	m.focusedPane = "chat"

	cmds := []tea.Cmd{
		tea.EnterAltScreen,
	}

	// Get initial window size
	return tea.Batch(cmds...)
}

// Update handles all Bubbletea messages for the main model.
func (m *MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height - 2 // Reserve space for top margin
		// Calculate panel widths
		leftPanelWidth := m.width * 3 / 10
		if leftPanelWidth < 20 {
			leftPanelWidth = 20
		}
		rightPanelWidth := m.width - leftPanelWidth
		if rightPanelWidth < 40 {
			rightPanelWidth = 40
		}

		// Use consistent height for both panels
		panelHeight := m.height - 2 // Reserve space for margins

		if m.chat != nil {
			m.chat.updateSize(rightPanelWidth, panelHeight)
		}
		if m.sidebar != nil {
			m.sidebar.updateSize(leftPanelWidth, panelHeight)
		}
		if m.codeview != nil {
			m.codeview.viewport.Width = leftPanelWidth - 2
			m.codeview.viewport.Height = panelHeight - 2
		}
		return m, nil
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			m.quitting = true
			return m, tea.Quit
		}
		if m.sidebarShowingFile && msg.Type == tea.KeyEsc {
			// Escape closes file view and returns to sidebar
			m.sidebarShowingFile = false
			return m, nil
		}
		if msg.Type == tea.KeyTab {
			if m.focusedPane == "chat" {
				m.focusedPane = "sidebar"
			} else {
				m.focusedPane = "chat"
			}
			return m, nil
		}
		if msg.Type == tea.KeyEnter && !m.waitingForClaude {
			input := m.chat.textarea.Value()
			if input != "" {
				m.conversation = append(m.conversation, "You: "+input)
				m.chat.textarea.Reset()
				m.chat.AddMessage("User", input)
				m.waitingForClaude = true
				return m, m.sendToClaude(input)
			}
		}
	case openFileMsg:
		// Read file and open in codeview (in sidebar panel)
		content, err := ioutil.ReadFile(msg.FileName)
		if err == nil && m.codeview != nil {
			m.codeview.OpenTab(msg.FileName, string(content))
			m.sidebarShowingFile = true
		}
		return m, nil
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
	// Forward input to focused pane
	if m.focusedPane == "sidebar" && m.sidebar != nil && !m.sidebarShowingFile {
		cmd := m.sidebar.Update(msg)
		return m, cmd
	}
	if m.focusedPane == "chat" && m.chat != nil {
		updatedModel, cmd := m.chat.Update(msg)
		m.chat = updatedModel.(*chatModel)
		return m, cmd
	}
	return m, nil
}

// sendToClaude sends a user message to Claude and returns a command that will deliver the response asynchronously.
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

// View renders the main UI, including both panels and their borders.
func (m *MainModel) View() string {
	if m.quitting {
		return "Goodbye!"
	}
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	// Calculate panel widths - using same calculation as Update method
	leftPanelWidth := m.width * 3 / 10
	if leftPanelWidth < 20 {
		leftPanelWidth = 20
	}
	rightPanelWidth := m.width - leftPanelWidth - 1 // -1 for the gap between panels
	if rightPanelWidth < 40 {
		rightPanelWidth = 40
	}

	// Calculate usable height - make sure both panels use the same exact height
	usableHeight := m.height - 4

	// Ensure both panels have exactly the same height
	exactHeight := usableHeight - 2 // Account for borders

	// Create left panel: either sidebar or codeview (not both)
	var leftPanel string
	if m.sidebarShowingFile && m.codeview != nil && len(m.codeview.tabs) > 0 {
		leftPanel = m.codeview.View()
	} else if m.sidebar != nil {
		leftPanel = m.sidebar.View()
	}
	if leftPanel == "" {
		leftPanel = " " // Empty panel fallback
	}

	// Right panel: chat
	var rightPanel string
	if m.chat != nil {
		rightPanel = m.chat.View()
	} else {
		rightPanel = "Chat"
	}

	// Always draw borders on both panels, but only highlight the focused one
	// This ensures content doesn't move when toggling focus

	// Base styles with consistent borders and padding for both panels
	leftStyle := lipgloss.NewStyle().
		Width(leftPanelWidth-2).              // Account for border width
		Height(exactHeight).                  // Set a fixed height for consistency
		BorderStyle(lipgloss.NormalBorder()). // Always have borders
		Padding(0, 1)                         // Consistent padding

	rightStyle := lipgloss.NewStyle().
		Width(rightPanelWidth-2).             // Account for border width
		Height(exactHeight).                  // Same fixed height as left panel
		BorderStyle(lipgloss.NormalBorder()). // Always have borders
		Padding(0, 1)                         // Consistent padding

	// Set border color based on focus - use a neutral color for unfocused panels
	// and highlight color for focused panel
	unfocusedBorderColor := lipgloss.Color("240") // Subtle gray
	focusedBorderColor := lipgloss.Color("69")    // Highlight color

	// Apply appropriate border colors based on focus
	if m.focusedPane == "sidebar" {
		leftStyle = leftStyle.BorderForeground(focusedBorderColor)
		rightStyle = rightStyle.BorderForeground(unfocusedBorderColor)
	} else {
		leftStyle = leftStyle.BorderForeground(unfocusedBorderColor)
		rightStyle = rightStyle.BorderForeground(focusedBorderColor)
	}

	// Force a space between panels
	spacer := lipgloss.NewStyle().Width(1).Render(" ")

	// Combine panels horizontally
	row := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftStyle.Render(leftPanel),
		spacer,
		rightStyle.Render(rightPanel),
	)

	// Add a newline at the top to ensure top border is visible
	return "\n\n" + row // Add extra newline for top margin
}

// joinConversation joins conversation history into a single string for display.
func joinConversation(conv []string) string {
	result := ""
	for _, line := range conv {
		result += line + "\n"
	}
	return result
}

package models

import (
	"agent/agent"
	"agent/logger"
	"context"
	"encoding/json"
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

// toolResultMsg is used to deliver tool execution results asynchronously.
type toolResultMsg struct {
	ID     string
	Name   string
	Result string
	Err    error
}

type ToolStatus struct {
	Name   string
	Status string // "pending", "done", "error"
	Result string
	Err    error
}

// toolRequest is used to dispatch tool requests asynchronously.
type toolRequest struct {
	ID   string
	Name string
	Args map[string]interface{}
}

// MainModel is the root model for the Bubbletea application.
type MainModel struct {
	chat               *chatModel
	codeview           *codeviewModel
	sidebar            *sidebarModel
	Agent              *agent.Agent
	conversation       []string // Conversation history as plain strings for now
	quitting           bool
	waitingForClaude   bool
	width              int    // Terminal width
	height             int    // Terminal height
	focusedPane        string // "sidebar" or "chat"
	sidebarShowingFile bool
	inFlightTools      map[string]ToolStatus // Track running tool commands
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
	m.inFlightTools = make(map[string]ToolStatus)

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
			m.sidebarShowingFile = false
			return m, nil
		}
		if m.sidebarShowingFile {
			// Scroll codeview with j/k, arrow keys, and mouse wheel
			var scrollCmd tea.Cmd
			switch msg.String() {
			case "j", "down":
				m.codeview.viewport.LineDown(1)
			case "k", "up":
				m.codeview.viewport.LineUp(1)
			}
			return m, scrollCmd
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
				logger.LogMessage("User", input)
				m.waitingForClaude = true
				return m, m.sendToClaude(input)
			}
		}
	case tea.MouseMsg:
		if m.sidebarShowingFile {
			// Forward mouse wheel events to codeview viewport
			_, cmd := m.codeview.viewport.Update(msg)
			return m, cmd
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
			logger.LogMessage("Claude (error)", msg.Err.Error())
		} else {
			m.conversation = append(m.conversation, "Claude: "+msg.Text)
			m.chat.AddMessage("Claude", msg.Text)
			logger.LogMessage("Claude", msg.Text)
		}
	case toolRequest:
		// Mark as pending
		m.inFlightTools[msg.ID] = ToolStatus{Name: msg.Name, Status: "pending"}
		return m, m.executeToolAsync(msg)
	case toolResultMsg:
		status := "done"
		if msg.Err != nil {
			status = "error"
		}
		m.inFlightTools[msg.ID] = ToolStatus{Name: msg.Name, Status: status, Result: msg.Result, Err: msg.Err}
		if msg.Err == nil {
			// Show result in codeview if relevant (for read_file, edit_file, list_files, etc.)
			if m.codeview != nil {
				m.codeview.OpenTab(msg.ID, msg.Result)
				m.sidebarShowingFile = true
			}
			// Always send tool result back to Claude for follow-up
			return m, m.sendToClaude(msg.Result)
		}
		if msg.Err != nil && m.codeview != nil {
			m.codeview.OpenTab(msg.ID, "[ERROR] "+msg.Err.Error())
			m.sidebarShowingFile = true
		}
		return m, nil
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
		// If there are tool uses, dispatch toolRequest for the first one
		if len(resp.ToolUses) > 0 {
			tu := resp.ToolUses[0]
			return toolRequest{ID: tu.ID, Name: tu.Name, Args: tu.Args}
		}
		// Otherwise, return the first text response (if any)
		if len(resp.Texts) > 0 {
			return claudeResponseMsg{Text: resp.Texts[0]}
		}
		return claudeResponseMsg{Text: "[No Claude response]"}
	}
}

// executeToolAsync executes a tool asynchronously and returns a command that will deliver the result.
func (m *MainModel) executeToolAsync(req toolRequest) tea.Cmd {
	return func() tea.Msg {
		if m.Agent == nil {
			return toolResultMsg{ID: req.ID, Name: req.Name, Err: context.DeadlineExceeded}
		}
		// Marshal args to JSON
		input, err := json.Marshal(req.Args)
		if err != nil {
			return toolResultMsg{ID: req.ID, Name: req.Name, Err: err}
		}
		// Call agent's executeTool (returns string, error)
		result, err := m.Agent.ExecuteTool(req.Name, input)
		return toolResultMsg{ID: req.ID, Name: req.Name, Result: result, Err: err}
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

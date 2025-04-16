package models

import (
	tea "github.com/charmbracelet/bubbletea"
	"context"
	"agent/agent"
	"github.com/charmbracelet/lipgloss"
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
	width       int // Terminal width
	height      int // Terminal height
	focusedPane string // "sidebar" or "chat"
}

func (m *MainModel) Init() tea.Cmd {
	m.chat = newChatModel()
	m.conversation = []string{}
	m.waitingForClaude = false
	m.sidebar = newSidebarModelFromDir(".")
	m.focusedPane = "chat"

	cmds := []tea.Cmd{
		tea.EnterAltScreen,
	}
	
	// Get initial window size
	return tea.Batch(cmds...)
}

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
		
		// Update both panels with their respective sizes
		if m.chat != nil {
			m.chat.updateSize(rightPanelWidth, m.height - 2) // Reserve space for top margin
		}
		if m.sidebar != nil {
			m.sidebar.updateSize(leftPanelWidth, m.height - 2) // Reserve space for top margin
		}
		return m, nil
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			m.quitting = true
			return m, tea.Quit
		}
		if msg.Type == tea.KeyTab {
			if m.focusedPane == "chat" {
				m.focusedPane = "sidebar"
			} else {
				m.focusedPane = "chat"
			}
			return m, nil
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
	// Forward input to focused pane
	if m.focusedPane == "sidebar" && m.sidebar != nil {
		m.sidebar.Update(msg)
		return m, nil
	}
	if m.focusedPane == "chat" && m.chat != nil {
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
	
	// Calculate usable height (leave some space at the top and bottom)
	usableHeight := m.height - 4 // Reserve space for top margin and borders

	// Create left panel: sidebar (if present) + codeview (if present)
	var leftPanel string
	if m.sidebar != nil {
		leftPanel += m.sidebar.View()
	}
	if m.codeview != nil {
		if leftPanel != "" {
			leftPanel += "\n"
		}
		leftPanel += "[CodeView]"
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

	// Highlight focused pane with properly sized borders
	// Use a simpler approach with fewer calculations
	leftStyle := lipgloss.NewStyle().
		Width(leftPanelWidth - 2). // Account for border width
		Height(usableHeight)       // Set a fixed height
		
	rightStyle := lipgloss.NewStyle().
		Width(rightPanelWidth - 2). // Account for border width
		Height(usableHeight)        // Set a fixed height
	
	// Apply borders based on focus
	if m.focusedPane == "sidebar" {
		leftStyle = leftStyle.
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("69")).
			Padding(0, 1)
	}
	
	if m.focusedPane == "chat" {
		rightStyle = rightStyle.
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("69")).
			Padding(0, 1)
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

func joinConversation(conv []string) string {
	result := ""
	for _, line := range conv {
		result += line + "\n"
	}
	return result
}

type codeviewModel struct{}

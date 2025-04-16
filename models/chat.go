package models

import (
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
	}
}

func (c *chatModel) Init() tea.Cmd {
	return nil
}

func (c *chatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	c.textarea, cmd = c.textarea.Update(msg)
	c.viewport, _ = c.viewport.Update(msg)
	return c, cmd
}

func (c *chatModel) View() string {
	return chatBorderStyle.Render(
		c.viewport.View()+"\n"+c.textarea.View(),
	)
}

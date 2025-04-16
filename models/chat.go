package models

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
)

type chatModel struct {
	textarea textarea.Model
	viewport viewport.Model
}

func newChatModel() *chatModel {
	return &chatModel{
		textarea: textarea.New(),
		viewport: viewport.New(80, 20),
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
	return c.viewport.View() + "\n" + c.textarea.View()
}

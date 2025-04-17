package models

import (
	"github.com/charmbracelet/bubbles/viewport"
)

// codeviewModel is responsible for displaying file contents and managing open file tabs.
type codeviewModel struct {
	viewport  viewport.Model // The viewport for displaying file contents
	tabs      []string       // Slice of open file paths (or tab names)
	activeTab int           // Index of the currently active tab
}

// NewCodeViewModel creates a new codeviewModel with default settings.
func NewCodeViewModel(width, height int) *codeviewModel {
	vp := viewport.New(width, height)
	return &codeviewModel{
		viewport:  vp,
		tabs:      []string{},
		activeTab: 0,
	}
}

// SetTabs sets the open tabs and active tab index.
func (m *codeviewModel) SetTabs(tabs []string, active int) {
	m.tabs = tabs
	if active >= 0 && active < len(tabs) {
		m.activeTab = active
	}
}

// SetFileContent sets the content of the viewport for the active tab.
func (m *codeviewModel) SetFileContent(content string) {
	m.viewport.SetContent(content)
}

// ActiveTab returns the name of the currently active tab.
func (m *codeviewModel) ActiveTab() string {
	if len(m.tabs) == 0 || m.activeTab >= len(m.tabs) {
		return ""
	}
	return m.tabs[m.activeTab]
}

// OpenTab adds a tab (if not present), sets it active, and sets its content.
func (m *codeviewModel) OpenTab(filename, content string) {
	// Check if tab already exists
	idx := -1
	for i, t := range m.tabs {
		if t == filename {
			idx = i
			break
		}
	}
	if idx == -1 {
		m.tabs = append(m.tabs, filename)
		m.activeTab = len(m.tabs) - 1
	} else {
		m.activeTab = idx
	}
	m.viewport.SetContent(content)
}

// View renders the code view (viewport + tabs).
func (m *codeviewModel) View() string {
	if len(m.tabs) == 0 {
		return "" // Nothing to show if no file is open
	}
	return m.viewport.View()
}

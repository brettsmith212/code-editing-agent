package models

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// fileItem represents a file or directory in the sidebar.
type fileItem struct {
	name string
}

// Title returns the name of the file item.
func (f fileItem) Title() string       { return f.name }
// Description returns an empty string for file items.
func (f fileItem) Description() string { return "" }
// FilterValue returns the name of the file item for filtering purposes.
func (f fileItem) FilterValue() string { return f.name }

// CompactDelegate renders a single line with no extra spacing and highlights the selected item.
type CompactDelegate struct{}

// Height returns the height of the delegate.
func (d CompactDelegate) Height() int                               { return 1 }
// Spacing returns the spacing of the delegate.
func (d CompactDelegate) Spacing() int                              { return 0 }
// Update handles updates to the delegate.
func (d CompactDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
// Render renders the delegate.
func (d CompactDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	name := item.(fileItem).name
	style := lipgloss.NewStyle()
	if index == m.Index() {
		style = style.Bold(true).Foreground(lipgloss.Color(SidebarHighlightColor))
	}
	io.WriteString(w, style.Render(name))
}

// sidebarModel holds state for the sidebar panel.
type sidebarModel struct {
	list   list.Model
	width  int
	height int
}

// newSidebarModelFromDir creates a new sidebarModel and loads files from the given directory.
func newSidebarModelFromDir(dir string) *sidebarModel {
	entries, err := os.ReadDir(dir)
	items := make([]list.Item, 0, len(entries))
	if err == nil {
		for _, entry := range entries {
			name := entry.Name()
			if entry.IsDir() {
				name += "/"
			}
			items = append(items, fileItem{name: name})
		}
	} else {
		items = append(items, fileItem{name: fmt.Sprintf("Error: %v", err)})
	}

	l := list.New(items, CompactDelegate{}, LeftPanelInitialWidth, LeftPanelInitialHeight)
	l.Title = SidebarTitle
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(SidebarHighlightColor))

	return &sidebarModel{
		list:   l,
		width:  LeftPanelInitialWidth,
		height: LeftPanelInitialHeight,
	}
}

// updateSize updates the sidebar dimensions based on terminal size.
func (m *sidebarModel) updateSize(width, height int) {
	m.width = width
	m.height = height

	// Adjust width for padding and borders - use the same calculation as chat viewport
	contentWidth := width - LeftPanelPaddingWidth // Account for padding and borders
	if contentWidth < LeftPanelMinContentWidth {    // Minimum reasonable width
		contentWidth = LeftPanelMinContentWidth
	}

	// Set height to fill available space, matching chat viewport's calculation
	viewportHeight := height
	if viewportHeight < LeftPanelMinHeight { // Minimum reasonable height
		viewportHeight = LeftPanelMinHeight
	}

	// Update the list size to fill the full available space
	m.list.SetSize(contentWidth, viewportHeight)
}

// Update handles Bubbletea messages for the sidebar panel.
func (m *sidebarModel) Update(msg any) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.updateSize(msg.Width, msg.Height)
	case tea.KeyMsg:
		if msg.Type == tea.KeyEnter {
			item := m.list.SelectedItem()
			if file, ok := item.(fileItem); ok {
				return func() tea.Msg {
					return openFileMsg{FileName: file.name}
				}
			}
		}
	}
	l, cmd := m.list.Update(msg)
	m.list = l
	return cmd
}

// View renders the sidebar panel with a custom title and vertical padding.
func (m *sidebarModel) View() string {
	// Make a custom wrapper to ensure the title always appears
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(SidebarHighlightColor))
	title := titleStyle.Render(SidebarTitle)

	// Get the list content without the default title
	content := m.list.View()

	// If the view already has a title, we need to remove it to avoid duplication
	lines := strings.Split(content, "\n")
	if len(lines) > 0 && strings.Contains(lines[0], SidebarTitle) {
		// Remove the first line (title) and join the rest
		content = strings.Join(lines[1:], "\n")
	}

	// Calculate available height after title
	availableHeight := m.height

	// Pad the content vertically to match chat height exactly
	contentLines := strings.Split(content, "\n")

	// Calculate how many lines to pad to match chat height
	// Account for title line and a bit of padding
	paddingNeeded := availableHeight - len(contentLines)
	if paddingNeeded > 0 {
		// Add padding at the bottom
		bottomPadding := strings.Repeat("\n", paddingNeeded)
		content = content + bottomPadding
	}

	// Combine title and content with consistent styling
	return title + "\n" + content
}

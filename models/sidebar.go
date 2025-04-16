package models

import (
	"os"
	"fmt"
	"io"
	"strings"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

type fileItem struct {
	name string
}

func (f fileItem) Title() string       { return f.name }
func (f fileItem) Description() string { return "" }
func (f fileItem) FilterValue() string { return f.name }

// CompactDelegate renders a single line with no extra spacing
// and highlights the selected item.
type CompactDelegate struct{}

func (d CompactDelegate) Height() int          { return 1 }
func (d CompactDelegate) Spacing() int         { return 0 }
func (d CompactDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d CompactDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	name := item.(fileItem).name
	style := lipgloss.NewStyle()
	if index == m.Index() {
		style = style.Bold(true).Foreground(lipgloss.Color("69"))
	}
	io.WriteString(w, style.Render(name))
}

type sidebarModel struct {
	list   list.Model
	width  int
	height int
}

func newSidebarModelFromDir(dir string) *sidebarModel {
	entries, err := os.ReadDir(dir)
	items := make([]list.Item, 0, len(entries))
	if err == nil {
		for _, entry := range entries {
			items = append(items, fileItem{name: entry.Name()})
		}
	} else {
		items = append(items, fileItem{name: fmt.Sprintf("Error: %v", err)})
	}
	
	// Initial default size, will be updated on WindowSizeMsg
	initialWidth := 30
	initialHeight := 20
	
	l := list.New(items, CompactDelegate{}, initialWidth, initialHeight)
	l.Title = "Files"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("69"))
	
	return &sidebarModel{
		list:   l,
		width:  initialWidth,
		height: initialHeight,
	}
}

// updateSize updates the sidebar dimensions based on terminal size
func (s *sidebarModel) updateSize(width, height int) {
	s.width = width
	s.height = height
	
	// Adjust width for padding and borders - use the same calculation as chat viewport
	contentWidth := width - 6 // Account for padding and borders
	if contentWidth < 20 {    // Minimum reasonable width
		contentWidth = 20
	}
	
	// Set height to fill available space, matching chat viewport's calculation
	// Account for title and border spacing
	viewportHeight := height
	if viewportHeight < 5 {      // Minimum reasonable height
		viewportHeight = 5
	}
	
	// Update the list size to fill the full available space
	s.list.SetSize(contentWidth, viewportHeight)
}

func (s *sidebarModel) Update(msg any) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Handle window resize
		s.updateSize(msg.Width, msg.Height)
	default:
		m, _ := s.list.Update(msg)
		s.list = m
	}
}

func (s *sidebarModel) View() string {
	// Make a custom wrapper to ensure the title always appears
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("69"))
	title := titleStyle.Render("Files")
	
	// Get the list content without the default title
	content := s.list.View()
	
	// If the view already has a title, we need to remove it to avoid duplication
	lines := strings.Split(content, "\n")
	if len(lines) > 0 && strings.Contains(lines[0], "Files") {
		// Remove the first line (title) and join the rest
		content = strings.Join(lines[1:], "\n")
	}
	
	// Combine title and content with consistent styling
	return title + "\n" + content
}

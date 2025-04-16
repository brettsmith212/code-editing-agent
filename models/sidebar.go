package models

import (
	"os"
	"fmt"
	"io"
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
	list list.Model
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
	l := list.New(items, CompactDelegate{}, 30, 20)
	l.Title = "Files"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("69"))
	return &sidebarModel{list: l}
}

func (s *sidebarModel) Update(msg any) {
	m, _ := s.list.Update(msg)
	s.list = m
}

func (s *sidebarModel) View() string {
	return s.list.View()
}

package models

import (
	"os"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	"fmt"
)

type fileItem struct {
	name string
}

func (f fileItem) Title() string       { return f.name }
func (f fileItem) Description() string { return "" }
func (f fileItem) FilterValue() string { return f.name }

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
	l := list.New(items, list.NewDefaultDelegate(), 30, 20)
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

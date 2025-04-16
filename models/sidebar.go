package models

import (
	"github.com/charmbracelet/lipgloss"
	"strings"
)

type sidebarModel struct {
	files   []string // List of files
	selected int    // Index of selected file
}

func newSidebarModel(files []string) *sidebarModel {
	return &sidebarModel{
		files: files,
		selected: 0,
	}
}

func (s *sidebarModel) View() string {
	var b strings.Builder
	for i, f := range s.files {
		style := lipgloss.NewStyle()
		if i == s.selected {
			style = style.Bold(true).Foreground(lipgloss.Color("69"))
		}
		b.WriteString(style.Render(f) + "\n")
	}
	return lipgloss.NewStyle().Padding(1, 1).Render(strings.TrimRight(b.String(), "\n"))
}

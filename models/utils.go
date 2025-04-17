package models

import "strings"

// wrapText wraps text at the given width, breaking long words if necessary.
func wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}

	var result strings.Builder
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		if len(line) <= width {
			result.WriteString(line)
			result.WriteString("\n")
			continue
		}

		start := 0
		for start < len(line) {
			end := start + width
			if end > len(line) {
				end = len(line)
			}

			// If we're in the middle of a word, find the last space before 'end'
			if end < len(line) && line[end] != ' ' {
				lastSpace := strings.LastIndex(line[start:end], " ")
				if lastSpace != -1 {
					end = start + lastSpace
				}
			}

			result.WriteString(line[start:end])
			result.WriteString("\n")
			start = end
			for start < len(line) && line[start] == ' ' {
				start++
			}
		}
	}
	return result.String()
}

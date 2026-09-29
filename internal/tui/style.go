package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

var (
	activeTabStyle   = lipgloss.NewStyle().Bold(true).Reverse(true)
	inactiveTabStyle = lipgloss.NewStyle().Faint(true)
	headingStyle     = lipgloss.NewStyle().Bold(true)
	dimStyle         = lipgloss.NewStyle().Faint(true)
	keyStyle         = lipgloss.NewStyle().Bold(true)
	ruleStyle        = lipgloss.NewStyle().Faint(true)
)

func rule(width int) string {
	if width < 1 {
		return ""
	}
	return ruleStyle.Render(strings.Repeat("─", width))
}

func wrap(text string, width int) string {
	if width < 20 {
		width = 20
	}

	var out strings.Builder
	line := 0

	for i, word := range strings.Fields(text) {
		switch {
		case i == 0:
			out.WriteString(word)
			line = len(word)
		case line+1+len(word) > width:
			out.WriteString("\n")
			out.WriteString(word)
			line = len(word)
		default:
			out.WriteString(" ")
			out.WriteString(word)
			line += 1 + len(word)
		}
	}

	return out.String()
}

func truncate(s string, width int) string {
	if width < 1 {
		return ""
	}

	runes := []rune(s)
	if len(runes) <= width {
		return s
	}
	if width == 1 {
		return "…"
	}

	return string(runes[:width-1]) + "…"
}

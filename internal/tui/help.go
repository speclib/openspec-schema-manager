package tui

import "strings"

var globalKeys = []KeyHelp{
	{Key: "tab / shift+tab", Description: "move between tabs"},
	{Key: "1 to 4", Description: "select a tab by number"},
	{Key: "?", Description: "open and close this help"},
	{Key: "q", Description: "close an overlay, or quit"},
	{Key: "ctrl+c", Description: "quit from anywhere"},
}

func renderHelp(current Screen, width int) string {
	var b strings.Builder

	b.WriteString(headingStyle.Render("Help"))
	b.WriteString("\n")
	b.WriteString(rule(width))
	b.WriteString("\n\n")
	b.WriteString(headingStyle.Render("Everywhere"))
	b.WriteString("\n")
	b.WriteString(renderKeys(globalKeys))

	b.WriteString("\n\n")
	b.WriteString(headingStyle.Render(current.Title()))
	b.WriteString("\n")

	if keys := current.Keys(); len(keys) > 0 {
		b.WriteString(renderKeys(keys))
	} else {
		b.WriteString(dimStyle.Render("  This tab adds no keys of its own."))
	}

	b.WriteString("\n\n")
	b.WriteString(rule(width))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("? · esc · q  close this help"))

	return b.String()
}

func renderKeys(keys []KeyHelp) string {
	width := 0
	for _, k := range keys {
		if len(k.Key) > width {
			width = len(k.Key)
		}
	}

	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString("  ")
		b.WriteString(keyStyle.Render(k.Key))
		b.WriteString(strings.Repeat(" ", width-len(k.Key)))
		b.WriteString("  ")
		b.WriteString(k.Description)
	}

	return b.String()
}

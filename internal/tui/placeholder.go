package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
)

type placeholder struct {
	title     string
	milestone string
	holds     string
	keys      []KeyHelp
}

func (p placeholder) Title() string { return p.title }

func (p placeholder) Keys() []KeyHelp { return p.keys }

func (p placeholder) Capturing() bool { return false }

func (p placeholder) Update(tea.Msg) (Screen, tea.Cmd) { return p, nil }

func (p placeholder) View(width, height int) string {
	var b strings.Builder

	b.WriteString(headingStyle.Render(p.title))
	b.WriteString("\n\n")
	b.WriteString(wrap(p.holds, width))
	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render(fmt.Sprintf("Built in milestone %s. Nothing here works yet.", p.milestone)))

	if len(p.keys) > 0 {
		b.WriteString("\n\n")
		b.WriteString(dimStyle.Render("Keys this screen will handle:"))
		b.WriteString("\n")
		b.WriteString(renderKeys(p.keys))
	}

	return b.String()
}

func composerScreen() Screen {
	return placeholder{
		title:     "Composer",
		milestone: "07",
		holds:     "Build a standalone schema from artifacts taken out of several others, editing the dependency graph from the keyboard.",
		keys: []KeyHelp{
			{Key: "a", Description: "add the palette selection"},
			{Key: "x", Description: "remove the selected artifact"},
			{Key: "l", Description: "link two artifacts"},
			{Key: "w", Description: "write the schema"},
		},
	}
}

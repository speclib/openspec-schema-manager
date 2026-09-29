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

func projectScreen(inProject bool) Screen {
	holds := "The OpenSpec project you are standing in: its default schema, every schema available to it and where each resolves from, and the schema each change uses."
	if !inProject {
		holds = "Not in an OpenSpec project. Browse Registry or Local; installing needs a project. Start ossm inside one, or point it at one with --path."
	}

	return placeholder{
		title:     "Project",
		milestone: "05",
		holds:     holds,
		keys: []KeyHelp{
			{Key: "enter", Description: "open the selected schema"},
			{Key: "s", Description: "set the project default schema"},
			{Key: "u", Description: "update an installed schema"},
		},
	}
}

func localScreen() Screen {
	return placeholder{
		title:     "Local",
		milestone: "06",
		holds:     "Schemas in your configured directories, any folder you open by path, and the ones you opened recently.",
		keys: []KeyHelp{
			{Key: "enter", Description: "open the selected schema"},
			{Key: "e", Description: "edit the selected file"},
			{Key: "c", Description: "duplicate under a new name"},
		},
	}
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

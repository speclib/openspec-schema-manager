package tui

import (
	"fmt"
	"strings"

	"github.com/speclib/openspec-schema-manager/internal/compose"
	"github.com/speclib/openspec-schema-manager/internal/config"
)

// KeyReference renders every screen's keys from the same Keys() the help
// overlay reads, so a written reference cannot drift from what the code does.
func KeyReference() string {
	var b strings.Builder

	b.WriteString("# Keys\n\n")
	b.WriteString("Generated from the code by `go test ./internal/tui -run TestWriteKeyReference -update`.\n")
	b.WriteString("Every screen reports its own keys, and the help overlay inside ossm reads the same list.\n\n")

	b.WriteString("## Everywhere\n\n")
	b.WriteString(keyTable(globalKeys))

	frame := New(Options{Version: "generated", Recents: config.Recents{}, Drafts: compose.Drafts{}})

	for _, screen := range frame.screens {
		b.WriteString("\n## " + screen.Title() + "\n\n")
		b.WriteString(keyTable(screen.Keys()))
	}

	b.WriteString("\n## Modes\n\n")
	b.WriteString("Some screens change what the keys mean while they are taking text or offering a choice.\n")
	b.WriteString("The help overlay follows, so `?` always lists what works where you are standing.\n\n")

	for _, mode := range modeReference() {
		b.WriteString("### " + mode.name + "\n\n")
		b.WriteString(keyTable(mode.keys))
		b.WriteString("\n")
	}

	return b.String()
}

type modeKeys struct {
	name string
	keys []KeyHelp
}

func modeReference() []modeKeys {
	local := newLocalScreen(nil, config.Recents{}, nil)
	local.tree = true
	tree := local.Keys()

	local.tree = false
	local.duplicating = true
	duplicate := local.Keys()

	composer := newComposerScreen(nil, compose.Drafts{}, true)
	composer.mode = composerLinking
	linking := composer.Keys()

	composer.mode = composerDrafts
	drafts := composer.Keys()

	composer.mode = composerNamingWrite
	prompt := composer.Keys()

	return []modeKeys{
		{name: "Local, browsing a schema's files", keys: tree},
		{name: "Local, naming a duplicate", keys: duplicate},
		{name: "Composer, choosing what to link to", keys: linking},
		{name: "Composer, choosing a draft", keys: drafts},
		{name: "Composer, any prompt", keys: prompt},
		{name: "Registry, confirming an install", keys: installKeys()},
		{name: "Anywhere, the path prompt", keys: pathPromptKeys()},
	}
}

func installKeys() []KeyHelp {
	return []KeyHelp{
		{Key: "y", Description: "confirm the install"},
		{Key: "o", Description: "overwrite what is already there"},
		{Key: "n / esc", Description: "cancel"},
	}
}

func pathPromptKeys() []KeyHelp {
	return []KeyHelp{
		{Key: "enter", Description: "open the folder"},
		{Key: "tab", Description: "complete the path"},
		{Key: "esc", Description: "cancel"},
	}
}

func keyTable(keys []KeyHelp) string {
	if len(keys) == 0 {
		return "This screen adds no keys of its own.\n"
	}

	width := len("Key")
	description := len("Does")

	for _, k := range keys {
		if len(k.Key) > width {
			width = len(k.Key)
		}
		if len(k.Description) > description {
			description = len(k.Description)
		}
	}

	var b strings.Builder

	b.WriteString(fmt.Sprintf("| %-*s | %-*s |\n", width, "Key", description, "Does"))
	b.WriteString("| " + strings.Repeat("-", width) + " | " + strings.Repeat("-", description) + " |\n")

	for _, k := range keys {
		b.WriteString(fmt.Sprintf("| %-*s | %-*s |\n", width, k.Key, description, k.Description))
	}

	return b.String()
}

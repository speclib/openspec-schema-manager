package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/speclib/openspec-schema-manager/internal/compose"
	"github.com/speclib/openspec-schema-manager/internal/graph"
	"github.com/speclib/openspec-schema-manager/internal/schema"
)

type SourceAdded struct {
	Schema schema.Schema
	Dir    string
	Ref    string
	Err    error
}

type composerMode int

const (
	composerBrowsing composerMode = iota
	composerNamingAdd
	composerLinking
	composerRenaming
	composerNamingTracks
	composerNamingWrite
	composerNamingDraft
	composerDrafts
)

type paletteEntry struct {
	source   compose.Source
	artifact schema.Artifact
	heading  bool
}

type composerModel struct {
	canvas *compose.Composition

	palette      []paletteEntry
	paletteIndex int
	canvasIndex  int
	onCanvas     bool

	mode          composerMode
	typed         string
	pending       schema.Artifact
	pendingSource compose.Source
	linkIndex     int
	message       string

	diagram      bool
	diagramView  viewport
	diagramDrawn bool
	diagramErr   error

	drafts     compose.Drafts
	draftList  []compose.Listed
	draftIndex int
	dirs       []string
	ascii      bool
}

func newComposerScreen(dirs []string, drafts compose.Drafts, ascii bool) *composerModel {
	return &composerModel{canvas: compose.New(), drafts: drafts, dirs: dirs, ascii: ascii}
}

func (c *composerModel) Title() string { return "Composer" }

func (c *composerModel) Capturing() bool { return c.mode != composerBrowsing }

func (c *composerModel) Keys() []KeyHelp {
	switch c.mode {
	case composerBrowsing:
		return []KeyHelp{
			{Key: "tab is taken", Description: "use left / right to move between the panes"},
			{Key: "a", Description: "add the palette selection to the canvas"},
			{Key: "x", Description: "remove the selected artifact"},
			{Key: "l", Description: "link the selected artifact to another"},
			{Key: "u", Description: "remove a link"},
			{Key: "g", Description: "toggle the apply gate"},
			{Key: "t", Description: "set the tracked file"},
			{Key: "R", Description: "rename the selected artifact"},
			{Key: "d", Description: "show and hide the diagram"},
			{Key: "w", Description: "write the schema"},
			{Key: "s", Description: "save a draft"},
			{Key: "o", Description: "open a draft"},
		}
	case composerLinking:
		return []KeyHelp{
			{Key: "up / down", Description: "choose what to require"},
			{Key: "enter", Description: "add the link"},
			{Key: "esc", Description: "cancel"},
		}
	case composerDrafts:
		return []KeyHelp{
			{Key: "up / down", Description: "choose a draft"},
			{Key: "enter", Description: "resume it"},
			{Key: "esc", Description: "cancel"},
		}
	}

	return []KeyHelp{
		{Key: "enter", Description: "confirm"},
		{Key: "esc", Description: "cancel"},
	}
}

func (c *composerModel) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case SourceAdded:
		if msg.Err != nil {
			c.message = "could not add that schema: " + msg.Err.Error()
			return c, nil
		}
		c.canvas.AddSource(msg.Schema, msg.Dir, msg.Ref)
		c.rebuildPalette()
		c.message = msg.Schema.Name + " added to the palette"
		return c, nil

	case tea.KeyPressMsg:
		c.handleKey(msg)
		return c, nil
	}

	return c, nil
}

func (c *composerModel) rebuildPalette() {
	c.palette = nil

	for _, source := range c.canvas.Sources {
		c.palette = append(c.palette, paletteEntry{source: source, heading: true})
		for _, a := range source.Artifacts {
			c.palette = append(c.palette, paletteEntry{source: source, artifact: a})
		}
	}

	c.clampPalette()
	c.diagramDrawn = false
}

func (c *composerModel) clampPalette() {
	if c.paletteIndex < 0 {
		c.paletteIndex = 0
	}
	if n := len(c.palette); c.paletteIndex >= n {
		c.paletteIndex = n - 1
	}
	if c.paletteIndex < 0 {
		c.paletteIndex = 0
	}
	for c.paletteIndex < len(c.palette) && c.palette[c.paletteIndex].heading {
		c.paletteIndex++
	}
	if c.paletteIndex >= len(c.palette) {
		c.paletteIndex = len(c.palette) - 1
	}
	if c.paletteIndex < 0 {
		c.paletteIndex = 0
	}
}

func (c *composerModel) clampCanvas() {
	if c.canvasIndex < 0 {
		c.canvasIndex = 0
	}
	if n := len(c.canvas.Artifacts); c.canvasIndex >= n {
		c.canvasIndex = n - 1
	}
	if c.canvasIndex < 0 {
		c.canvasIndex = 0
	}
}

func (c *composerModel) selectedPalette() (paletteEntry, bool) {
	if c.paletteIndex < 0 || c.paletteIndex >= len(c.palette) {
		return paletteEntry{}, false
	}
	entry := c.palette[c.paletteIndex]
	if entry.heading {
		return paletteEntry{}, false
	}
	return entry, true
}

func (c *composerModel) selectedArtifact() (compose.Artifact, bool) {
	if c.canvasIndex < 0 || c.canvasIndex >= len(c.canvas.Artifacts) {
		return compose.Artifact{}, false
	}
	return c.canvas.Artifacts[c.canvasIndex], true
}

func (c *composerModel) handleKey(msg tea.KeyPressMsg) {
	key := msg.String()

	if c.mode != composerBrowsing {
		c.handleModeKey(msg, key)
		return
	}

	switch key {
	case "left", "h":
		c.onCanvas = false
	case "right":
		c.onCanvas = true
	case "up", "k":
		c.move(-1)
	case "down", "j":
		c.move(1)
	case "a":
		c.startAdd()
	case "x":
		c.remove()
	case "l":
		c.startLink()
	case "u":
		c.unlink()
	case "g":
		c.toggleGate()
	case "t":
		c.enter(composerNamingTracks, c.canvas.Tracks)
	case "R":
		c.startRename()
	case "d":
		c.diagram = !c.diagram
		c.diagramDrawn = false
	case "w":
		c.enter(composerNamingWrite, "")
	case "s":
		c.enter(composerNamingDraft, "")
	case "o":
		c.openDrafts()
	}
}

func (c *composerModel) enter(mode composerMode, typed string) {
	c.mode = mode
	c.typed = typed
	c.message = ""
}

func (c *composerModel) handleModeKey(msg tea.KeyPressMsg, key string) {
	if c.mode == composerLinking || c.mode == composerDrafts {
		switch key {
		case "esc":
			c.mode = composerBrowsing
		case "up", "k":
			c.moveChoice(-1)
		case "down", "j":
			c.moveChoice(1)
		case "enter":
			c.confirmChoice()
		}
		return
	}

	switch key {
	case "esc":
		c.mode = composerBrowsing
		c.typed = ""
	case "enter":
		c.confirm()
	case "backspace":
		if c.typed != "" {
			runes := []rune(c.typed)
			c.typed = string(runes[:len(runes)-1])
		}
	default:
		if msg.Text != "" {
			c.typed += msg.Text
		}
	}
}

func (c *composerModel) move(by int) {
	if c.onCanvas {
		c.canvasIndex += by
		c.clampCanvas()
		return
	}

	c.paletteIndex += by
	for c.paletteIndex >= 0 && c.paletteIndex < len(c.palette) && c.palette[c.paletteIndex].heading {
		c.paletteIndex += by
		if by == 0 {
			break
		}
	}
	c.clampPalette()
}

func (c *composerModel) moveChoice(by int) {
	switch c.mode {
	case composerLinking:
		c.linkIndex += by
		if c.linkIndex < 0 {
			c.linkIndex = 0
		}
		if n := len(c.linkTargets()); c.linkIndex >= n {
			c.linkIndex = n - 1
		}
		if c.linkIndex < 0 {
			c.linkIndex = 0
		}
	case composerDrafts:
		c.draftIndex += by
		if c.draftIndex < 0 {
			c.draftIndex = 0
		}
		if n := len(c.draftList); c.draftIndex >= n {
			c.draftIndex = n - 1
		}
		if c.draftIndex < 0 {
			c.draftIndex = 0
		}
	}
}

func (c *composerModel) startAdd() {
	entry, ok := c.selectedPalette()
	if !ok {
		return
	}

	if err := c.canvas.Add(entry.source, entry.artifact, ""); err == nil {
		c.after(entry.artifact.ID + " added")
		return
	}

	c.pending = entry.artifact
	c.pendingSource = entry.source
	c.enter(composerNamingAdd, "")
	c.message = entry.artifact.ID + " is already on the canvas; give it another id"
}

func (c *composerModel) remove() {
	a, ok := c.selectedArtifact()
	if !ok {
		return
	}

	if err := c.canvas.Remove(a.ID); err != nil {
		c.message = err.Error()
		return
	}

	c.clampCanvas()
	c.after(a.ID + " removed")
}

func (c *composerModel) linkTargets() []string {
	a, ok := c.selectedArtifact()
	if !ok {
		return nil
	}

	var targets []string
	for _, other := range c.canvas.Artifacts {
		if other.ID == a.ID {
			continue
		}
		targets = append(targets, other.ID)
	}

	return targets
}

func (c *composerModel) startLink() {
	if _, ok := c.selectedArtifact(); !ok {
		return
	}
	if len(c.linkTargets()) == 0 {
		c.message = "there is nothing else on the canvas to link to"
		return
	}

	c.linkIndex = 0
	c.mode = composerLinking
	c.message = ""
}

func (c *composerModel) unlink() {
	a, ok := c.selectedArtifact()
	if !ok || len(a.Requires) == 0 {
		return
	}

	last := a.Requires[len(a.Requires)-1]
	if err := c.canvas.Unlink(a.ID, last); err != nil {
		c.message = err.Error()
		return
	}

	c.after(a.ID + " no longer requires " + last)
}

func (c *composerModel) toggleGate() {
	a, ok := c.selectedArtifact()
	if !ok {
		return
	}

	if err := c.canvas.ToggleGate(a.ID); err != nil {
		c.message = err.Error()
		return
	}

	if c.canvas.IsGate(a.ID) {
		c.after(a.ID + " gates apply")
	} else {
		c.after(a.ID + " no longer gates apply")
	}
}

func (c *composerModel) startRename() {
	a, ok := c.selectedArtifact()
	if !ok {
		return
	}

	c.enter(composerRenaming, a.ID)
}

func (c *composerModel) openDrafts() {
	c.draftList = c.drafts.List()
	c.draftIndex = 0

	if len(c.draftList) == 0 {
		c.message = "there are no saved drafts"
		return
	}

	c.mode = composerDrafts
	c.message = ""
}

func (c *composerModel) confirmChoice() {
	switch c.mode {
	case composerLinking:
		targets := c.linkTargets()
		if c.linkIndex < 0 || c.linkIndex >= len(targets) {
			c.mode = composerBrowsing
			return
		}

		a, _ := c.selectedArtifact()
		if err := c.canvas.Link(a.ID, targets[c.linkIndex]); err != nil {
			c.message = err.Error()
		} else {
			c.after(a.ID + " requires " + targets[c.linkIndex])
		}
		c.mode = composerBrowsing

	case composerDrafts:
		if c.draftIndex < 0 || c.draftIndex >= len(c.draftList) {
			c.mode = composerBrowsing
			return
		}

		listed := c.draftList[c.draftIndex]
		resumed, err := c.drafts.Resume(listed.Path)
		if err != nil {
			c.message = "could not open that draft: " + err.Error()
			c.mode = composerBrowsing
			return
		}

		c.canvas = resumed.Composition
		c.rebuildPalette()
		c.canvasIndex = 0
		c.mode = composerBrowsing

		if len(resumed.Missing) > 0 {
			c.message = "resumed, but " + strings.Join(resumed.Missing, "; ")
		} else {
			c.message = "resumed " + listed.Name
		}
	}
}

func (c *composerModel) confirm() {
	mode := c.mode
	typed := c.typed

	c.mode = composerBrowsing
	c.typed = ""

	switch mode {
	case composerNamingAdd:
		if err := c.canvas.Add(c.pendingSource, c.pending, typed); err != nil {
			c.message = err.Error()
			return
		}
		c.after(strings.TrimSpace(typed) + " added")

	case composerRenaming:
		a, ok := c.selectedArtifact()
		if !ok {
			return
		}
		if err := c.canvas.Rename(a.ID, typed); err != nil {
			c.message = err.Error()
			return
		}
		c.after("renamed to " + strings.TrimSpace(typed))

	case composerNamingTracks:
		c.canvas.SetTracks(typed)
		c.after("tracking " + c.canvas.Tracks)

	case composerNamingWrite:
		c.write(typed)

	case composerNamingDraft:
		if err := c.drafts.Save(typed, c.canvas); err != nil {
			c.message = "could not save the draft: " + err.Error()
			return
		}
		c.message = "saved as a draft"
	}
}

func (c *composerModel) write(name string) {
	if len(c.dirs) == 0 {
		c.message = "no local schemas directory is configured; set schemas_dirs in config.yml"
		return
	}

	destination, err := c.canvas.Write(c.dirs[0], name)
	if err != nil {
		c.message = "could not write: " + err.Error()
		return
	}

	c.message = "written to " + destination
}

func (c *composerModel) after(message string) {
	c.message = message
	c.diagramDrawn = false
	c.clampCanvas()
}

func (c *composerModel) View(width, height int) string {
	var b strings.Builder

	b.WriteString(c.header(width))
	b.WriteString("\n")

	switch c.mode {
	case composerLinking:
		b.WriteString(c.linkView(width))
		return b.String()
	case composerDrafts:
		b.WriteString(c.draftsView(width))
		return b.String()
	case composerBrowsing:
	default:
		b.WriteString(c.promptView(width))
		return b.String()
	}

	if c.diagram {
		b.WriteString(c.diagramPane(width, height-3))
		return b.String()
	}

	b.WriteString(c.panes(width, height-6))
	b.WriteString("\n")
	b.WriteString(c.findings(width))

	return b.String()
}

func (c *composerModel) header(width int) string {
	parts := []string{fmt.Sprintf("Composer · %d artifact(s)", len(c.canvas.Artifacts))}

	if len(c.canvas.Gates) > 0 {
		parts = append(parts, "gate: "+strings.Join(c.canvas.Gates, ", "))
	} else {
		parts = append(parts, "no gate")
	}

	if c.canvas.Tracks != "" {
		parts = append(parts, "tracks: "+c.canvas.Tracks)
	} else {
		parts = append(parts, "nothing tracked")
	}

	head := headingStyle.Render(truncate(strings.Join(parts, "  ·  "), width))
	if c.message != "" {
		head += "\n" + truncate(c.message, width)
	}

	return head
}

func (c *composerModel) panes(width, height int) string {
	if height < 3 {
		height = 3
	}

	left := width / 2
	if left < 20 {
		left = 20
	}
	right := width - left - 3
	if right < 20 {
		right = 20
	}

	paletteLines := c.paletteLines(left, height)
	canvasLines := c.canvasLines(right, height)

	rows := len(paletteLines)
	if len(canvasLines) > rows {
		rows = len(canvasLines)
	}

	var b strings.Builder
	for i := range rows {
		var l, r string
		if i < len(paletteLines) {
			l = paletteLines[i]
		}
		if i < len(canvasLines) {
			r = canvasLines[i]
		}

		b.WriteString(pad(l, left) + dimStyle.Render(" │ ") + r)
		if i < rows-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (c *composerModel) paletteLines(width, height int) []string {
	lines := []string{dimStyle.Render("Palette")}

	if len(c.palette) == 0 {
		return append(lines, dimStyle.Render("press p on a schema"), dimStyle.Render("to add it as a source"))
	}

	start := 0
	if c.paletteIndex >= height-1 {
		start = c.paletteIndex - height + 2
	}

	for i := start; i < len(c.palette) && len(lines) < height; i++ {
		entry := c.palette[i]

		if entry.heading {
			lines = append(lines, dimStyle.Render(truncate(entry.source.Name, width)))
			continue
		}

		marker := "   "
		if !c.onCanvas && i == c.paletteIndex {
			marker = " ▸ "
		}

		line := marker + entry.artifact.ID
		if c.canvas.Has(entry.artifact.ID) {
			line += "  (on canvas)"
		}

		if !c.onCanvas && i == c.paletteIndex {
			lines = append(lines, keyStyle.Render(truncate(line, width)))
		} else {
			lines = append(lines, truncate(line, width))
		}
	}

	return lines
}

func (c *composerModel) canvasLines(width, height int) []string {
	lines := []string{dimStyle.Render("Canvas")}

	if len(c.canvas.Artifacts) == 0 {
		return append(lines, dimStyle.Render("nothing added yet"))
	}

	start := 0
	if c.canvasIndex >= height-1 {
		start = c.canvasIndex - height + 2
	}

	for i := start; i < len(c.canvas.Artifacts) && len(lines) < height; i++ {
		a := c.canvas.Artifacts[i]

		marker := "  "
		if c.onCanvas && i == c.canvasIndex {
			marker = "▸ "
		}

		line := marker + a.ID
		if c.canvas.IsGate(a.ID) {
			line += " ◆"
		}
		if len(a.Requires) > 0 {
			line += "  ← " + strings.Join(a.Requires, ", ")
		}

		if c.onCanvas && i == c.canvasIndex {
			lines = append(lines, keyStyle.Render(truncate(line, width)))
		} else {
			lines = append(lines, truncate(line, width))
		}
	}

	return lines
}

func (c *composerModel) findings(width int) string {
	var b strings.Builder

	findings := c.canvas.Findings("composed")

	if fatal := findings.Fatal(); len(fatal) > 0 {
		b.WriteString(headingStyle.Render(fmt.Sprintf("%d problem(s)", len(fatal))))
		for i, f := range fatal {
			if i == 3 {
				b.WriteString("\n  " + dimStyle.Render(fmt.Sprintf("and %d more", len(fatal)-3)))
				break
			}
			b.WriteString("\n  " + truncate(describe(f), width-2))
		}
	}

	warnings := c.canvas.Scan()
	if len(warnings) > 0 {
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		b.WriteString(dimStyle.Render(fmt.Sprintf("%d template warning(s)", len(warnings))))
		for i, w := range warnings {
			if i == 3 {
				b.WriteString("\n  " + dimStyle.Render(fmt.Sprintf("and %d more", len(warnings)-3)))
				break
			}
			text := w.String()
			if w.Message != "" {
				text = w.Artifact + ": " + w.Message
			}
			b.WriteString("\n  " + dimStyle.Render(truncate(text, width-2)))
		}
	}

	if b.Len() == 0 {
		return dimStyle.Render("This composition is ready to write with w.")
	}

	return b.String()
}

func (c *composerModel) promptView(width int) string {
	labels := map[composerMode]string{
		composerNamingAdd:    "add under the id",
		composerRenaming:     "rename to",
		composerNamingTracks: "tracked file",
		composerNamingWrite:  "write as",
		composerNamingDraft:  "save the draft as",
	}

	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(labels[c.mode] + ": " + c.typed + "▏")
	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render("enter confirm · esc cancel"))

	return b.String()
}

func (c *composerModel) linkView(width int) string {
	a, _ := c.selectedArtifact()

	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(a.ID + " requires:")
	b.WriteString("\n")

	for i, target := range c.linkTargets() {
		marker := "  "
		if i == c.linkIndex {
			marker = "▸ "
		}
		line := marker + target
		if i == c.linkIndex {
			b.WriteString("\n" + keyStyle.Render(truncate(line, width)))
		} else {
			b.WriteString("\n" + truncate(line, width))
		}
	}

	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render("enter link · esc cancel"))

	return b.String()
}

func (c *composerModel) draftsView(width int) string {
	var b strings.Builder

	b.WriteString("\nDrafts:\n")

	for i, listed := range c.draftList {
		marker := "  "
		if i == c.draftIndex {
			marker = "▸ "
		}

		line := marker + listed.Name
		if listed.Problem != "" {
			line += "  " + listed.Problem
		}

		if i == c.draftIndex {
			b.WriteString("\n" + keyStyle.Render(truncate(line, width)))
		} else {
			b.WriteString("\n" + truncate(line, width))
		}
	}

	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render("enter resume · esc cancel"))

	return b.String()
}

func (c *composerModel) diagramPane(width, height int) string {
	if !c.diagramDrawn {
		g := graph.Build(c.canvas.Schema("composed"))
		drawn, err := g.Draw(c.ascii)
		c.diagramErr = err
		c.diagramDrawn = true
		if err == nil {
			c.diagramView = newViewport(drawn)
		}
	}

	if c.diagramErr != nil {
		return wrap("This composition cannot be drawn: "+c.diagramErr.Error(), width)
	}

	body := c.diagramView.render(height-1, width)

	hint := "d closes the diagram"
	if !c.diagramView.fits(height-1, width) {
		hint = "arrows scroll · " + hint
	}

	return body + "\n" + dimStyle.Render(hint)
}

package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/speclib/openspec-schema-manager/internal/graph"
	"github.com/speclib/openspec-schema-manager/internal/registry"
	"github.com/speclib/openspec-schema-manager/internal/schema"
	"github.com/speclib/openspec-schema-manager/internal/source"
)

type SchemaResolved struct {
	Row      registry.Row
	Resolved source.Resolved
	Err      error
}

type detailModel struct {
	row      registry.Row
	resolved source.Resolved
	loaded   bool

	err     error
	busy    bool
	message string

	selected int
	top      int

	diagram      bool
	diagramView  viewport
	diagramErr   error
	diagramDrawn bool

	resolver *Resolver
	ascii    bool
}

func newDetail(row registry.Row, resolver *Resolver, ascii bool) *detailModel {
	return &detailModel{row: row, resolver: resolver, busy: true, ascii: ascii}
}

func (d *detailModel) Title() string { return d.row.Name }

func (d *detailModel) Capturing() bool { return false }

func (d *detailModel) Keys() []KeyHelp {
	return []KeyHelp{
		{Key: "up / down", Description: "move through the artifacts, or scroll the diagram"},
		{Key: "left / right", Description: "scroll the diagram sideways"},
		{Key: "d", Description: "show and hide the flow diagram"},
		{Key: "R", Description: "fetch the source again"},
		{Key: "o", Description: "show the source repository"},
		{Key: "esc / q", Description: "back to the list"},
	}
}

func (d *detailModel) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case SchemaResolved:
		d.busy = false
		d.err = msg.Err
		if msg.Err == nil {
			d.resolved = msg.Resolved
			d.loaded = true
			d.diagramDrawn = false
		}
		return d, nil

	case tea.KeyPressMsg:
		return d.handleKey(msg)
	}

	return d, nil
}

func (d *detailModel) handleKey(msg tea.KeyPressMsg) (Screen, tea.Cmd) {
	switch msg.String() {
	case "d":
		d.diagram = !d.diagram
		d.message = ""
	case "o":
		d.message = d.where()
	case "R":
		if d.resolver != nil && !d.busy {
			d.busy = true
			d.message = "fetching"
			return d, d.resolver.ResolveCmd(d.row, true)
		}
	case "up", "k":
		d.move(-1)
	case "down", "j":
		d.move(1)
	case "left", "h":
		if d.diagram {
			d.diagramView.left -= 4
		}
	case "right", "l":
		if d.diagram {
			d.diagramView.left += 4
		}
	case "home", "g":
		d.selected, d.diagramView.top, d.diagramView.left = 0, 0, 0
	}

	return d, nil
}

func (d *detailModel) move(by int) {
	if d.diagram {
		d.diagramView.top += by
		return
	}

	d.selected += by
	if d.selected < 0 {
		d.selected = 0
	}
	if n := len(d.resolved.Schema.Artifacts); d.selected >= n {
		d.selected = n - 1
	}
	if d.selected < 0 {
		d.selected = 0
	}
}

func (d *detailModel) where() string {
	if d.row.Entry != nil {
		return d.row.Entry.Source.Repo
	}
	if d.resolved.Dir != "" {
		return d.resolved.Dir
	}
	return d.row.Path
}

func (d *detailModel) View(width, height int) string {
	var b strings.Builder

	b.WriteString(d.header(width))
	b.WriteString("\n")

	body := height - 3
	if body < 1 {
		body = 1
	}

	switch {
	case d.busy && !d.loaded:
		b.WriteString(dimStyle.Render("Fetching " + d.row.Name + " from " + d.where() + "…"))
	case d.err != nil:
		b.WriteString(wrap("Could not read this schema: "+d.err.Error(), width))
		b.WriteString("\n\n")
		b.WriteString(dimStyle.Render("Press R to try again, or esc to go back."))
	case !d.loaded:
		b.WriteString(dimStyle.Render("Nothing loaded."))
	case d.diagram:
		b.WriteString(d.drawDiagram(width, body))
	default:
		b.WriteString(d.body(width, body))
	}

	return b.String()
}

func (d *detailModel) header(width int) string {
	var parts []string

	if d.loaded {
		s := d.resolved.Schema
		head := s.Name
		if s.Version > 0 {
			head += fmt.Sprintf(" v%d", s.Version)
		}
		parts = append(parts, head)

		if d.row.Entry != nil {
			parts = append(parts, d.row.Entry.Source.Repo+" @ "+d.row.RefLabel())
		} else if d.resolved.Dir != "" {
			parts = append(parts, d.resolved.Dir)
		}
	} else {
		parts = append(parts, d.row.Name)
	}

	if d.busy {
		parts = append(parts, "fetching")
	}
	if d.message != "" {
		parts = append(parts, d.message)
	}

	head := headingStyle.Render(truncate(strings.Join(parts, "  ·  "), width))

	if d.loaded && d.resolved.Schema.Description != "" {
		head += "\n" + dimStyle.Render(truncate(d.resolved.Schema.Description, width))
	}

	return head
}

func (d *detailModel) body(width, height int) string {
	s := d.resolved.Schema
	m := d.resolved.Metrics

	var b strings.Builder

	b.WriteString(dimStyle.Render(fmt.Sprintf(
		"%d artifacts · longest chain %d · %d gate(s) · %d leaves",
		m.Artifacts, m.LongestChain, m.Gates, m.Leaves,
	)))
	b.WriteString("\n")

	gate := "nothing"
	if len(s.Apply.Requires) > 0 {
		gate = strings.Join(s.Apply.Requires, ", ")
	}
	tracks := s.Apply.Tracks
	if tracks == "" {
		tracks = "nothing"
	}
	b.WriteString(dimStyle.Render("apply gate: " + gate + " · tracks: " + tracks))
	b.WriteString("\n\n")

	rows := height - 4
	if rows < 1 {
		rows = 1
	}
	b.WriteString(d.artifacts(width, rows))

	if findings := d.findings(width); findings != "" {
		b.WriteString("\n\n")
		b.WriteString(findings)
	}

	return b.String()
}

func (d *detailModel) artifacts(width, rows int) string {
	list := d.resolved.Schema.Artifacts
	if len(list) == 0 {
		return dimStyle.Render("This schema declares no artifacts.")
	}

	idWidth, genWidth := 2, 9
	for _, a := range list {
		if n := len(a.ID); n > idWidth {
			idWidth = n
		}
		if n := len(a.Generates); n > genWidth {
			genWidth = n
		}
	}
	if idWidth > 20 {
		idWidth = 20
	}
	if genWidth > 24 {
		genWidth = 24
	}

	if d.selected < d.top {
		d.top = d.selected
	}
	if d.selected >= d.top+rows {
		d.top = d.selected - rows + 1
	}
	if d.top < 0 {
		d.top = 0
	}

	end := d.top + rows
	if end > len(list) {
		end = len(list)
	}

	var b strings.Builder

	b.WriteString(dimStyle.Render("  " + pad("id", idWidth) + "  " + pad("generates", genWidth) + "  requires"))
	b.WriteString("\n")

	for i := d.top; i < end; i++ {
		a := list[i]

		requires := "nothing"
		if len(a.Requires) > 0 {
			requires = strings.Join(a.Requires, ", ")
		}

		marker := "  "
		if i == d.selected {
			marker = "▸ "
		}

		line := marker + pad(truncate(a.ID, idWidth), idWidth) + "  " +
			pad(truncate(a.Generates, genWidth), genWidth) + "  " + requires

		if i == d.selected {
			b.WriteString(keyStyle.Render(truncate(line, width)))
		} else {
			b.WriteString(truncate(line, width))
		}

		if i < end-1 {
			b.WriteString("\n")
		}
	}

	if selected := list[d.selected]; selected.Template != "" {
		b.WriteString("\n")
		b.WriteString(dimStyle.Render(truncate("template: "+selected.Template, width)))
	}

	return b.String()
}

func (d *detailModel) findings(width int) string {
	findings := d.resolved.Findings
	if len(findings) == 0 {
		return ""
	}

	var b strings.Builder

	if fatal := findings.Fatal(); len(fatal) > 0 {
		b.WriteString(headingStyle.Render(fmt.Sprintf("%d fatal problem(s)", len(fatal))))
		for _, f := range fatal {
			b.WriteString("\n  ")
			b.WriteString(truncate(describe(f), width-2))
		}
	}

	if warnings := findings.Warnings(); len(warnings) > 0 {
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		b.WriteString(dimStyle.Render(fmt.Sprintf("%d warning(s)", len(warnings))))
		for _, f := range warnings {
			b.WriteString("\n  ")
			b.WriteString(dimStyle.Render(truncate(describe(f), width-2)))
		}
	}

	return b.String()
}

func describe(f schema.Finding) string {
	if f.Artifact == "" {
		return f.Message
	}
	return f.Artifact + ": " + f.Message
}

func (d *detailModel) drawDiagram(width, height int) string {
	if !d.diagramDrawn {
		drawn, err := d.resolved.Graph.Draw(d.ascii)
		d.diagramErr = err
		d.diagramDrawn = true
		if err == nil {
			d.diagramView = newViewport(drawn)
		}
	}

	switch {
	case errors.Is(d.diagramErr, graph.ErrCycle):
		return wrap("This schema cannot be drawn: its artifacts require each other in a cycle. "+d.cycleNames(), width)
	case errors.Is(d.diagramErr, graph.ErrNothingToDraw):
		return dimStyle.Render("This schema declares no artifacts, so there is nothing to draw.")
	case d.diagramErr != nil:
		return wrap("This schema could not be drawn: "+d.diagramErr.Error(), width)
	}

	body := d.diagramView.render(height-1, width)

	hint := "d closes the diagram"
	if !d.diagramView.fits(height-1, width) {
		hint = "arrows scroll · " + hint
	}

	return body + "\n" + dimStyle.Render(hint)
}

func (d *detailModel) cycleNames() string {
	for _, f := range d.resolved.Findings.Fatal() {
		if strings.Contains(f.Message, "cycle") {
			return f.Message
		}
	}
	return ""
}

type Resolver struct {
	Fetcher source.Fetcher
	ASCII   bool
}

func (r *Resolver) ResolveCmd(row registry.Row, refetch bool) tea.Cmd {
	return func() tea.Msg {
		resolved, err := r.resolve(row, refetch)
		return SchemaResolved{Row: row, Resolved: resolved, Err: err}
	}
}

func (r *Resolver) resolve(row registry.Row, refetch bool) (source.Resolved, error) {
	if row.Entry == nil {
		if row.Path == "" {
			return source.Resolved{}, errors.New("this schema has neither a source nor a path")
		}
		return source.ResolveDir(row.Path, row.Name)
	}

	if r.Fetcher == nil {
		return source.Resolved{}, errors.New("no way to fetch a schema is configured")
	}

	return source.Resolve(context.Background(), r.Fetcher, source.Source{
		Repo: row.Entry.Source.Repo,
		Path: row.Entry.Source.Path,
		Ref:  row.Entry.Source.Ref,
	}, row.Name, refetch)
}

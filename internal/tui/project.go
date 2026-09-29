package tui

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/speclib/openspec-schema-manager/internal/openspec"
	"github.com/speclib/openspec-schema-manager/internal/registry"
	"github.com/speclib/openspec-schema-manager/internal/source"
)

type ProjectRead struct {
	Project openspec.Project
}

type SchemaCompared struct {
	Name       string
	Comparison source.Comparison
}

type projectModel struct {
	inProject bool
	root      string
	project   openspec.Project
	loaded    bool
	busy      bool
	message   string

	selected int
	top      int

	reader     *ProjectReader
	resolver   *Resolver
	comparer   *Comparer
	detail     *detailModel
	comparison string
}

func newProjectScreen(inProject bool, root string, reader *ProjectReader, resolver *Resolver, comparer *Comparer) *projectModel {
	return &projectModel{inProject: inProject, root: root, reader: reader, resolver: resolver, comparer: comparer}
}

func (p *projectModel) Title() string { return "Project" }

func (p *projectModel) Capturing() bool {
	if p.detail != nil {
		return p.detail.Capturing()
	}
	return false
}

func (p *projectModel) Keys() []KeyHelp {
	if p.detail != nil {
		return p.detail.Keys()
	}

	return []KeyHelp{
		{Key: "up / down", Description: "move through the schemas"},
		{Key: "enter", Description: "open the selected schema"},
		{Key: "s", Description: "set the selected schema as the project default"},
		{Key: "u", Description: "compare with the registry"},
		{Key: "p", Description: "send to the composer"},
		{Key: "r", Description: "read the project again"},
	}
}

func (p *projectModel) Update(msg tea.Msg) (Screen, tea.Cmd) {
	if p.detail != nil {
		if key, ok := msg.(tea.KeyPressMsg); ok {
			switch key.String() {
			case "esc", "q":
				p.detail = nil
				return p, nil
			}
		}

		_, cmd := p.detail.Update(msg)
		return p, cmd
	}

	switch msg := msg.(type) {
	case SchemaCompared:
		p.busy = false
		p.comparison = msg.Comparison.Explain(msg.Name)
		p.message = ""
		return p, nil

	case ProjectRead:
		p.busy = false
		p.loaded = true
		p.project = msg.Project
		p.clampSelection()
		return p, nil

	case InstallFinished:
		if msg.Err == nil && p.reader != nil {
			return p, p.reader.ReadCmd()
		}
		return p, nil

	case tea.KeyPressMsg:
		return p.handleKey(msg)
	}

	return p, nil
}

func (p *projectModel) handleKey(msg tea.KeyPressMsg) (Screen, tea.Cmd) {
	switch msg.String() {
	case "r":
		if p.reader != nil && !p.busy {
			p.busy = true
			p.message = "reading the project"
			return p, p.reader.ReadCmd()
		}
	case "up", "k":
		p.move(-1)
	case "down", "j":
		p.move(1)
	case "home", "g":
		p.selected = 0
	case "end", "G":
		p.selected = len(p.project.Schemas) - 1
		p.clampSelection()
	case "enter":
		return p.open()
	case "s":
		return p.setDefault()
	case "p":
		return p, p.sendToComposer()
	case "u":
		return p.compare()
	}

	return p, nil
}

func (p *projectModel) move(by int) {
	p.selected += by
	p.clampSelection()
}

func (p *projectModel) clampSelection() {
	if p.selected < 0 {
		p.selected = 0
	}
	if n := len(p.project.Schemas); p.selected >= n {
		p.selected = n - 1
	}
	if p.selected < 0 {
		p.selected = 0
	}
}

func (p *projectModel) selectedSchema() *openspec.ResolvedSchema {
	if p.selected < 0 || p.selected >= len(p.project.Schemas) {
		return nil
	}
	s := p.project.Schemas[p.selected]
	return &s
}

func (p *projectModel) open() (Screen, tea.Cmd) {
	s := p.selectedSchema()
	if s == nil || p.resolver == nil {
		return p, nil
	}

	row := registry.RowFromInstalled(s.Name, s.Source, s.Path, nil)
	p.detail = newDetail(row, p.resolver, p.resolver.ASCII)

	return p, p.resolver.ResolveCmd(row, false)
}

func (p *projectModel) compare() (Screen, tea.Cmd) {
	s := p.selectedSchema()
	if s == nil || p.comparer == nil || p.busy {
		return p, nil
	}

	p.busy = true
	p.comparison = ""
	p.message = "comparing " + s.Name + " with the registry"

	return p, p.comparer.CompareCmd(*s)
}

func (p *projectModel) sendToComposer() tea.Cmd {
	s := p.selectedSchema()
	if s == nil {
		return nil
	}

	p.message = "sending " + s.Name + " to the composer"

	return sendToComposerCmd(p.resolver, registry.RowFromInstalled(s.Name, s.Source, s.Path, nil))
}

func (p *projectModel) setDefault() (Screen, tea.Cmd) {
	s := p.selectedSchema()
	if s == nil || !p.inProject {
		return p, nil
	}

	if err := openspec.SetDefaultSchema(p.root, s.Name); err != nil {
		p.message = "could not set the default: " + err.Error()
		return p, nil
	}

	p.message = s.Name + " is now the project default"
	p.project.Default = s.Name

	if p.reader != nil {
		return p, p.reader.ReadCmd()
	}

	return p, nil
}

func (p *projectModel) View(width, height int) string {
	if p.detail != nil {
		return p.detail.View(width, height)
	}

	if !p.inProject {
		return headingStyle.Render("Project") + "\n\n" +
			wrap("There is no OpenSpec project here. Browsing the registry and local schemas works anywhere; installing one needs a project. Start ossm inside a project, or point it at one with --path.", width)
	}

	var b strings.Builder

	b.WriteString(p.header(width))
	b.WriteString("\n\n")

	if !p.loaded {
		b.WriteString(dimStyle.Render("Reading the project…"))
		return b.String()
	}

	b.WriteString(p.schemas(width, height))
	b.WriteString("\n\n")
	b.WriteString(p.changes(width))

	if p.comparison != "" {
		b.WriteString("\n\n")
		b.WriteString(wrap(p.comparison, width))
	}

	for _, problem := range p.project.Problems {
		b.WriteString("\n")
		b.WriteString(dimStyle.Render(truncate(problem, width)))
	}

	return b.String()
}

func (p *projectModel) header(width int) string {
	parts := []string{p.root}

	if p.loaded {
		if p.project.DefaultKnown() {
			parts = append(parts, "default schema: "+p.project.Default)
		} else {
			parts = append(parts, "default schema unknown")
		}
	}
	if p.busy {
		parts = append(parts, "reading")
	}

	head := headingStyle.Render(truncate(strings.Join(parts, "  ·  "), width))

	// The message goes on its own line rather than into the header chain: the
	// project root is often long enough to truncate everything after it, and
	// the message is the part the user just asked for.
	if p.message != "" {
		head += "\n" + truncate(p.message, width)
	}

	return head
}

func (p *projectModel) schemas(width, height int) string {
	if len(p.project.Schemas) == 0 {
		return dimStyle.Render("No schemas resolve for this project.")
	}

	rows := height / 2
	if rows < 3 {
		rows = 3
	}

	if p.selected < p.top {
		p.top = p.selected
	}
	if p.selected >= p.top+rows {
		p.top = p.selected - rows + 1
	}
	if p.top < 0 {
		p.top = 0
	}

	end := p.top + rows
	if end > len(p.project.Schemas) {
		end = len(p.project.Schemas)
	}

	nameWidth := 4
	for _, s := range p.project.Schemas {
		if n := len(s.Name); n > nameWidth {
			nameWidth = n
		}
	}
	if nameWidth > 24 {
		nameWidth = 24
	}

	var b strings.Builder

	b.WriteString(dimStyle.Render("Schemas available"))
	b.WriteString("\n")

	for i := p.top; i < end; i++ {
		s := p.project.Schemas[i]

		marker := "  "
		if i == p.selected {
			marker = "▸ "
		}

		line := marker + pad(truncate(s.Name, nameWidth), nameWidth) + "  " + pad(s.OriginLabel(), 9)

		if s.Name == p.project.Default {
			line += "  (default)"
		}
		if s.Shadowing() {
			line += "  shadows " + strings.Join(s.Shadows, ", ")
		}

		if i == p.selected {
			b.WriteString(keyStyle.Render(truncate(line, width)))
		} else {
			b.WriteString(truncate(line, width))
		}

		if i < end-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (p *projectModel) changes(width int) string {
	var b strings.Builder

	b.WriteString(dimStyle.Render("Changes"))
	b.WriteString("\n")

	if len(p.project.Changes) == 0 {
		b.WriteString(dimStyle.Render("  This project has no changes yet."))
		return b.String()
	}

	nameWidth := 6
	for _, c := range p.project.Changes {
		if n := len(c.Name); n > nameWidth {
			nameWidth = n
		}
	}
	if nameWidth > 28 {
		nameWidth = 28
	}

	for i, c := range p.project.Changes {
		if i > 0 {
			b.WriteString("\n")
		}

		progress := ""
		if c.TotalTasks > 0 {
			progress = fmt.Sprintf("  %d/%d tasks", c.CompletedTasks, c.TotalTasks)
		}

		b.WriteString(truncate("  "+pad(truncate(c.Name, nameWidth), nameWidth)+"  "+c.Schema+progress, width))
	}

	return b.String()
}

type ProjectReader struct {
	CLI  openspec.CLI
	Root string
}

func (r *ProjectReader) Read(ctx context.Context) openspec.Project {
	return openspec.ReadProject(ctx, r.CLI, r.Root)
}

func (r *ProjectReader) ReadCmd() tea.Cmd {
	return func() tea.Msg {
		return ProjectRead{Project: r.Read(context.Background())}
	}
}

type Comparer struct {
	Fetcher source.Fetcher
	Entries func() []registry.Entry
}

func (c *Comparer) CompareCmd(s openspec.ResolvedSchema) tea.Cmd {
	return func() tea.Msg {
		var entries []registry.Entry
		if c.Entries != nil {
			entries = c.Entries()
		}

		return SchemaCompared{
			Name:       s.Name,
			Comparison: source.Compare(context.Background(), c.Fetcher, s.Path, s.BuiltIn(), s.Name, entries),
		}
	}
}

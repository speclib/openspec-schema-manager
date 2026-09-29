package tui

import (
	"fmt"
	"os/exec"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/speclib/openspec-schema-manager/internal/config"
	"github.com/speclib/openspec-schema-manager/internal/registry"
	"github.com/speclib/openspec-schema-manager/internal/schema"
	"github.com/speclib/openspec-schema-manager/internal/source"
)

type EditorFinished struct{ Err error }

type localRow struct {
	local   source.Local
	section string
	missing bool
}

type localModel struct {
	rows     []localRow
	problems []string
	selected int
	top      int
	message  string

	tree      bool
	files     []source.File
	fileIndex int
	findings  schema.Findings
	fileError string

	duplicating bool
	newName     string

	dirs     []string
	recents  config.Recents
	resolver *Resolver
	detail   *detailModel
	editor   func() (string, []string)
}

func newLocalScreen(dirs []string, recents config.Recents, resolver *Resolver) *localModel {
	m := &localModel{dirs: dirs, recents: recents, resolver: resolver, editor: editorFromEnv}
	m.reload()
	return m
}

func (m *localModel) Title() string { return "Local" }

func (m *localModel) Capturing() bool {
	if m.duplicating {
		return true
	}
	if m.detail != nil {
		return m.detail.Capturing()
	}
	return false
}

func (m *localModel) Keys() []KeyHelp {
	switch {
	case m.duplicating:
		return []KeyHelp{
			{Key: "enter", Description: "write the copy"},
			{Key: "esc", Description: "cancel"},
		}
	case m.detail != nil:
		return m.detail.Keys()
	case m.tree:
		return []KeyHelp{
			{Key: "up / down", Description: "move through the files"},
			{Key: "e", Description: "edit the selected file"},
			{Key: "t", Description: "back to the schema list"},
		}
	}

	return []KeyHelp{
		{Key: "up / down", Description: "move through the schemas"},
		{Key: "enter", Description: "open the selected schema"},
		{Key: "t", Description: "browse the schema's files"},
		{Key: "c", Description: "duplicate under a new name"},
		{Key: "p", Description: "send to the composer"},
		{Key: "r", Description: "scan the directories again"},
	}
}

func (m *localModel) reload() {
	m.rows = nil
	m.problems = nil

	for _, scan := range source.ScanAll(m.dirs) {
		if scan.Problem != "" {
			m.problems = append(m.problems, scan.Problem)
			continue
		}
		for _, local := range scan.Schemas {
			m.rows = append(m.rows, localRow{local: local, section: scan.Directory})
		}
	}

	for _, path := range m.recents.Read() {
		local, ok := source.Read(path, "")
		if !ok {
			m.rows = append(m.rows, localRow{
				local:   source.Local{Dir: path},
				section: "Recent",
				missing: true,
			})
			continue
		}
		m.rows = append(m.rows, localRow{local: local, section: "Recent"})
	}

	m.clamp()
}

func (m *localModel) Remember(path string) {
	if _, err := m.recents.Add(path); err != nil {
		m.message = "could not remember that path: " + err.Error()
	}
	m.reload()
	m.selectDir(path)
}

func (m *localModel) selectDir(path string) {
	for i, row := range m.rows {
		if row.local.Dir == path {
			m.selected = i
			return
		}
	}
}

func (m *localModel) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case EditorFinished:
		if msg.Err != nil {
			m.fileError = "the editor reported: " + msg.Err.Error()
		} else {
			m.fileError = ""
		}
		m.revalidate()
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	if m.detail != nil {
		_, cmd := m.detail.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *localModel) handleKey(msg tea.KeyPressMsg) (Screen, tea.Cmd) {
	key := msg.String()

	if m.duplicating {
		switch key {
		case "esc":
			m.duplicating = false
			m.newName = ""
		case "enter":
			m.finishDuplicate()
		case "backspace":
			if m.newName != "" {
				runes := []rune(m.newName)
				m.newName = string(runes[:len(runes)-1])
			}
		default:
			if msg.Text != "" {
				m.newName += msg.Text
			}
		}
		return m, nil
	}

	if m.detail != nil {
		switch key {
		case "esc", "q":
			m.detail = nil
			return m, nil
		}
		_, cmd := m.detail.Update(msg)
		return m, cmd
	}

	if m.tree {
		switch key {
		case "t", "esc":
			m.tree = false
		case "up", "k":
			m.moveFile(-1)
		case "down", "j":
			m.moveFile(1)
		case "e":
			return m, m.edit()
		}
		return m, nil
	}

	switch key {
	case "up", "k":
		m.move(-1)
	case "down", "j":
		m.move(1)
	case "home", "g":
		m.selected = 0
	case "end", "G":
		m.selected = len(m.rows) - 1
		m.clamp()
	case "r":
		m.reload()
		m.message = "scanned again"
	case "enter":
		return m.open()
	case "t":
		m.openTree()
	case "c":
		m.startDuplicate()
	case "p":
		return m, m.sendToComposer()
	}

	return m, nil
}

func (m *localModel) move(by int) {
	m.selected += by
	m.clamp()
}

func (m *localModel) clamp() {
	if m.selected < 0 {
		m.selected = 0
	}
	if n := len(m.rows); m.selected >= n {
		m.selected = n - 1
	}
	if m.selected < 0 {
		m.selected = 0
	}
}

func (m *localModel) moveFile(by int) {
	m.fileIndex += by
	if m.fileIndex < 0 {
		m.fileIndex = 0
	}
	if n := len(m.files); m.fileIndex >= n {
		m.fileIndex = n - 1
	}
	if m.fileIndex < 0 {
		m.fileIndex = 0
	}
}

func (m *localModel) selectedRow() *localRow {
	if m.selected < 0 || m.selected >= len(m.rows) {
		return nil
	}
	row := m.rows[m.selected]
	return &row
}

func (m *localModel) open() (Screen, tea.Cmd) {
	row := m.selectedRow()
	if row == nil || row.missing || m.resolver == nil {
		return m, nil
	}

	listed := registry.RowFromInstalled(row.local.Label(), "user", row.local.Dir, nil)
	m.detail = newDetail(listed, m.resolver, m.resolver.ASCII)

	return m, m.resolver.ResolveCmd(listed, false)
}

func (m *localModel) openTree() {
	row := m.selectedRow()
	if row == nil || row.missing {
		return
	}

	files, err := source.Files(row.local.Dir)
	if err != nil {
		m.message = "could not read that folder: " + err.Error()
		return
	}

	m.tree = true
	m.files = files
	m.fileIndex = 0
	m.fileError = ""
	m.revalidate()
}

func (m *localModel) revalidate() {
	row := m.selectedRow()
	if row == nil {
		return
	}

	s, err := schema.Load(row.local.Dir)
	if err != nil {
		m.findings = nil
		m.fileError = err.Error()
		return
	}

	m.findings = schema.ValidateDir(s, row.local.Dir)

	if files, err := source.Files(row.local.Dir); err == nil {
		m.files = files
		m.moveFile(0)
	}
}

func (m *localModel) edit() tea.Cmd {
	if m.fileIndex < 0 || m.fileIndex >= len(m.files) {
		return nil
	}

	name, args := m.editor()
	cmd := exec.Command(name, append(args, m.files[m.fileIndex].Path)...)

	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return EditorFinished{Err: err}
	})
}

func (m *localModel) sendToComposer() tea.Cmd {
	row := m.selectedRow()
	if row == nil || row.missing {
		return nil
	}

	m.message = "sending " + row.local.Label() + " to the composer"

	return sendToComposerCmd(m.resolver, registry.RowFromInstalled(row.local.Label(), "user", row.local.Dir, nil))
}

func (m *localModel) startDuplicate() {
	row := m.selectedRow()
	if row == nil || row.missing {
		return
	}

	if len(m.dirs) == 0 {
		m.message = "no local schemas directory is configured; set schemas_dirs in config.yml"
		return
	}

	m.duplicating = true
	m.newName = ""
}

func (m *localModel) finishDuplicate() {
	m.duplicating = false

	row := m.selectedRow()
	if row == nil {
		return
	}

	destination, err := source.Duplicate(row.local.Dir, m.dirs[0], m.newName)
	if err != nil {
		m.message = "could not duplicate: " + err.Error()
		m.newName = ""
		return
	}

	m.message = "copied to " + destination
	m.newName = ""
	m.reload()
	m.selectDir(destination)
}

func (m *localModel) View(width, height int) string {
	switch {
	case m.duplicating:
		return m.duplicateView(width)
	case m.detail != nil:
		return m.detail.View(width, height)
	case m.tree:
		return m.treeView(width, height)
	}

	return m.listView(width, height)
}

func (m *localModel) duplicateView(width int) string {
	row := m.selectedRow()

	var b strings.Builder

	b.WriteString(headingStyle.Render("Duplicate " + row.local.Label()))
	b.WriteString("\n\n")
	b.WriteString("new name: " + m.newName + "▏")
	b.WriteString("\n\n")

	if strings.TrimSpace(m.newName) != "" {
		b.WriteString(dimStyle.Render(truncate("into "+m.dirs[0]+"/"+strings.TrimSpace(m.newName), width)))
	} else {
		b.WriteString(dimStyle.Render(truncate("into "+m.dirs[0], width)))
	}

	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render("enter write · esc cancel"))

	return b.String()
}

func (m *localModel) listView(width, height int) string {
	var b strings.Builder

	b.WriteString(m.header(width))
	b.WriteString("\n\n")

	if len(m.rows) == 0 {
		b.WriteString(wrap("No local schemas. Set schemas_dirs in config.yml to point at a directory of schema folders, or press : to open one by path.", width))
	} else {
		b.WriteString(m.list(width, height-4))
	}

	for _, problem := range m.problems {
		b.WriteString("\n")
		b.WriteString(dimStyle.Render(truncate(problem, width)))
	}

	return b.String()
}

func (m *localModel) header(width int) string {
	head := headingStyle.Render(fmt.Sprintf("Local · %d schema(s)", len(m.rows)))
	if m.message != "" {
		head += "\n" + truncate(m.message, width)
	}
	return head
}

func (m *localModel) list(width, rows int) string {
	if rows < 1 {
		rows = 1
	}

	if m.selected < m.top {
		m.top = m.selected
	}
	if m.selected >= m.top+rows {
		m.top = m.selected - rows + 1
	}
	if m.top < 0 {
		m.top = 0
	}

	end := m.top + rows
	if end > len(m.rows) {
		end = len(m.rows)
	}

	var b strings.Builder
	section := ""

	for i := m.top; i < end; i++ {
		row := m.rows[i]

		if row.section != section {
			if b.Len() > 0 {
				b.WriteString("\n")
			}
			b.WriteString(dimStyle.Render(truncate(row.section, width)))
			b.WriteString("\n")
			section = row.section
		}

		marker := "  "
		if i == m.selected {
			marker = "▸ "
		}

		line := marker + row.local.Label()

		switch {
		case row.missing:
			line += "    missing: " + row.local.Dir
		case !row.local.Readable():
			line += "    unreadable"
		default:
			line += fmt.Sprintf("    %d artifacts    %s", row.local.Artifacts, row.local.Dir)
		}

		if i == m.selected {
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

func (m *localModel) treeView(width, height int) string {
	row := m.selectedRow()

	var b strings.Builder

	b.WriteString(headingStyle.Render("Files of " + row.local.Label()))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render(truncate(row.local.Dir, width)))
	b.WriteString("\n\n")

	if len(m.files) == 0 {
		b.WriteString(dimStyle.Render("This folder holds no schema files."))
		return b.String()
	}

	for i, file := range m.files {
		marker := "  "
		if i == m.fileIndex {
			marker = "▸ "
		}

		name := file.Rel
		if file.Depth > 0 {
			name = strings.Repeat("  ", file.Depth) + file.Rel[strings.LastIndex(file.Rel, "/")+1:]
		}

		line := marker + name
		if i == m.fileIndex {
			b.WriteString(keyStyle.Render(truncate(line, width)))
		} else {
			b.WriteString(truncate(line, width))
		}
		b.WriteString("\n")
	}

	if m.fileError != "" {
		b.WriteString("\n")
		b.WriteString(wrap(m.fileError, width))
		return b.String()
	}

	if len(m.findings) == 0 {
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("This schema validates."))
		return b.String()
	}

	b.WriteString("\n")
	if fatal := m.findings.Fatal(); len(fatal) > 0 {
		b.WriteString(headingStyle.Render(fmt.Sprintf("%d fatal problem(s)", len(fatal))))
		for _, f := range fatal {
			b.WriteString("\n  ")
			b.WriteString(truncate(describe(f), width-2))
		}
	}
	if warnings := m.findings.Warnings(); len(warnings) > 0 {
		b.WriteString("\n")
		b.WriteString(dimStyle.Render(fmt.Sprintf("%d warning(s)", len(warnings))))
		for _, f := range warnings {
			b.WriteString("\n  ")
			b.WriteString(dimStyle.Render(truncate(describe(f), width-2)))
		}
	}

	return b.String()
}

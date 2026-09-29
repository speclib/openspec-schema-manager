package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/speclib/openspec-schema-manager/internal/openspec"
	"github.com/speclib/openspec-schema-manager/internal/registry"
)

type RegistryLoaded struct {
	Rows     []registry.Row
	CacheAge time.Duration
	HasCache bool
	Note     string
}

type RegistryRefreshed struct {
	Rows     []registry.Row
	CacheAge time.Duration
	Err      error
}

type registryScreenModel struct {
	rows     []registry.Row
	shown    []registry.Row
	selected int
	top      int

	filtering bool
	filter    string

	hasCache bool
	cacheAge time.Duration
	note     string
	message  string
	busy     bool

	loader   *RegistryLoader
	resolver *Resolver
	detail   *detailModel
	install  *installFlow
}

func newRegistryScreen(loader *RegistryLoader, resolver *Resolver, installer *Installer) Screen {
	return &registryScreenModel{loader: loader, resolver: resolver, install: newInstallFlow(installer)}
}

func (r *registryScreenModel) Title() string { return "Registry" }

func (r *registryScreenModel) Capturing() bool {
	if r.install != nil && r.install.active() {
		return true
	}
	if r.detail != nil {
		return r.detail.Capturing()
	}
	return r.filtering
}

func (r *registryScreenModel) Keys() []KeyHelp {
	if r.install != nil && r.install.active() {
		return []KeyHelp{
			{Key: "y", Description: "confirm the install"},
			{Key: "o", Description: "overwrite what is already there"},
			{Key: "n / esc", Description: "cancel"},
		}
	}

	if r.detail != nil {
		return r.detail.Keys()
	}

	return []KeyHelp{
		{Key: "enter", Description: "open the selected schema"},
		{Key: "i", Description: "install into this project"},
		{Key: "p", Description: "send to the composer"},
		{Key: "up / down", Description: "move the selection"},
		{Key: "/", Description: "filter over id, name and description"},
		{Key: "esc", Description: "clear the filter"},
		{Key: "r", Description: "refresh the registry now"},
	}
}

func (r *registryScreenModel) Update(msg tea.Msg) (Screen, tea.Cmd) {
	if r.install != nil {
		if _, prepared := msg.(InstallPrepared); prepared || r.install.active() {
			return r, r.install.Update(msg)
		}
		if _, finished := msg.(InstallFinished); finished {
			return r, r.install.Update(msg)
		}
	}

	if r.detail != nil {
		if key, ok := msg.(tea.KeyPressMsg); ok {
			switch key.String() {
			case "esc", "q":
				r.detail = nil
				return r, nil
			}
		}

		_, cmd := r.detail.Update(msg)
		return r, cmd
	}

	switch msg := msg.(type) {
	case RegistryLoaded:
		r.rows = msg.Rows
		r.hasCache = msg.HasCache
		r.cacheAge = msg.CacheAge
		r.note = msg.Note
		r.reselect()
		return r, nil

	case RegistryRefreshed:
		r.busy = false
		if msg.Err != nil {
			r.message = "refresh failed: " + msg.Err.Error()
			return r, nil
		}
		r.rows = msg.Rows
		r.hasCache = true
		r.cacheAge = msg.CacheAge
		r.message = "refreshed"
		r.reselect()
		return r, nil

	case tea.KeyPressMsg:
		return r.handleKey(msg)
	}

	return r, nil
}

func (r *registryScreenModel) handleKey(msg tea.KeyPressMsg) (Screen, tea.Cmd) {
	key := msg.String()

	if r.filtering {
		switch key {
		case "esc":
			r.filtering = false
			r.filter = ""
			r.reselect()
		case "enter":
			r.filtering = false
		case "backspace":
			if r.filter != "" {
				r.filter = r.filter[:len(r.filter)-1]
				r.reselect()
			}
		default:
			if text := msg.Text; text != "" {
				r.filter += text
				r.reselect()
			}
		}
		return r, nil
	}

	switch key {
	case "/":
		r.filtering = true
		r.message = ""
	case "esc":
		if r.filter != "" {
			r.filter = ""
			r.reselect()
		}
	case "up", "k":
		r.move(-1)
	case "down", "j":
		r.move(1)
	case "home", "g":
		r.selected = 0
	case "end", "G":
		r.selected = len(r.shown) - 1
	case "r":
		if r.loader != nil && !r.busy {
			r.busy = true
			r.message = "refreshing"
			return r, r.loader.RefreshCmd()
		}
	case "enter":
		return r.open()
	case "i":
		return r.startInstall()
	case "p":
		return r.sendToComposer()
	}

	return r, nil
}

func (r *registryScreenModel) sendToComposer() (Screen, tea.Cmd) {
	row := r.selectedRow()
	if row == nil {
		return r, nil
	}

	r.message = "sending " + row.Name + " to the composer"

	return r, sendToComposerCmd(r.resolver, *row)
}

func (r *registryScreenModel) startInstall() (Screen, tea.Cmd) {
	row := r.selectedRow()
	if row == nil || r.install == nil {
		return r, nil
	}

	cmd, refusal := r.install.start(*row)
	if refusal != "" {
		r.message = refusal
	}

	return r, cmd
}

func (r *registryScreenModel) open() (Screen, tea.Cmd) {
	row := r.selectedRow()
	if row == nil || r.resolver == nil {
		return r, nil
	}

	r.detail = newDetail(*row, r.resolver, r.resolver.ASCII)

	return r, r.resolver.ResolveCmd(*row, false)
}

func (r *registryScreenModel) move(by int) {
	if len(r.shown) == 0 {
		r.selected = 0
		return
	}

	r.selected += by
	if r.selected < 0 {
		r.selected = 0
	}
	if r.selected >= len(r.shown) {
		r.selected = len(r.shown) - 1
	}
}

func (r *registryScreenModel) reselect() {
	previous := r.selectedRow()

	r.shown = registry.Filter(r.rows, r.filter)

	if len(r.shown) == 0 {
		r.selected = 0
		return
	}

	if previous != nil {
		for i, row := range r.shown {
			if sameRow(row, *previous) {
				r.selected = i
				return
			}
		}
	}

	r.selected = 0
}

func (r *registryScreenModel) selectedRow() *registry.Row {
	if r.selected < 0 || r.selected >= len(r.shown) {
		return nil
	}
	row := r.shown[r.selected]
	return &row
}

func sameRow(a, b registry.Row) bool {
	if a.Entry != nil && b.Entry != nil {
		return a.Entry.ID == b.Entry.ID
	}
	if a.Entry != nil || b.Entry != nil {
		return false
	}
	return a.Name == b.Name && a.Origin == b.Origin && a.Path == b.Path
}

func (r *registryScreenModel) View(width, height int) string {
	if r.install != nil && r.install.active() {
		return r.install.View(width)
	}

	if r.detail != nil {
		return r.detail.View(width, height)
	}

	var b strings.Builder

	b.WriteString(r.header(width))
	b.WriteString("\n")

	rows := height - 3
	if rows < 1 {
		rows = 1
	}
	r.scroll(rows)

	switch {
	case len(r.shown) == 0 && r.filter != "":
		b.WriteString(dimStyle.Render(fmt.Sprintf("Nothing matches %q.", r.filter)))
	case len(r.shown) == 0 && r.hasCache:
		b.WriteString(dimStyle.Render("The registry holds no entries."))
	case len(r.shown) > 0:
		b.WriteString(r.list(width, rows))
	}

	for _, note := range r.notes(width) {
		b.WriteString("\n\n")
		b.WriteString(note)
	}

	return b.String()
}

func (r *registryScreenModel) notes(width int) []string {
	var notes []string

	if !r.hasCache && !r.busy {
		notes = append(notes, dimStyle.Render(wrap("The registry has never been fetched. Press r to try again. Everything else in ossm works without it.", width)))
	}
	if r.note != "" {
		notes = append(notes, dimStyle.Render(truncate(r.note, width)))
	}

	return notes
}

func (r *registryScreenModel) header(width int) string {
	parts := []string{fmt.Sprintf("%d schemas", len(r.rows))}

	switch {
	case r.busy:
		parts = append(parts, "refreshing")
	case r.hasCache:
		parts = append(parts, "cached "+humanAge(r.cacheAge)+" ago")
	default:
		parts = append(parts, "never fetched")
	}

	if r.filtering || r.filter != "" {
		parts = append(parts, "filter: "+r.filter+"▏")
	}
	if r.message != "" {
		parts = append(parts, r.message)
	}

	return headingStyle.Render(truncate("Registry · "+strings.Join(parts, " · "), width))
}

func (r *registryScreenModel) scroll(rows int) {
	if r.selected < r.top {
		r.top = r.selected
	}
	if r.selected >= r.top+rows {
		r.top = r.selected - rows + 1
	}
	if r.top < 0 {
		r.top = 0
	}
}

func (r *registryScreenModel) list(width, rows int) string {
	var b strings.Builder

	end := r.top + rows
	if end > len(r.shown) {
		end = len(r.shown)
	}

	nameWidth := 0
	for _, row := range r.shown {
		if len(row.Name) > nameWidth {
			nameWidth = len(row.Name)
		}
	}
	if nameWidth > 24 {
		nameWidth = 24
	}

	for i := r.top; i < end; i++ {
		row := r.shown[i]

		marker := "  "
		if i == r.selected {
			marker = "▸ "
		}

		line := marker + pad(truncate(row.Name, nameWidth), nameWidth) +
			fmt.Sprintf("  %2d artifacts  ", row.ArtifactCount()) +
			pad(truncate(row.Shape(), 28), 28) + "  " +
			row.Origin

		if label := row.RefLabel(); label != "" {
			line += " @ " + label
		}
		if row.Deprecated {
			line += "  deprecated"
			if row.SupersededBy != "" {
				line += " → " + row.SupersededBy
			}
		}

		if i == r.selected {
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

func pad(s string, width int) string {
	if n := width - len([]rune(s)); n > 0 {
		return s + strings.Repeat(" ", n)
	}
	return s
}

func humanAge(d time.Duration) string {
	switch {
	case d < time.Minute:
		return "moments"
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 48*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}

type RegistryLoader struct {
	Store   registry.Store
	Fetcher registry.Fetcher
	CLI     openspec.CLI
	WorkDir string
	TTL     time.Duration
	Now     func() time.Time
}

func (l *RegistryLoader) now() time.Time {
	if l.Now != nil {
		return l.Now()
	}
	return time.Now()
}

func (l *RegistryLoader) installedRows(ctx context.Context) ([]registry.Row, string) {
	if l.CLI == nil {
		return nil, "built-in schemas were not listed: no OpenSpec adapter"
	}

	schemas, err := l.CLI.Schemas(ctx, l.WorkDir)
	if err != nil {
		if errors.Is(err, openspec.ErrNotInstalled) {
			return nil, "built-in schemas were not listed: the openspec CLI is not on PATH"
		}
		return nil, "built-in schemas were not listed: " + err.Error()
	}

	rows := make([]registry.Row, 0, len(schemas))
	for _, s := range schemas {
		rows = append(rows, registry.RowFromInstalled(s.Name, s.Source, s.Path, nil))
	}

	return rows, ""
}

func (l *RegistryLoader) Load(ctx context.Context) RegistryLoaded {
	installed, note := l.installedRows(ctx)

	cached, err := l.Store.Read()
	if err != nil {
		return RegistryLoaded{Rows: registry.Merge(nil, installed), Note: note}
	}

	return RegistryLoaded{
		Rows:     registry.Merge(cached.Registry.Schemas, installed),
		CacheAge: cached.Age(l.now()),
		HasCache: true,
		Note:     note,
	}
}

func (l *RegistryLoader) Stale() bool {
	cached, err := l.Store.Read()
	if err != nil {
		return true
	}
	return cached.Stale(l.now(), l.TTL)
}

func (l *RegistryLoader) RefreshCmd() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		cached, err := l.Store.Refresh(ctx, l.Fetcher)
		if err != nil {
			return RegistryRefreshed{Err: err}
		}

		installed, _ := l.installedRows(ctx)

		return RegistryRefreshed{
			Rows:     registry.Merge(cached.Registry.Schemas, installed),
			CacheAge: cached.Age(l.now()),
		}
	}
}

// Entries hands the registry's entries to whatever needs them, which today is
// the comparison in the Project view.
func (r *registryScreenModel) Entries() []registry.Entry {
	entries := make([]registry.Entry, 0, len(r.rows))
	for _, row := range r.rows {
		if row.Entry != nil {
			entries = append(entries, *row.Entry)
		}
	}
	return entries
}

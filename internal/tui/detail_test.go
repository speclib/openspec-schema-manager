package tui

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/speclib/openspec-schema-manager/internal/ansi"
	"github.com/speclib/openspec-schema-manager/internal/registry"
	"github.com/speclib/openspec-schema-manager/internal/source"
)

type fakeFetcher struct {
	dir     string
	err     error
	calls   int
	refetch bool
}

func (f *fakeFetcher) Fetch(_ context.Context, _ source.Source, refetch bool) (string, error) {
	f.calls++
	f.refetch = refetch

	if f.err != nil {
		return "", f.err
	}

	return f.dir, nil
}

func schemaDir(t *testing.T, fixture string) string {
	t.Helper()

	dir := t.TempDir()

	raw, err := os.ReadFile(filepath.Join("..", "schema", "testdata", fixture))
	if err != nil {
		t.Fatalf("reading the fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "schema.yaml"), raw, 0o644); err != nil {
		t.Fatalf("writing the schema: %v", err)
	}

	for _, template := range []string{"specs/spec.md", "tasks.md", "proposal.md", "design.md", "spec.md", "research.md", "risks.md", "budget.md", "sketch.md", "other.md", "one.md", "two.md", "three.md"} {
		path := filepath.Join(dir, "templates", filepath.FromSlash(template))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("creating %s: %v", filepath.Dir(path), err)
		}
		if err := os.WriteFile(path, []byte("# template\n"), 0o644); err != nil {
			t.Fatalf("writing %s: %v", path, err)
		}
	}

	return dir
}

func registryRow(name string) registry.Row {
	return registry.RowFromEntry(registry.Entry{
		ID:          "speclib/" + name,
		Name:        name,
		Description: "a schema for testing",
		Artifacts:   []string{"specs", "tasks"},
		Source: registry.Source{
			Repo: "https://github.com/speclib/" + name,
			Path: "openspec/schemas/" + name,
		},
	})
}

func openDetail(t *testing.T, row registry.Row, r *Resolver) *detailModel {
	t.Helper()

	d := newDetail(row, r, true)

	cmd := r.ResolveCmd(row, false)
	updated, _ := d.Update(cmd())

	model, ok := updated.(*detailModel)
	if !ok {
		t.Fatalf("Update returned %T", updated)
	}

	return model
}

func detailView(d *detailModel) string { return flat(ansi.Strip(d.View(120, 30))) }

func TestTheDetailShowsTheSchema(t *testing.T) {
	t.Parallel()

	fetcher := &fakeFetcher{dir: schemaDir(t, "chain.yaml")}
	d := openDetail(t, registryRow("minimalist"), &Resolver{Fetcher: fetcher, ASCII: true})

	got := detailView(d)

	for _, want := range []string{
		"minimalist v1",
		"Lightweight schema",
		"github.com/speclib/minimalist",
		"default branch",
		"2 artifacts",
		"longest chain 2",
		"1 gate(s)",
		"apply gate: tasks",
		"tracks: tasks.md",
		"specs",
		"specs/**/*.md",
		"tasks.md",
		"nothing",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("the detail does not carry %q:\n%s", want, got)
		}
	}
}

func TestTheDetailShowsArtifactsInDeclarationOrder(t *testing.T) {
	t.Parallel()

	fetcher := &fakeFetcher{dir: schemaDir(t, "branchy.yaml")}
	d := openDetail(t, registryRow("branchy"), &Resolver{Fetcher: fetcher, ASCII: true})

	got := detailView(d)

	table := got[strings.Index(got, "requires"):]

	last := -1
	for _, id := range []string{"proposal", "specs", "design", "tasks"} {
		at := strings.Index(table, id)
		if at < 0 {
			t.Fatalf("%s is missing:\n%s", id, table)
		}
		if at < last {
			t.Errorf("%s is out of declaration order:\n%s", id, table)
		}
		last = at
	}
}

func TestAPinnedSchemaNamesItsRef(t *testing.T) {
	t.Parallel()

	row := registryRow("pinned")
	row.Entry.Source.Ref = "v0.2.0"
	row.Ref = "v0.2.0"
	row.Pinned = true

	fetcher := &fakeFetcher{dir: schemaDir(t, "chain.yaml")}
	d := openDetail(t, row, &Resolver{Fetcher: fetcher, ASCII: true})

	got := detailView(d)
	if !strings.Contains(got, "@ v0.2.0") {
		t.Errorf("the pinned ref is not shown:\n%s", got)
	}
	if strings.Contains(got, "default branch") {
		t.Errorf("a pinned schema was shown as tracking the default branch:\n%s", got)
	}
}

func TestALocalSchemaIsReadFromItsPathWithoutFetching(t *testing.T) {
	t.Parallel()

	dir := schemaDir(t, "chain.yaml")
	fetcher := &fakeFetcher{dir: dir}

	row := registry.RowFromInstalled("minimalist", "package", dir, nil)
	d := openDetail(t, row, &Resolver{Fetcher: fetcher, ASCII: true})

	if fetcher.calls != 0 {
		t.Errorf("a local schema was fetched %d times", fetcher.calls)
	}

	got := detailView(d)
	if !strings.Contains(got, "minimalist") {
		t.Errorf("the local schema was not read:\n%s", got)
	}
	if !strings.Contains(got, dir) {
		t.Errorf("the path is not shown for a local schema:\n%s", got)
	}
}

func TestALocalSchemaWithNoPathIsReported(t *testing.T) {
	t.Parallel()

	d := openDetail(t, registry.RowFromInstalled("nowhere", "package", "", nil), &Resolver{ASCII: true})

	if got := detailView(d); !strings.Contains(got, "neither a source nor a path") {
		t.Errorf("a row with no path and no source is not reported:\n%s", got)
	}
}

func TestAFetchFailureIsShownWithItsReason(t *testing.T) {
	t.Parallel()

	fetcher := &fakeFetcher{err: source.ErrNoGit}
	d := openDetail(t, registryRow("minimalist"), &Resolver{Fetcher: fetcher, ASCII: true})

	got := detailView(d)
	if !strings.Contains(got, "git is needed") {
		t.Errorf("the reason is not shown:\n%s", got)
	}
	if !strings.Contains(got, "Press R to try again") {
		t.Errorf("no way to retry is offered:\n%s", got)
	}
}

func TestNoFetcherAtAllIsReported(t *testing.T) {
	t.Parallel()

	d := openDetail(t, registryRow("minimalist"), &Resolver{ASCII: true})

	if got := detailView(d); !strings.Contains(got, "no way to fetch") {
		t.Errorf("a missing fetcher is not reported:\n%s", got)
	}
}

func TestFetchingIsReportedWhileItRuns(t *testing.T) {
	t.Parallel()

	d := newDetail(registryRow("minimalist"), &Resolver{Fetcher: &fakeFetcher{}, ASCII: true}, true)

	got := flat(ansi.Strip(d.View(120, 30)))
	if !strings.Contains(got, "Fetching minimalist") {
		t.Errorf("the view does not report that it is fetching:\n%s", got)
	}
}

func TestRefetching(t *testing.T) {
	t.Parallel()

	fetcher := &fakeFetcher{dir: schemaDir(t, "chain.yaml")}
	resolver := &Resolver{Fetcher: fetcher, ASCII: true}
	d := openDetail(t, registryRow("minimalist"), resolver)

	if fetcher.calls != 1 {
		t.Fatalf("the first open fetched %d times", fetcher.calls)
	}

	_, cmd := d.Update(keyPress("R"))
	if cmd == nil {
		t.Fatal("R produced no command")
	}
	if got := detailView(d); !strings.Contains(got, "fetching") {
		t.Errorf("the refetch is not reported:\n%s", got)
	}

	_, _ = d.Update(cmd())

	if fetcher.calls != 2 {
		t.Errorf("the refetch did not run: %d calls", fetcher.calls)
	}
	if !fetcher.refetch {
		t.Error("R did not ask for a refetch")
	}
}

func TestShowingWhereTheSchemaComesFrom(t *testing.T) {
	t.Parallel()

	fetcher := &fakeFetcher{dir: schemaDir(t, "chain.yaml")}
	d := openDetail(t, registryRow("minimalist"), &Resolver{Fetcher: fetcher, ASCII: true})

	_, _ = d.Update(keyPress("o"))

	if got := detailView(d); !strings.Contains(got, "https://github.com/speclib/minimalist") {
		t.Errorf("o did not show the repository:\n%s", got)
	}
}

func TestShowingWhereALocalSchemaIs(t *testing.T) {
	t.Parallel()

	dir := schemaDir(t, "chain.yaml")
	d := openDetail(t, registry.RowFromInstalled("minimalist", "package", dir, nil), &Resolver{ASCII: true})

	_, _ = d.Update(keyPress("o"))

	if got := detailView(d); !strings.Contains(got, dir) {
		t.Errorf("o did not show the path:\n%s", got)
	}
}

func TestFindingsTravelWithTheSchema(t *testing.T) {
	t.Parallel()

	t.Run("a fatal problem", func(t *testing.T) {
		t.Parallel()

		fetcher := &fakeFetcher{dir: schemaDir(t, "cycle.yaml")}
		d := openDetail(t, registryRow("cycle"), &Resolver{Fetcher: fetcher, ASCII: true})

		got := detailView(d)
		if !strings.Contains(got, "fatal problem") {
			t.Errorf("the fatal problem is not reported:\n%s", got)
		}
		if !strings.Contains(got, "cycle") {
			t.Errorf("the problem is not described:\n%s", got)
		}
		if !strings.Contains(got, "design") {
			t.Errorf("the schema was not shown despite being invalid:\n%s", got)
		}
	})

	t.Run("warnings only", func(t *testing.T) {
		t.Parallel()

		fetcher := &fakeFetcher{dir: schemaDir(t, "orphan.yaml")}
		d := openDetail(t, registryRow("orphan"), &Resolver{Fetcher: fetcher, ASCII: true})

		got := detailView(d)
		if !strings.Contains(got, "warning") {
			t.Errorf("the warning is not reported:\n%s", got)
		}
		if strings.Contains(got, "fatal problem") {
			t.Errorf("a schema with only warnings was reported as broken:\n%s", got)
		}
	})

	t.Run("nothing to report", func(t *testing.T) {
		t.Parallel()

		fetcher := &fakeFetcher{dir: schemaDir(t, "chain.yaml")}
		d := openDetail(t, registryRow("minimalist"), &Resolver{Fetcher: fetcher, ASCII: true})

		got := detailView(d)
		if strings.Contains(got, "warning") || strings.Contains(got, "fatal") {
			t.Errorf("a valid schema reported findings:\n%s", got)
		}
	})
}

func TestTheDiagramToggles(t *testing.T) {
	t.Parallel()

	fetcher := &fakeFetcher{dir: schemaDir(t, "chain.yaml")}
	d := openDetail(t, registryRow("minimalist"), &Resolver{Fetcher: fetcher, ASCII: true})

	_, _ = d.Update(keyPress("d"))

	got := detailView(d)
	if !strings.Contains(got, "d closes the diagram") {
		t.Errorf("the diagram did not open:\n%s", got)
	}
	if !strings.Contains(got, "specs") || !strings.Contains(got, "tasks") {
		t.Errorf("the diagram does not carry the artifacts:\n%s", got)
	}

	_, _ = d.Update(keyPress("d"))

	if got := detailView(d); strings.Contains(got, "d closes the diagram") {
		t.Errorf("the diagram did not close:\n%s", got)
	}
}

func TestTheDiagramReportsACycle(t *testing.T) {
	t.Parallel()

	fetcher := &fakeFetcher{dir: schemaDir(t, "cycle.yaml")}
	d := openDetail(t, registryRow("cycle"), &Resolver{Fetcher: fetcher, ASCII: true})

	_, _ = d.Update(keyPress("d"))

	got := detailView(d)
	if !strings.Contains(got, "cannot be drawn") {
		t.Errorf("the cycle is not reported:\n%s", got)
	}
	if !strings.Contains(got, "design") {
		t.Errorf("the artifacts on the cycle are not named:\n%s", got)
	}
}

func TestTheDiagramReportsNothingToDraw(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "schema.yaml"), []byte("name: bare\nversion: 1\nartifacts: []\napply:\n  requires: []\n"), 0o644); err != nil {
		t.Fatalf("writing the schema: %v", err)
	}

	fetcher := &fakeFetcher{dir: dir}
	d := openDetail(t, registryRow("bare"), &Resolver{Fetcher: fetcher, ASCII: true})

	_, _ = d.Update(keyPress("d"))

	if got := detailView(d); !strings.Contains(got, "nothing to draw") {
		t.Errorf("an empty schema is not reported:\n%s", got)
	}
}

func TestTheDiagramScrolls(t *testing.T) {
	t.Parallel()

	fetcher := &fakeFetcher{dir: schemaDir(t, "branchy.yaml")}
	d := openDetail(t, registryRow("branchy"), &Resolver{Fetcher: fetcher, ASCII: true})

	_, _ = d.Update(keyPress("d"))

	before := ansi.Strip(d.View(16, 8))

	for range 5 {
		_, _ = d.Update(keyPress("down"))
	}
	afterDown := ansi.Strip(d.View(16, 8))

	if before == afterDown {
		t.Errorf("scrolling down changed nothing:\n%s", before)
	}

	for range 5 {
		_, _ = d.Update(keyPress("right"))
	}
	afterRight := ansi.Strip(d.View(16, 8))

	if afterDown == afterRight {
		t.Errorf("scrolling right changed nothing:\n%s", afterDown)
	}

	_, _ = d.Update(keyPress("g"))
	if got := ansi.Strip(d.View(16, 8)); got != before {
		t.Errorf("g did not return to the top left:\n%s", got)
	}
}

func TestScrollingDoesNothingWithTheDiagramClosed(t *testing.T) {
	t.Parallel()

	fetcher := &fakeFetcher{dir: schemaDir(t, "branchy.yaml")}
	d := openDetail(t, registryRow("branchy"), &Resolver{Fetcher: fetcher, ASCII: true})

	_, _ = d.Update(keyPress("right"))

	if d.diagramView.left != 0 {
		t.Errorf("the diagram scrolled while closed: left = %d", d.diagramView.left)
	}
}

func TestMovingThroughTheArtifacts(t *testing.T) {
	t.Parallel()

	fetcher := &fakeFetcher{dir: schemaDir(t, "branchy.yaml")}
	d := openDetail(t, registryRow("branchy"), &Resolver{Fetcher: fetcher, ASCII: true})

	for range 10 {
		_, _ = d.Update(keyPress("down"))
	}
	if d.selected != 3 {
		t.Errorf("moving down past the end left the selection at %d, want 3", d.selected)
	}

	for range 10 {
		_, _ = d.Update(keyPress("up"))
	}
	if d.selected != 0 {
		t.Errorf("moving up past the start left the selection at %d, want 0", d.selected)
	}
}

func TestMovingWithNoArtifacts(t *testing.T) {
	t.Parallel()

	d := &detailModel{}
	d.move(1)

	if d.selected != 0 {
		t.Errorf("moving with no artifacts left the selection at %d", d.selected)
	}
}

func TestTheDetailReportsItsKeys(t *testing.T) {
	t.Parallel()

	d := newDetail(registryRow("minimalist"), nil, true)

	var found []string
	for _, k := range d.Keys() {
		found = append(found, k.Key)
	}

	for _, want := range []string{"d", "R", "o", "esc / q"} {
		if !contains(found, want) {
			t.Errorf("the detail does not report %q among %v", want, found)
		}
	}

	if d.Capturing() {
		t.Error("the detail reports that it captures text")
	}
	if d.Title() != "minimalist" {
		t.Errorf("title = %q", d.Title())
	}
}

func TestRefetchWithNoResolverDoesNothing(t *testing.T) {
	t.Parallel()

	d := newDetail(registryRow("minimalist"), nil, true)

	if _, cmd := d.Update(keyPress("R")); cmd != nil {
		t.Error("R produced a command with no resolver")
	}
}

func TestAnUnknownMessageIsIgnoredByTheDetail(t *testing.T) {
	t.Parallel()

	type odd struct{}

	d := newDetail(registryRow("minimalist"), nil, true)

	if _, cmd := d.Update(odd{}); cmd != nil {
		t.Error("an unknown message produced a command")
	}
}

func TestResolveReportsASchemaThatWillNotLoad(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "schema.yaml"), []byte("name: [unclosed\n"), 0o644); err != nil {
		t.Fatalf("writing the schema: %v", err)
	}

	fetcher := &fakeFetcher{dir: dir}
	d := openDetail(t, registryRow("broken"), &Resolver{Fetcher: fetcher, ASCII: true})

	if got := detailView(d); !strings.Contains(got, "Could not read this schema") {
		t.Errorf("a schema that will not parse is not reported:\n%s", got)
	}
}

func TestResolvedReportsACyclicGraph(t *testing.T) {
	t.Parallel()

	fetcher := &fakeFetcher{dir: schemaDir(t, "cycle.yaml")}

	resolved, err := source.Resolve(context.Background(), fetcher, source.Source{}, "cycle", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !resolved.Cyclic() {
		t.Error("a cyclic schema does not report itself as cyclic")
	}
	if resolved.Metrics.Artifacts != 3 {
		t.Errorf("a cyclic schema reported %d artifacts, want 3", resolved.Metrics.Artifacts)
	}
}

func TestTheDetailScrollsItsArtifactTable(t *testing.T) {
	t.Parallel()

	fetcher := &fakeFetcher{dir: schemaDir(t, "spec-driven.yaml")}
	d := openDetail(t, registryRow("spec-driven"), &Resolver{Fetcher: fetcher, ASCII: true})

	before := ansi.Strip(d.View(80, 9))

	for range 3 {
		_, _ = d.Update(keyPress("down"))
	}

	after := ansi.Strip(d.View(80, 9))
	if before == after {
		t.Errorf("the artifact table did not scroll:\n%s", before)
	}
	if !strings.Contains(after, "tasks") {
		t.Errorf("the selected artifact is not visible:\n%s", after)
	}
}

func TestTheDetailShowsTheSelectedTemplate(t *testing.T) {
	t.Parallel()

	fetcher := &fakeFetcher{dir: schemaDir(t, "chain.yaml")}
	d := openDetail(t, registryRow("minimalist"), &Resolver{Fetcher: fetcher, ASCII: true})

	if got := detailView(d); !strings.Contains(got, "template: specs/spec.md") {
		t.Errorf("the selected artifact's template is not shown:\n%s", got)
	}

	_, _ = d.Update(keyPress("down"))

	if got := detailView(d); !strings.Contains(got, "template: tasks.md") {
		t.Errorf("the template did not follow the selection:\n%s", got)
	}
}

func TestADetailWithNoArtifactsSaysSo(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "schema.yaml"), []byte("name: bare\nversion: 1\nartifacts: []\napply:\n  requires: []\n"), 0o644); err != nil {
		t.Fatalf("writing the schema: %v", err)
	}

	fetcher := &fakeFetcher{dir: dir}
	d := openDetail(t, registryRow("bare"), &Resolver{Fetcher: fetcher, ASCII: true})

	got := detailView(d)
	if !strings.Contains(got, "declares no artifacts") {
		t.Errorf("an empty schema is not reported:\n%s", got)
	}
	if !strings.Contains(got, "apply gate: nothing") {
		t.Errorf("an absent gate is not reported:\n%s", got)
	}
	if !strings.Contains(got, "tracks: nothing") {
		t.Errorf("an absent tracked file is not reported:\n%s", got)
	}
}

func TestADetailWithNothingLoadedSaysSo(t *testing.T) {
	t.Parallel()

	d := newDetail(registryRow("minimalist"), nil, true)
	d.busy = false

	if got := flat(ansi.Strip(d.View(80, 20))); !strings.Contains(got, "Nothing loaded") {
		t.Errorf("a detail with nothing loaded draws:\n%s", got)
	}
}

func TestTheDiagramHintMentionsScrollingOnlyWhenItIsNeeded(t *testing.T) {
	t.Parallel()

	fetcher := &fakeFetcher{dir: schemaDir(t, "chain.yaml")}
	d := openDetail(t, registryRow("minimalist"), &Resolver{Fetcher: fetcher, ASCII: true})

	_, _ = d.Update(keyPress("d"))

	if got := detailView(d); strings.Contains(got, "arrows scroll") {
		t.Errorf("a diagram that fits offers scrolling:\n%s", got)
	}

	if got := flat(ansi.Strip(d.View(12, 5))); !strings.Contains(got, "arrows scroll") {
		t.Errorf("a diagram that does not fit does not offer scrolling:\n%s", got)
	}
}

func TestACycleWithNoFindingToQuote(t *testing.T) {
	t.Parallel()

	d := &detailModel{}

	if got := d.cycleNames(); got != "" {
		t.Errorf("cycleNames with no findings = %q", got)
	}
}

func TestWhereFallsBackToTheRowPath(t *testing.T) {
	t.Parallel()

	d := newDetail(registry.RowFromInstalled("x", "package", "/some/path", nil), nil, true)

	if got := d.where(); got != "/some/path" {
		t.Errorf("where = %q, want the row's path", got)
	}
}

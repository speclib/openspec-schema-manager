package tui

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"time"

	"github.com/speclib/openspec-schema-manager/internal/ansi"
	"github.com/speclib/openspec-schema-manager/internal/openspec"
	"github.com/speclib/openspec-schema-manager/internal/registry"
	"github.com/speclib/openspec-schema-manager/internal/source"
)

func demoProject(t *testing.T, config string) string {
	t.Helper()

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "openspec", "changes"), 0o755); err != nil {
		t.Fatalf("creating the project: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "openspec", "config.yaml"), []byte(config), 0o644); err != nil {
		t.Fatalf("writing the config: %v", err)
	}

	return root
}

func fullProject() *openspec.Fake {
	return &openspec.Fake{
		SchemasResult: []openspec.ResolvedSchema{
			{Name: "minimalist", Source: openspec.SourceProject, Path: "/p/openspec/schemas/minimalist"},
			{Name: "research-first", Source: openspec.SourceUser, Path: "/h/.local/share/openspec/schemas/research-first", Shadows: []string{"package"}},
			{Name: "spec-driven", Source: openspec.SourcePackage, Path: "/nix/store/x/spec-driven"},
		},
		ChangesResult: []openspec.Change{
			{Name: "add-auth", CompletedTasks: 3, TotalTasks: 7},
			{Name: "fix-export", CompletedTasks: 0, TotalTasks: 4},
		},
		ChangeSchemas: map[string]string{"add-auth": "spec-driven", "fix-export": "minimalist"},
	}
}

func openProject(t *testing.T, root string, cli openspec.CLI) *projectModel {
	t.Helper()

	reader := &ProjectReader{CLI: cli, Root: root}
	p := newProjectScreen(true, root, reader, &Resolver{ASCII: true}, nil)

	_, _ = p.Update(ProjectRead{Project: reader.Read(context.Background())})

	return p
}

func projectView(p *projectModel) string { return flat(ansi.Strip(p.View(120, 30))) }

func TestTheProjectTabShowsWhatTheProjectHas(t *testing.T) {
	t.Parallel()

	root := demoProject(t, "schema: minimalist\n")
	p := openProject(t, root, fullProject())

	got := projectView(p)

	for _, want := range []string{
		root,
		"default schema: minimalist",
		"Schemas available",
		"minimalist",
		"project",
		"(default)",
		"research-first",
		"user",
		"shadows package",
		"spec-driven",
		"built-in",
		"Changes",
		"add-auth",
		"3/7 tasks",
		"fix-export",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("the project view does not carry %q:\n%s", want, got)
		}
	}
}

func TestAProjectWithNoChangesSaysSo(t *testing.T) {
	t.Parallel()

	p := openProject(t, demoProject(t, "schema: spec-driven\n"), &openspec.Fake{
		SchemasResult: []openspec.ResolvedSchema{{Name: "spec-driven", Source: openspec.SourcePackage}},
	})

	if got := projectView(p); !strings.Contains(got, "no changes yet") {
		t.Errorf("a project with no changes does not say so:\n%s", got)
	}
}

func TestAProjectWithNoSchemasSaysSo(t *testing.T) {
	t.Parallel()

	p := openProject(t, demoProject(t, ""), &openspec.Fake{})

	got := projectView(p)
	if !strings.Contains(got, "No schemas resolve") {
		t.Errorf("a project with no schemas does not say so:\n%s", got)
	}
	if !strings.Contains(got, "default schema unknown") {
		t.Errorf("an unknown default is not reported:\n%s", got)
	}
}

func TestOutsideAProjectTheTabSaysWhatWouldMakeItWork(t *testing.T) {
	t.Parallel()

	p := newProjectScreen(false, "", nil, nil, nil)

	got := flat(ansi.Strip(p.View(120, 30)))
	for _, want := range []string{"no OpenSpec project here", "Browsing", "installing one needs a project", "--path"} {
		if !strings.Contains(got, want) {
			t.Errorf("the view does not carry %q:\n%s", want, got)
		}
	}
}

func TestBeforeItIsReadTheTabSaysSo(t *testing.T) {
	t.Parallel()

	p := newProjectScreen(true, "/p", nil, nil, nil)

	if got := flat(ansi.Strip(p.View(120, 30))); !strings.Contains(got, "Reading the project") {
		t.Errorf("the view does not say it is reading:\n%s", got)
	}
}

func TestAProblemIsReportedWithoutHidingTheRest(t *testing.T) {
	t.Parallel()

	fake := fullProject()
	fake.ChangesErr = openspec.ErrNotInstalled

	p := openProject(t, demoProject(t, "schema: minimalist\n"), fake)

	got := projectView(p)
	if !strings.Contains(got, "not on PATH") {
		t.Errorf("the problem is not reported:\n%s", got)
	}
	if !strings.Contains(got, "minimalist") {
		t.Errorf("the schemas were hidden by a problem listing changes:\n%s", got)
	}
}

func TestRefreshingTheProject(t *testing.T) {
	t.Parallel()

	root := demoProject(t, "schema: minimalist\n")
	reader := &ProjectReader{CLI: fullProject(), Root: root}
	p := newProjectScreen(true, root, reader, nil, nil)

	_, cmd := p.Update(keyPress("r"))
	if cmd == nil {
		t.Fatal("r produced no command")
	}
	if got := projectView(p); !strings.Contains(got, "reading the project") {
		t.Errorf("the refresh is not reported:\n%s", got)
	}

	_, _ = p.Update(cmd())

	if got := projectView(p); !strings.Contains(got, "add-auth") {
		t.Errorf("the refresh did not take:\n%s", got)
	}
}

func TestRefreshingWithNoReaderDoesNothing(t *testing.T) {
	t.Parallel()

	p := newProjectScreen(true, "/p", nil, nil, nil)

	if _, cmd := p.Update(keyPress("r")); cmd != nil {
		t.Error("r produced a command with no reader")
	}
}

func TestMovingThroughTheProjectSchemas(t *testing.T) {
	t.Parallel()

	p := openProject(t, demoProject(t, "schema: minimalist\n"), fullProject())

	for range 10 {
		_, _ = p.Update(keyPress("down"))
	}
	if p.selected != 2 {
		t.Errorf("moving down past the end left the selection at %d, want 2", p.selected)
	}

	for range 10 {
		_, _ = p.Update(keyPress("up"))
	}
	if p.selected != 0 {
		t.Errorf("moving up past the start left the selection at %d", p.selected)
	}

	_, _ = p.Update(keyPress("G"))
	if p.selected != 2 {
		t.Errorf("G left the selection at %d", p.selected)
	}

	_, _ = p.Update(keyPress("g"))
	if p.selected != 0 {
		t.Errorf("g left the selection at %d", p.selected)
	}
}

func TestOpeningAProjectSchema(t *testing.T) {
	t.Parallel()

	dir := schemaDir(t, "chain.yaml")

	fake := &openspec.Fake{
		SchemasResult: []openspec.ResolvedSchema{{Name: "minimalist", Source: openspec.SourceProject, Path: dir}},
	}

	fetcher := &fakeFetcher{dir: dir}
	root := demoProject(t, "schema: minimalist\n")

	reader := &ProjectReader{CLI: fake, Root: root}
	p := newProjectScreen(true, root, reader, &Resolver{Fetcher: fetcher, ASCII: true}, nil)
	_, _ = p.Update(ProjectRead{Project: reader.Read(context.Background())})

	_, cmd := p.Update(keyPress("enter"))
	if cmd == nil {
		t.Fatal("enter produced no command")
	}

	_, _ = p.Update(cmd())

	if fetcher.calls != 0 {
		t.Errorf("a project schema was fetched %d times", fetcher.calls)
	}
	if got := projectView(p); !strings.Contains(got, "apply gate") {
		t.Errorf("the detail did not open:\n%s", got)
	}

	_, _ = p.Update(keyPress("esc"))
	if got := projectView(p); !strings.Contains(got, "Schemas available") {
		t.Errorf("esc did not return to the project view:\n%s", got)
	}
}

func TestOpeningWithNoSelectionDoesNothing(t *testing.T) {
	t.Parallel()

	p := newProjectScreen(true, "/p", nil, &Resolver{ASCII: true}, nil)

	if _, cmd := p.Update(keyPress("enter")); cmd != nil {
		t.Error("enter with nothing selected produced a command")
	}
}

func TestSettingTheProjectDefault(t *testing.T) {
	t.Parallel()

	root := demoProject(t, "schema: spec-driven\n\n# a comment\n")
	p := openProject(t, root, fullProject())

	if p.project.Default != "spec-driven" {
		t.Fatalf("the default starts at %q", p.project.Default)
	}

	_, _ = p.Update(keyPress("s"))

	if got := openspec.DefaultSchema(root); got != "minimalist" {
		t.Errorf("the default is %q, want minimalist", got)
	}
	if got := projectView(p); !strings.Contains(got, "minimalist is now the project default") {
		t.Errorf("setting the default is not reported:\n%s", got)
	}

	raw, err := os.ReadFile(filepath.Join(root, "openspec", "config.yaml"))
	if err != nil {
		t.Fatalf("reading the config: %v", err)
	}
	if !strings.Contains(string(raw), "# a comment") {
		t.Errorf("the comment was lost:\n%s", raw)
	}
}

func TestSettingTheDefaultOutsideAProjectDoesNothing(t *testing.T) {
	t.Parallel()

	p := newProjectScreen(false, "", nil, nil, nil)

	if _, cmd := p.Update(keyPress("s")); cmd != nil {
		t.Error("s produced a command outside a project")
	}
}

func TestSettingTheDefaultReportsAFailure(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	reader := &ProjectReader{CLI: fullProject(), Root: root}
	p := newProjectScreen(true, root, reader, nil, nil)
	_, _ = p.Update(ProjectRead{Project: openspec.Project{
		Root:    root,
		Schemas: []openspec.ResolvedSchema{{Name: "minimalist", Source: openspec.SourceProject}},
	}})

	_, _ = p.Update(keyPress("s"))

	if got := projectView(p); !strings.Contains(got, "could not set the default") {
		t.Errorf("a failure to set the default is not reported:\n%s", got)
	}
}

func TestTheProjectRefreshesAfterAnInstall(t *testing.T) {
	t.Parallel()

	root := demoProject(t, "schema: spec-driven\n")
	reader := &ProjectReader{CLI: fullProject(), Root: root}
	p := newProjectScreen(true, root, reader, nil, nil)

	_, cmd := p.Update(InstallFinished{Name: "minimalist"})
	if cmd == nil {
		t.Fatal("an install did not trigger a re-read")
	}

	_, _ = p.Update(cmd())

	if got := projectView(p); !strings.Contains(got, "add-auth") {
		t.Errorf("the project was not re-read:\n%s", got)
	}
}

func TestAFailedInstallDoesNotTriggerARead(t *testing.T) {
	t.Parallel()

	root := demoProject(t, "schema: spec-driven\n")
	p := newProjectScreen(true, root, &ProjectReader{CLI: fullProject(), Root: root}, nil, nil)

	if _, cmd := p.Update(InstallFinished{Err: errors.New("no")}); cmd != nil {
		t.Error("a failed install triggered a re-read")
	}
}

func TestTheProjectReportsItsKeys(t *testing.T) {
	t.Parallel()

	p := newProjectScreen(true, "/p", nil, nil, nil)

	var found []string
	for _, k := range p.Keys() {
		found = append(found, k.Key)
	}

	for _, want := range []string{"enter", "s", "r"} {
		if !contains(found, want) {
			t.Errorf("the project screen does not report %q among %v", want, found)
		}
	}
	if p.Capturing() {
		t.Error("the project screen reports that it captures text")
	}
	if p.Title() != "Project" {
		t.Errorf("title = %q", p.Title())
	}
}

func TestTheProjectSchemaListScrolls(t *testing.T) {
	t.Parallel()

	var schemas []openspec.ResolvedSchema
	for i := range 20 {
		schemas = append(schemas, openspec.ResolvedSchema{
			Name:   "schema-" + string(rune('a'+i)),
			Source: openspec.SourceProject,
		})
	}

	root := demoProject(t, "schema: spec-driven\n")
	p := openProject(t, root, &openspec.Fake{SchemasResult: schemas})

	before := ansi.Strip(p.View(80, 12))

	for range 15 {
		_, _ = p.Update(keyPress("down"))
	}

	if after := ansi.Strip(p.View(80, 12)); after == before {
		t.Errorf("the schema list did not scroll:\n%s", before)
	}
}

func TestAnUnknownMessageIsIgnoredByTheProject(t *testing.T) {
	t.Parallel()

	type odd struct{}

	p := newProjectScreen(true, "/p", nil, nil, nil)

	if _, cmd := p.Update(odd{}); cmd != nil {
		t.Error("an unknown message produced a command")
	}
}

func TestTheProjectDelegatesToTheDetail(t *testing.T) {
	t.Parallel()

	dir := schemaDir(t, "chain.yaml")
	root := demoProject(t, "schema: minimalist\n")

	fake := &openspec.Fake{
		SchemasResult: []openspec.ResolvedSchema{{Name: "minimalist", Source: openspec.SourceProject, Path: dir}},
	}

	reader := &ProjectReader{CLI: fake, Root: root}
	p := newProjectScreen(true, root, reader, &Resolver{ASCII: true}, nil)
	_, _ = p.Update(ProjectRead{Project: reader.Read(context.Background())})

	_, cmd := p.Update(keyPress("enter"))
	_, _ = p.Update(cmd())

	var keys []string
	for _, k := range p.Keys() {
		keys = append(keys, k.Key)
	}
	if !contains(keys, "d") {
		t.Errorf("the help does not follow into the detail: %v", keys)
	}
	if p.Capturing() {
		t.Error("the detail reports that it captures text")
	}

	_, _ = p.Update(keyPress("d"))
	if got := projectView(p); !strings.Contains(got, "d closes the diagram") {
		t.Errorf("the detail did not take the d key:\n%s", got)
	}
}

func TestTheProjectDefaultMarkerFollowsTheSetting(t *testing.T) {
	t.Parallel()

	root := demoProject(t, "schema: spec-driven\n")
	p := openProject(t, root, fullProject())

	got := projectView(p)
	if !strings.Contains(got, "spec-driven built-in (default)") {
		t.Errorf("the default marker is not on spec-driven:\n%s", got)
	}
}

func TestComparingAnInstalledSchema(t *testing.T) {
	t.Parallel()

	files := map[string]string{"schema.yaml": "name: minimalist\n"}

	installed := t.TempDir()
	for rel, body := range files {
		writeFile(t, filepath.Join(installed, rel), body)
	}

	fetched := t.TempDir()
	for rel, body := range files {
		writeFile(t, filepath.Join(fetched, rel), body)
	}

	entries := []registry.Entry{{
		ID:          "speclib/minimalist",
		Name:        "minimalist",
		Description: "a schema",
		Artifacts:   []string{"tasks"},
		Source:      registry.Source{Repo: "https://example.test/r", Path: "p", Ref: "v1"},
	}}

	comparer := &Comparer{
		Fetcher: &fakeFetcher{dir: fetched},
		Entries: func() []registry.Entry { return entries },
	}

	root := demoProject(t, "schema: minimalist\n")
	fake := &openspec.Fake{
		SchemasResult: []openspec.ResolvedSchema{{Name: "minimalist", Source: openspec.SourceProject, Path: installed}},
	}

	reader := &ProjectReader{CLI: fake, Root: root}
	p := newProjectScreen(true, root, reader, nil, comparer)
	_, _ = p.Update(ProjectRead{Project: reader.Read(context.Background())})

	_, cmd := p.Update(keyPress("u"))
	if cmd == nil {
		t.Fatal("u produced no command")
	}
	if got := projectView(p); !strings.Contains(got, "comparing minimalist") {
		t.Errorf("the comparison is not reported while it runs:\n%s", got)
	}

	_, _ = p.Update(cmd())

	if got := projectView(p); !strings.Contains(got, "matches what the registry offers at v1") {
		t.Errorf("the result is not shown:\n%s", got)
	}
}

func TestADifferenceIsNeverCalledAnUpdate(t *testing.T) {
	t.Parallel()

	installed := t.TempDir()
	writeFile(t, filepath.Join(installed, "schema.yaml"), "name: minimalist\n# edited here\n")

	fetched := t.TempDir()
	writeFile(t, filepath.Join(fetched, "schema.yaml"), "name: minimalist\n")

	entries := []registry.Entry{{
		ID:          "speclib/minimalist",
		Name:        "minimalist",
		Description: "a schema",
		Artifacts:   []string{"tasks"},
		Source:      registry.Source{Repo: "https://example.test/r", Path: "p"},
	}}

	comparer := &Comparer{
		Fetcher: &fakeFetcher{dir: fetched},
		Entries: func() []registry.Entry { return entries },
	}

	root := demoProject(t, "schema: minimalist\n")
	fake := &openspec.Fake{
		SchemasResult: []openspec.ResolvedSchema{{Name: "minimalist", Source: openspec.SourceProject, Path: installed}},
	}

	reader := &ProjectReader{CLI: fake, Root: root}
	p := newProjectScreen(true, root, reader, nil, comparer)
	_, _ = p.Update(ProjectRead{Project: reader.Read(context.Background())})

	_, cmd := p.Update(keyPress("u"))
	_, _ = p.Update(cmd())

	if got := projectView(p); !strings.Contains(got, "differs from the registry") {
		t.Errorf("the difference is not reported:\n%s", got)
	}

	// Checked against the message itself rather than the whole screen: the
	// temporary directory in the header carries this test's own name.
	if strings.Contains(strings.ToLower(p.comparison), "update") {
		t.Errorf("a difference was described as an update:\n%s", p.comparison)
	}
	if !strings.Contains(p.comparison, "cannot tell an upstream change from a local edit") {
		t.Errorf("the limitation is not stated:\n%s", p.comparison)
	}
}

func TestDrawingTheProjectFetchesNothing(t *testing.T) {
	t.Parallel()

	fetcher := &fakeFetcher{dir: t.TempDir()}
	comparer := &Comparer{Fetcher: fetcher, Entries: func() []registry.Entry { return nil }}

	root := demoProject(t, "schema: minimalist\n")
	reader := &ProjectReader{CLI: fullProject(), Root: root}

	p := newProjectScreen(true, root, reader, nil, comparer)
	_, _ = p.Update(ProjectRead{Project: reader.Read(context.Background())})

	_ = projectView(p)

	_, cmd := p.Update(keyPress("r"))
	if cmd != nil {
		_, _ = p.Update(cmd())
	}
	_ = projectView(p)

	if fetcher.calls != 0 {
		t.Errorf("drawing or refreshing the project fetched %d times", fetcher.calls)
	}
}

func TestComparingWithNoComparerDoesNothing(t *testing.T) {
	t.Parallel()

	root := demoProject(t, "schema: minimalist\n")
	p := openProject(t, root, fullProject())

	if _, cmd := p.Update(keyPress("u")); cmd != nil {
		t.Error("u produced a command with no comparer")
	}
}

func TestComparingWithNothingSelected(t *testing.T) {
	t.Parallel()

	comparer := &Comparer{Fetcher: &fakeFetcher{}, Entries: func() []registry.Entry { return nil }}
	p := newProjectScreen(true, "/p", nil, nil, comparer)

	if _, cmd := p.Update(keyPress("u")); cmd != nil {
		t.Error("u produced a command with nothing selected")
	}
}

func TestTheComparerWorksWithNoEntrySource(t *testing.T) {
	t.Parallel()

	comparer := &Comparer{Fetcher: &fakeFetcher{}}

	msg := comparer.CompareCmd(openspec.ResolvedSchema{Name: "x", Source: openspec.SourceProject, Path: t.TempDir()})()

	compared, ok := msg.(SchemaCompared)
	if !ok {
		t.Fatalf("CompareCmd returned %T", msg)
	}
	if compared.Comparison.Verdict != source.VerdictNoEntry {
		t.Errorf("verdict = %v, want no entry", compared.Comparison.Verdict)
	}
}

func TestTheRegistryScreenHandsOutItsEntries(t *testing.T) {
	t.Parallel()

	l := newLoader(t, fixtureRegistry, nil)
	seed(t, l, fixtureRegistry, time.Hour)

	s := newRegistryScreen(l, nil, nil)
	_, _ = s.Update(l.Load(context.Background()))

	entries := s.(*registryScreenModel).Entries()
	if len(entries) != 4 {
		t.Fatalf("got %d entries, want the four in the fixture", len(entries))
	}
	for _, e := range entries {
		if e.ID == "" {
			t.Errorf("an entry came back empty: %+v", e)
		}
	}
}

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
)

func installerFor(t *testing.T, root, sourceDir string, cli openspec.CLI) *Installer {
	t.Helper()

	return &Installer{
		Fetcher:   &fakeFetcher{dir: sourceDir},
		CLI:       cli,
		Root:      root,
		InProject: true,
	}
}

func runInstall(t *testing.T, f *installFlow, row registry.Row, keys ...string) {
	t.Helper()

	cmd, refusal := f.start(row)
	if refusal != "" {
		t.Fatalf("the install was refused: %s", refusal)
	}
	if cmd == nil {
		t.Fatal("start produced no command")
	}

	_ = f.Update(cmd())

	for _, key := range keys {
		if next := f.Update(keyPress(key)); next != nil {
			_ = f.Update(next())
		}
	}
}

func flowView(f *installFlow) string { return flat(ansi.Strip(f.View(100))) }

func TestInstallShowsWhatWouldBeWrittenBeforeWritingIt(t *testing.T) {
	t.Parallel()

	root := demoProject(t, "schema: spec-driven\n")
	f := newInstallFlow(installerFor(t, root, schemaDir(t, "chain.yaml"), &openspec.Fake{}))

	cmd, refusal := f.start(registryRow("minimalist"))
	if refusal != "" {
		t.Fatalf("refused: %s", refusal)
	}
	_ = f.Update(cmd())

	got := flowView(f)
	for _, want := range []string{"Install minimalist", "openspec/schemas/minimalist", "schema.yaml", "templates/tasks.md", "y write these files", "n cancel"} {
		if !strings.Contains(got, want) {
			t.Errorf("the preview does not carry %q:\n%s", want, got)
		}
	}

	if _, err := os.Stat(openspec.Destination(root, "minimalist")); err == nil {
		t.Error("something was written before the confirmation")
	}
}

func TestDecliningWritesNothing(t *testing.T) {
	t.Parallel()

	root := demoProject(t, "schema: spec-driven\n")
	f := newInstallFlow(installerFor(t, root, schemaDir(t, "chain.yaml"), &openspec.Fake{}))

	runInstall(t, f, registryRow("minimalist"), "n")

	if _, err := os.Stat(openspec.Destination(root, "minimalist")); err == nil {
		t.Error("declining wrote the schema anyway")
	}
	if f.active() {
		t.Error("declining left the flow active")
	}
}

func TestConfirmingWritesTheSchema(t *testing.T) {
	t.Parallel()

	root := demoProject(t, "schema: spec-driven\n")
	f := newInstallFlow(installerFor(t, root, schemaDir(t, "chain.yaml"), &openspec.Fake{}))

	runInstall(t, f, registryRow("minimalist"), "y")

	destination := openspec.Destination(root, "minimalist")
	if _, err := os.Stat(filepath.Join(destination, "schema.yaml")); err != nil {
		t.Fatalf("the schema was not written: %v", err)
	}

	got := flowView(f)
	if !strings.Contains(got, "minimalist installed") {
		t.Errorf("the install is not reported as done:\n%s", got)
	}
	if !strings.Contains(got, "OpenSpec validated it") {
		t.Errorf("the validation result is not shown:\n%s", got)
	}
	if !strings.Contains(got, "project default is unchanged") {
		t.Errorf("the view does not say the default is unchanged:\n%s", got)
	}
}

func TestInstallingDoesNotTouchTheProjectDefault(t *testing.T) {
	t.Parallel()

	config := "schema: spec-driven\n\n# a comment\n"
	root := demoProject(t, config)
	f := newInstallFlow(installerFor(t, root, schemaDir(t, "chain.yaml"), &openspec.Fake{}))

	runInstall(t, f, registryRow("minimalist"), "y")

	raw, err := os.ReadFile(filepath.Join(root, "openspec", "config.yaml"))
	if err != nil {
		t.Fatalf("reading the config: %v", err)
	}
	if string(raw) != config {
		t.Errorf("an install changed the config:\n%s", raw)
	}
}

func TestAnOccupiedDestinationNeedsItsOwnConfirmation(t *testing.T) {
	t.Parallel()

	root := demoProject(t, "schema: spec-driven\n")
	src := schemaDir(t, "chain.yaml")
	installer := installerFor(t, root, src, &openspec.Fake{})

	first := newInstallFlow(installer)
	runInstall(t, first, registryRow("minimalist"), "y")

	marker := filepath.Join(openspec.Destination(root, "minimalist"), "leftover.md")
	if err := os.WriteFile(marker, []byte("old\n"), 0o644); err != nil {
		t.Fatalf("writing the marker: %v", err)
	}

	second := newInstallFlow(installer)

	cmd, _ := second.start(registryRow("minimalist"))
	_ = second.Update(cmd())

	got := flowView(second)
	for _, want := range []string{"Overwrite minimalist?", "already holds a schema", "records no provenance", "o overwrite it", "n leave it alone"} {
		if !strings.Contains(got, want) {
			t.Errorf("the overwrite confirmation does not carry %q:\n%s", want, got)
		}
	}

	if next := second.Update(keyPress("y")); next != nil {
		t.Error("y overwrote without the explicit overwrite key")
	}
	if _, err := os.Stat(marker); err != nil {
		t.Error("y removed what was already there")
	}
}

func TestOverwritingReplacesWhatWasThere(t *testing.T) {
	t.Parallel()

	root := demoProject(t, "schema: spec-driven\n")
	installer := installerFor(t, root, schemaDir(t, "chain.yaml"), &openspec.Fake{})

	first := newInstallFlow(installer)
	runInstall(t, first, registryRow("minimalist"), "y")

	marker := filepath.Join(openspec.Destination(root, "minimalist"), "leftover.md")
	if err := os.WriteFile(marker, []byte("old\n"), 0o644); err != nil {
		t.Fatalf("writing the marker: %v", err)
	}

	second := newInstallFlow(installer)
	runInstall(t, second, registryRow("minimalist"), "o")

	if _, err := os.Stat(marker); err == nil {
		t.Error("overwriting kept the old directory")
	}
	if got := flowView(second); !strings.Contains(got, "minimalist installed") {
		t.Errorf("the overwrite is not reported as done:\n%s", got)
	}
}

func TestDecliningAnOverwriteLeavesItAlone(t *testing.T) {
	t.Parallel()

	root := demoProject(t, "schema: spec-driven\n")
	installer := installerFor(t, root, schemaDir(t, "chain.yaml"), &openspec.Fake{})

	first := newInstallFlow(installer)
	runInstall(t, first, registryRow("minimalist"), "y")

	marker := filepath.Join(openspec.Destination(root, "minimalist"), "leftover.md")
	if err := os.WriteFile(marker, []byte("old\n"), 0o644); err != nil {
		t.Fatalf("writing the marker: %v", err)
	}

	second := newInstallFlow(installer)
	runInstall(t, second, registryRow("minimalist"), "n")

	if _, err := os.Stat(marker); err != nil {
		t.Error("declining the overwrite removed what was there")
	}
	if second.active() {
		t.Error("declining left the flow active")
	}
}

func TestASchemaOpenSpecRejectsIsReportedAsSuch(t *testing.T) {
	t.Parallel()

	root := demoProject(t, "schema: spec-driven\n")
	cli := &openspec.Fake{Validations: map[string]openspec.Validation{
		"minimalist": {Name: "minimalist", Valid: false, Issues: []openspec.ValidationIssue{{Message: "template specs/spec.md is missing"}}},
	}}

	f := newInstallFlow(installerFor(t, root, schemaDir(t, "chain.yaml"), cli))
	runInstall(t, f, registryRow("minimalist"), "y")

	got := flowView(f)
	if !strings.Contains(got, "minimalist installed") {
		t.Errorf("the install is not reported as done:\n%s", got)
	}
	if !strings.Contains(got, "OpenSpec rejected it") {
		t.Errorf("the rejection is not reported:\n%s", got)
	}
	if !strings.Contains(got, "template specs/spec.md is missing") {
		t.Errorf("the issue is not shown:\n%s", got)
	}
}

func TestValidationThatCannotRunIsReportedAsUnverified(t *testing.T) {
	t.Parallel()

	root := demoProject(t, "schema: spec-driven\n")
	f := newInstallFlow(installerFor(t, root, schemaDir(t, "chain.yaml"), nil))

	runInstall(t, f, registryRow("minimalist"), "y")

	got := flowView(f)
	if !strings.Contains(got, "Installed but unverified") {
		t.Errorf("an unverified install is not reported:\n%s", got)
	}
}

func TestAFailedFetchIsReportedAndWritesNothing(t *testing.T) {
	t.Parallel()

	root := demoProject(t, "schema: spec-driven\n")

	installer := installerFor(t, root, "", &openspec.Fake{})
	installer.Fetcher = &fakeFetcher{err: errors.New("no route to host")}

	f := newInstallFlow(installer)
	runInstall(t, f, registryRow("minimalist"))

	got := flowView(f)
	if !strings.Contains(got, "Install failed") {
		t.Errorf("the failure is not reported:\n%s", got)
	}
	if !strings.Contains(got, "no route to host") {
		t.Errorf("the reason is not shown:\n%s", got)
	}
	if !strings.Contains(got, "Nothing was written") {
		t.Errorf("the view does not say nothing was written:\n%s", got)
	}

	if _, err := os.Stat(filepath.Join(root, "openspec", "schemas")); err == nil {
		t.Error("a failed fetch created the schemas directory")
	}
}

func TestInstallingIsRefusedOutsideAProject(t *testing.T) {
	t.Parallel()

	installer := installerFor(t, t.TempDir(), schemaDir(t, "chain.yaml"), &openspec.Fake{})
	installer.InProject = false

	cmd, refusal := newInstallFlow(installer).start(registryRow("minimalist"))

	if cmd != nil {
		t.Error("an install outside a project produced a command")
	}
	if !strings.Contains(refusal, "needs an OpenSpec project") {
		t.Errorf("refusal = %q", refusal)
	}
}

func TestInstallingIsRefusedWithNoInstaller(t *testing.T) {
	t.Parallel()

	cmd, refusal := newInstallFlow(nil).start(registryRow("minimalist"))

	if cmd != nil {
		t.Error("an install with no installer produced a command")
	}
	if refusal == "" {
		t.Error("no reason was given")
	}
}

func TestInstallingABuiltInSchemaIsRefused(t *testing.T) {
	t.Parallel()

	root := demoProject(t, "schema: spec-driven\n")
	f := newInstallFlow(installerFor(t, root, schemaDir(t, "chain.yaml"), &openspec.Fake{}))

	cmd, refusal := f.start(registry.RowFromInstalled("spec-driven", "package", "/nix/store/x", nil))

	if cmd != nil {
		t.Error("installing a built-in schema produced a command")
	}
	if !strings.Contains(refusal, "ships with OpenSpec") {
		t.Errorf("refusal = %q", refusal)
	}
}

func TestInstallingASchemaTheProjectAlreadyResolvesIsRefused(t *testing.T) {
	t.Parallel()

	root := demoProject(t, "schema: spec-driven\n")
	f := newInstallFlow(installerFor(t, root, schemaDir(t, "chain.yaml"), &openspec.Fake{}))

	cmd, refusal := f.start(registry.RowFromInstalled("minimalist", "project", "/p/openspec/schemas/minimalist", nil))

	if cmd != nil {
		t.Error("installing a schema the project already resolves produced a command")
	}
	if !strings.Contains(refusal, "already resolves") {
		t.Errorf("refusal = %q", refusal)
	}
}

func TestARowWithNoSourceIsReported(t *testing.T) {
	t.Parallel()

	root := demoProject(t, "schema: spec-driven\n")
	installer := installerFor(t, root, schemaDir(t, "chain.yaml"), &openspec.Fake{})

	row := registryRow("minimalist")
	row.Entry = nil

	msg := installer.PrepareCmd(row)()

	prepared, ok := msg.(InstallPrepared)
	if !ok {
		t.Fatalf("PrepareCmd returned %T", msg)
	}
	if prepared.Err == nil {
		t.Fatal("a row with no source was prepared anyway")
	}
}

func TestClosingAFinishedInstall(t *testing.T) {
	t.Parallel()

	root := demoProject(t, "schema: spec-driven\n")
	f := newInstallFlow(installerFor(t, root, schemaDir(t, "chain.yaml"), &openspec.Fake{}))

	runInstall(t, f, registryRow("minimalist"), "y")

	if !f.active() {
		t.Fatal("a finished install is not active")
	}

	_ = f.Update(keyPress("esc"))

	if f.active() {
		t.Error("esc did not close the finished install")
	}
}

func TestTheInstallFlowSwallowsOtherKeys(t *testing.T) {
	t.Parallel()

	root := demoProject(t, "schema: spec-driven\n")
	f := newInstallFlow(installerFor(t, root, schemaDir(t, "chain.yaml"), &openspec.Fake{}))

	cmd, _ := f.start(registryRow("minimalist"))
	_ = f.Update(cmd())

	if next := f.Update(keyPress("z")); next != nil {
		t.Error("an unrelated key produced a command")
	}
	if !f.active() {
		t.Error("an unrelated key closed the flow")
	}
}

func TestIInstallsFromTheRegistryList(t *testing.T) {
	t.Parallel()

	l := newLoader(t, fixtureRegistry, nil)
	seed(t, l, fixtureRegistry, time.Hour)

	root := demoProject(t, "schema: spec-driven\n")
	installer := installerFor(t, root, schemaDir(t, "chain.yaml"), &openspec.Fake{})

	s := newRegistryScreen(l, &Resolver{ASCII: true}, installer)
	_, _ = s.Update(l.Load(context.Background()))

	_, cmd := key(s, "i")
	if cmd == nil {
		t.Fatal("i produced no command")
	}

	_, _ = s.Update(cmd())

	if got := view(s); !strings.Contains(got, "write these files") {
		t.Errorf("the install preview did not open:\n%s", got)
	}
	if !s.Capturing() {
		t.Error("the install flow does not capture keys, so a tab key would leave it open")
	}
}

func TestIOnABuiltInRowReportsTheRefusalInTheList(t *testing.T) {
	t.Parallel()

	l := newLoader(t, fixtureRegistry, nil)
	l.CLI = &openspec.Fake{SchemasResult: []openspec.ResolvedSchema{{Name: "aaa-builtin", Source: openspec.SourcePackage, Path: "/p"}}}
	seed(t, l, fixtureRegistry, time.Hour)

	root := demoProject(t, "schema: spec-driven\n")
	s := newRegistryScreen(l, &Resolver{ASCII: true}, installerFor(t, root, schemaDir(t, "chain.yaml"), &openspec.Fake{}))
	_, _ = s.Update(l.Load(context.Background()))

	_, cmd := key(s, "g", "i")
	if cmd != nil {
		t.Error("i on a built-in row produced a command")
	}
	if got := view(s); !strings.Contains(got, "ships with OpenSpec") {
		t.Errorf("the refusal is not shown in the list:\n%s", got)
	}
}

func TestTheInstallPreviewTruncatesALongFileList(t *testing.T) {
	t.Parallel()

	src := schemaDir(t, "chain.yaml")
	for i := range 20 {
		path := filepath.Join(src, "templates", "extra", "file-"+string(rune('a'+i))+".md")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("creating %s: %v", filepath.Dir(path), err)
		}
		if err := os.WriteFile(path, []byte("# extra\n"), 0o644); err != nil {
			t.Fatalf("writing %s: %v", path, err)
		}
	}

	root := demoProject(t, "schema: spec-driven\n")
	f := newInstallFlow(installerFor(t, root, src, &openspec.Fake{}))

	cmd, _ := f.start(registryRow("minimalist"))
	_ = f.Update(cmd())

	if got := flowView(f); !strings.Contains(got, "more") {
		t.Errorf("a long file list is not truncated:\n%s", got)
	}
}

func TestWritingAPlanThatCannotBeWritten(t *testing.T) {
	t.Parallel()

	root := demoProject(t, "schema: spec-driven\n")
	installer := installerFor(t, root, schemaDir(t, "chain.yaml"), &openspec.Fake{})

	plan, err := openspec.PlanInstall(root, schemaDir(t, "chain.yaml"), "minimalist")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	plan.Source = filepath.Join(t.TempDir(), "gone")

	msg := installer.WriteCmd(plan, false)()

	finished, ok := msg.(InstallFinished)
	if !ok {
		t.Fatalf("WriteCmd returned %T", msg)
	}
	if finished.Err == nil {
		t.Fatal("writing from a source that does not exist succeeded")
	}
}

func TestTheFlowIsInactiveBeforeItStarts(t *testing.T) {
	t.Parallel()

	f := newInstallFlow(nil)

	if f.active() {
		t.Error("a fresh flow is active")
	}
	if got := flowView(f); got != "" {
		t.Errorf("an idle flow draws %q", got)
	}
	if cmd := f.Update(keyPress("y")); cmd != nil {
		t.Error("an idle flow answered a key")
	}
}

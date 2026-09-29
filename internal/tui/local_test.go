package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/speclib/openspec-schema-manager/internal/ansi"
	"github.com/speclib/openspec-schema-manager/internal/config"
	"github.com/speclib/openspec-schema-manager/internal/source"
)

func localSchema(t *testing.T, parent, folder, name string) string {
	t.Helper()

	dir := filepath.Join(parent, folder)

	writeFile(t, filepath.Join(dir, "schema.yaml"), "name: "+name+`
version: 1
description: a schema for testing
artifacts:
  - id: specs
    generates: specs.md
    description: the specifications
    template: spec.md
  - id: tasks
    generates: tasks.md
    description: the tasks
    template: tasks.md
    requires: [specs]
apply:
  requires: [tasks]
  tracks: tasks.md
`)
	writeFile(t, filepath.Join(dir, "templates", "spec.md"), "# spec\n")
	writeFile(t, filepath.Join(dir, "templates", "tasks.md"), "# tasks\n")

	return dir
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("creating %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
}

func localWith(t *testing.T, dirs []string) *localModel {
	t.Helper()

	recents := config.Recents{Path: filepath.Join(t.TempDir(), "recents.json"), Cap: 20}

	return newLocalScreen(dirs, recents, &Resolver{ASCII: true})
}

func localView(m *localModel) string { return flat(ansi.Strip(m.View(120, 30))) }

func TestTheLocalTabListsConfiguredDirectories(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	localSchema(t, dir, "team-review", "team-review")
	localSchema(t, dir, "minimalist", "minimalist")

	m := localWith(t, []string{dir})

	got := localView(m)
	for _, want := range []string{"2 schema(s)", dir, "minimalist", "team-review", "2 artifacts"} {
		if !strings.Contains(got, want) {
			t.Errorf("the local view does not carry %q:\n%s", want, got)
		}
	}
}

func TestNoConfiguredDirectoriesSaysWhatToDo(t *testing.T) {
	t.Parallel()

	m := localWith(t, nil)

	got := localView(m)
	for _, want := range []string{"No local schemas", "schemas_dirs", "press : to open one by path"} {
		if !strings.Contains(got, want) {
			t.Errorf("the local view does not carry %q:\n%s", want, got)
		}
	}
}

func TestAMissingDirectoryIsReported(t *testing.T) {
	t.Parallel()

	m := localWith(t, []string{filepath.Join(t.TempDir(), "absent")})

	if got := localView(m); !strings.Contains(got, "could not be read") {
		t.Errorf("a missing directory is not reported:\n%s", got)
	}
}

func TestAnUnreadableSchemaIsListed(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "broken", "schema.yaml"), "name: [unclosed\n")

	if got := localView(localWith(t, []string{dir})); !strings.Contains(got, "unreadable") {
		t.Errorf("an unreadable schema is not listed:\n%s", got)
	}
}

func TestRememberingAPath(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	awkward := localSchema(t, root, "somewhere-odd", "quick")

	m := localWith(t, nil)
	m.Remember(awkward)

	got := localView(m)
	if !strings.Contains(got, "Recent") {
		t.Errorf("the recents section is missing:\n%s", got)
	}
	if !strings.Contains(got, "quick") {
		t.Errorf("the remembered schema is not listed:\n%s", got)
	}

	if row := m.selectedRow(); row == nil || row.local.Dir != awkward {
		t.Errorf("the remembered path was not selected: %+v", row)
	}
}

func TestARememberedPathThatHasGoneIsShownAsMissing(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	gone := localSchema(t, root, "gone", "gone")

	m := localWith(t, nil)
	m.Remember(gone)

	if err := os.RemoveAll(gone); err != nil {
		t.Fatalf("removing the folder: %v", err)
	}
	m.reload()

	got := localView(m)
	if !strings.Contains(got, "missing") {
		t.Errorf("a remembered path that has gone is not shown as missing:\n%s", got)
	}
	if !strings.Contains(got, filepath.Base(gone)) {
		t.Errorf("the missing path is not named:\n%s", got)
	}
}

func TestOpeningALocalSchema(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	localSchema(t, dir, "minimalist", "minimalist")

	m := localWith(t, []string{dir})

	_, cmd := m.Update(keyPress("enter"))
	if cmd == nil {
		t.Fatal("enter produced no command")
	}
	_, _ = m.Update(cmd())

	if got := localView(m); !strings.Contains(got, "apply gate") {
		t.Errorf("the detail did not open:\n%s", got)
	}

	_, _ = m.Update(keyPress("esc"))
	if got := localView(m); !strings.Contains(got, "Local ·") {
		t.Errorf("esc did not return to the list:\n%s", got)
	}
}

func TestOpeningAMissingSchemaDoesNothing(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	gone := localSchema(t, root, "gone", "gone")

	m := localWith(t, nil)
	m.Remember(gone)
	if err := os.RemoveAll(gone); err != nil {
		t.Fatalf("removing the folder: %v", err)
	}
	m.reload()

	if _, cmd := m.Update(keyPress("enter")); cmd != nil {
		t.Error("enter on a missing schema produced a command")
	}
}

func TestBrowsingTheFileTree(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	localSchema(t, dir, "minimalist", "minimalist")

	m := localWith(t, []string{dir})
	_, _ = m.Update(keyPress("t"))

	got := localView(m)
	for _, want := range []string{"Files of minimalist", "schema.yaml", "spec.md", "tasks.md", "This schema validates"} {
		if !strings.Contains(got, want) {
			t.Errorf("the file tree does not carry %q:\n%s", want, got)
		}
	}

	_, _ = m.Update(keyPress("t"))
	if got := localView(m); strings.Contains(got, "Files of") {
		t.Errorf("t did not close the tree:\n%s", got)
	}
}

func TestTheTreeShowsFindings(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	schemaAt := localSchema(t, dir, "minimalist", "minimalist")

	if err := os.Remove(filepath.Join(schemaAt, "templates", "spec.md")); err != nil {
		t.Fatalf("removing the template: %v", err)
	}

	m := localWith(t, []string{dir})
	_, _ = m.Update(keyPress("t"))

	got := localView(m)
	if !strings.Contains(got, "fatal problem") {
		t.Errorf("the missing template is not reported:\n%s", got)
	}
	if !strings.Contains(got, "specs") {
		t.Errorf("the finding does not name the artifact:\n%s", got)
	}
}

func TestMovingThroughTheFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	localSchema(t, dir, "minimalist", "minimalist")

	m := localWith(t, []string{dir})
	_, _ = m.Update(keyPress("t"))

	for range 10 {
		_, _ = m.Update(keyPress("down"))
	}
	if m.fileIndex != len(m.files)-1 {
		t.Errorf("moving down past the end left the cursor at %d of %d", m.fileIndex, len(m.files))
	}

	for range 10 {
		_, _ = m.Update(keyPress("up"))
	}
	if m.fileIndex != 0 {
		t.Errorf("moving up past the start left the cursor at %d", m.fileIndex)
	}
}

func TestEditingRunsTheEditorAndRevalidates(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	schemaAt := localSchema(t, dir, "minimalist", "minimalist")

	missing := filepath.Join(schemaAt, "templates", "spec.md")
	if err := os.Remove(missing); err != nil {
		t.Fatalf("removing the template: %v", err)
	}

	m := localWith(t, []string{dir})
	m.editor = func() (string, []string) { return "touch", nil }

	_, _ = m.Update(keyPress("t"))

	if got := localView(m); !strings.Contains(got, "fatal problem") {
		t.Fatalf("the missing template was not reported first:\n%s", got)
	}

	// Put the cursor on schema.yaml's directory sibling and edit the file the
	// finding is about, by writing it the way the editor would.
	if err := os.WriteFile(missing, []byte("# spec\n"), 0o644); err != nil {
		t.Fatalf("writing the template: %v", err)
	}

	_, _ = m.Update(EditorFinished{})

	if got := localView(m); strings.Contains(got, "fatal problem") {
		t.Errorf("the finding survived the edit:\n%s", got)
	}
	if got := localView(m); !strings.Contains(got, "This schema validates") {
		t.Errorf("the schema was not revalidated:\n%s", got)
	}
}

func TestAnEditorThatFailsIsReported(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	localSchema(t, dir, "minimalist", "minimalist")

	m := localWith(t, []string{dir})
	_, _ = m.Update(keyPress("t"))

	_, _ = m.Update(EditorFinished{Err: os.ErrPermission})

	if got := localView(m); !strings.Contains(got, "the editor reported") {
		t.Errorf("an editor failure is not reported:\n%s", got)
	}
}

func TestAnEditThatBreaksTheSchemaIsReported(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	schemaAt := localSchema(t, dir, "minimalist", "minimalist")

	m := localWith(t, []string{dir})
	_, _ = m.Update(keyPress("t"))

	if err := os.WriteFile(filepath.Join(schemaAt, "schema.yaml"), []byte("name: [unclosed\n"), 0o644); err != nil {
		t.Fatalf("breaking the schema: %v", err)
	}

	_, _ = m.Update(EditorFinished{})

	got := localView(m)
	if !strings.Contains(got, "schema.yaml") {
		t.Errorf("the parse error is not reported:\n%s", got)
	}
	if !strings.Contains(got, "Files of") {
		t.Errorf("the tree stopped being usable:\n%s", got)
	}
}

func TestEditProducesACommand(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	localSchema(t, dir, "minimalist", "minimalist")

	m := localWith(t, []string{dir})
	m.editor = func() (string, []string) { return "true", nil }

	_, _ = m.Update(keyPress("t"))

	if _, cmd := m.Update(keyPress("e")); cmd == nil {
		t.Error("e produced no command")
	}
}

func TestEditWithNoFileSelectedDoesNothing(t *testing.T) {
	t.Parallel()

	m := localWith(t, nil)
	m.tree = true

	if _, cmd := m.Update(keyPress("e")); cmd != nil {
		t.Error("e produced a command with no files")
	}
}

func TestDuplicating(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	localSchema(t, dir, "minimalist", "minimalist")

	m := localWith(t, []string{dir})

	_, _ = m.Update(keyPress("c"))
	if !m.Capturing() {
		t.Fatal("c did not put the screen into capturing mode")
	}

	got := localView(m)
	if !strings.Contains(got, "Duplicate minimalist") {
		t.Errorf("the duplicate prompt did not open:\n%s", got)
	}

	for _, key := range []string{"t", "e", "a", "m"} {
		_, _ = m.Update(keyPress(key))
	}

	if got := localView(m); !strings.Contains(got, "team") {
		t.Errorf("the typed name is not shown:\n%s", got)
	}

	_, _ = m.Update(keyPress("enter"))

	destination := filepath.Join(dir, "team")
	raw, err := os.ReadFile(filepath.Join(destination, "schema.yaml"))
	if err != nil {
		t.Fatalf("the copy was not written: %v", err)
	}
	if !strings.Contains(string(raw), "name: team") {
		t.Errorf("the copy does not declare the new name:\n%s", raw)
	}

	got = localView(m)
	if !strings.Contains(got, "copied to") {
		t.Errorf("the copy is not reported:\n%s", got)
	}
	if !strings.Contains(got, "2 schema(s)") {
		t.Errorf("the list was not reloaded:\n%s", got)
	}
}

func TestCancellingADuplicate(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	localSchema(t, dir, "minimalist", "minimalist")

	m := localWith(t, []string{dir})

	_, _ = m.Update(keyPress("c"))
	for _, key := range []string{"t", "e", "a", "m"} {
		_, _ = m.Update(keyPress(key))
	}
	_, _ = m.Update(keyPress("esc"))

	if m.Capturing() {
		t.Error("esc left the screen capturing")
	}
	if _, err := os.Stat(filepath.Join(dir, "team")); err == nil {
		t.Error("cancelling wrote the copy anyway")
	}
}

func TestBackspaceInTheDuplicateName(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	localSchema(t, dir, "minimalist", "minimalist")

	m := localWith(t, []string{dir})

	_, _ = m.Update(keyPress("c"))
	for _, key := range []string{"a", "b", "c"} {
		_, _ = m.Update(keyPress(key))
	}
	for range 5 {
		_, _ = m.Update(keyPress("backspace"))
	}

	if m.newName != "" {
		t.Errorf("newName = %q, want empty", m.newName)
	}
}

func TestDuplicatingWithABadNameIsReported(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	localSchema(t, dir, "minimalist", "minimalist")

	m := localWith(t, []string{dir})

	_, _ = m.Update(keyPress("c"))
	_, _ = m.Update(keyPress("enter"))

	if got := localView(m); !strings.Contains(got, "could not duplicate") {
		t.Errorf("an empty name is not reported:\n%s", got)
	}
}

func TestDuplicatingWithNoDirectoryConfigured(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	somewhere := localSchema(t, root, "quick", "quick")

	m := localWith(t, nil)
	m.Remember(somewhere)

	_, _ = m.Update(keyPress("c"))

	if m.Capturing() {
		t.Error("the duplicate prompt opened with no directory configured")
	}
	if got := localView(m); !strings.Contains(got, "no local schemas directory is configured") {
		t.Errorf("the refusal is not reported:\n%s", got)
	}
}

func TestRescanning(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	localSchema(t, dir, "minimalist", "minimalist")

	m := localWith(t, []string{dir})

	localSchema(t, dir, "added-later", "added-later")

	_, _ = m.Update(keyPress("r"))

	got := localView(m)
	if !strings.Contains(got, "added-later") {
		t.Errorf("the rescan did not find the new schema:\n%s", got)
	}
	if !strings.Contains(got, "scanned again") {
		t.Errorf("the rescan is not reported:\n%s", got)
	}
}

func TestMovingThroughTheLocalList(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	for _, name := range []string{"a", "b", "c"} {
		localSchema(t, dir, name, name)
	}

	m := localWith(t, []string{dir})

	for range 10 {
		_, _ = m.Update(keyPress("down"))
	}
	if m.selected != 2 {
		t.Errorf("moving down past the end left the selection at %d", m.selected)
	}

	_, _ = m.Update(keyPress("g"))
	if m.selected != 0 {
		t.Errorf("g left the selection at %d", m.selected)
	}

	_, _ = m.Update(keyPress("G"))
	if m.selected != 2 {
		t.Errorf("G left the selection at %d", m.selected)
	}
}

func TestTheLocalScreenReportsItsKeys(t *testing.T) {
	t.Parallel()

	m := localWith(t, nil)

	var found []string
	for _, k := range m.Keys() {
		found = append(found, k.Key)
	}
	for _, want := range []string{"enter", "t", "c", "r"} {
		if !contains(found, want) {
			t.Errorf("the local screen does not report %q among %v", want, found)
		}
	}

	m.tree = true
	found = nil
	for _, k := range m.Keys() {
		found = append(found, k.Key)
	}
	if !contains(found, "e") {
		t.Errorf("the file tree does not report e among %v", found)
	}

	m.tree = false
	m.duplicating = true
	found = nil
	for _, k := range m.Keys() {
		found = append(found, k.Key)
	}
	if !contains(found, "enter") {
		t.Errorf("the duplicate prompt does not report enter among %v", found)
	}

	if m.Title() != "Local" {
		t.Errorf("title = %q", m.Title())
	}
}

func TestAnUnknownMessageIsIgnoredByTheLocalScreen(t *testing.T) {
	t.Parallel()

	type odd struct{}

	m := localWith(t, nil)

	if _, cmd := m.Update(odd{}); cmd != nil {
		t.Error("an unknown message produced a command")
	}
}

func TestTheLocalListScrolls(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	for i := range 20 {
		name := "schema-" + string(rune('a'+i))
		localSchema(t, dir, name, name)
	}

	m := localWith(t, []string{dir})

	before := ansi.Strip(m.View(80, 10))

	for range 15 {
		_, _ = m.Update(keyPress("down"))
	}

	if after := ansi.Strip(m.View(80, 10)); after == before {
		t.Errorf("the list did not scroll:\n%s", before)
	}
}

func TestTheTreeWithNoFilesSaysSo(t *testing.T) {
	t.Parallel()

	m := localWith(t, nil)
	m.tree = true
	m.rows = []localRow{{local: source.Local{Dir: t.TempDir(), Name: "bare"}}}

	if got := localView(m); !strings.Contains(got, "holds no schema files") {
		t.Errorf("a folder with no schema files is not reported:\n%s", got)
	}
}

func TestTheTreeShowsWarningsOnTheirOwn(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	base := filepath.Join(dir, "warned")

	writeFile(t, filepath.Join(base, "schema.yaml"), `name: warned
version: 1
description: a schema with a warning
artifacts:
  - id: specs
    generates: specs.md
    description: the specifications
    template: spec.md
  - id: sketch
    generates: sketch.md
    description: an artifact nothing reaches
    template: sketch.md
apply:
  requires: [specs]
  tracks: specs.md
`)
	writeFile(t, filepath.Join(base, "templates", "spec.md"), "# spec\n")
	writeFile(t, filepath.Join(base, "templates", "sketch.md"), "# sketch\n")

	m := localWith(t, []string{dir})
	_, _ = m.Update(keyPress("t"))

	got := localView(m)
	if !strings.Contains(got, "warning(s)") {
		t.Errorf("the warning is not reported:\n%s", got)
	}
	if strings.Contains(got, "fatal problem") {
		t.Errorf("a schema with only warnings was reported as broken:\n%s", got)
	}
}

func TestOpeningATreeOnAFolderThatHasGone(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	dir := localSchema(t, root, "gone", "gone")

	m := localWith(t, nil)
	m.Remember(dir)

	if err := os.RemoveAll(dir); err != nil {
		t.Fatalf("removing the folder: %v", err)
	}

	m.rows[0].missing = false

	_, _ = m.Update(keyPress("t"))

	if got := localView(m); !strings.Contains(got, "could not read that folder") {
		t.Errorf("a folder that has gone is not reported:\n%s", got)
	}
}

func TestOperationsWithNothingSelected(t *testing.T) {
	t.Parallel()

	m := localWith(t, nil)

	if row := m.selectedRow(); row != nil {
		t.Fatalf("an empty list has a selection: %+v", row)
	}

	m.openTree()
	if m.tree {
		t.Error("the tree opened with nothing selected")
	}

	m.startDuplicate()
	if m.duplicating {
		t.Error("the duplicate prompt opened with nothing selected")
	}

	m.finishDuplicate()
	m.revalidate()
}

func TestRememberingAPathThatCannotBeWritten(t *testing.T) {
	t.Parallel()

	if os.Geteuid() == 0 {
		t.Skip("running as root, so an unwritable directory is still writable")
	}

	root := t.TempDir()
	blocked := filepath.Join(root, "blocked")
	if err := os.Mkdir(blocked, 0o500); err != nil {
		t.Fatalf("creating the blocked directory: %v", err)
	}

	m := newLocalScreen(nil, config.Recents{Path: filepath.Join(blocked, "state", "recents.json"), Cap: 20}, nil)
	m.Remember(t.TempDir())

	if got := localView(m); !strings.Contains(got, "could not remember that path") {
		t.Errorf("a recents write failure is not reported:\n%s", got)
	}
}

func TestMovingFilesWithNoneLoaded(t *testing.T) {
	t.Parallel()

	m := localWith(t, nil)
	m.moveFile(1)

	if m.fileIndex != 0 {
		t.Errorf("moving with no files left the cursor at %d", m.fileIndex)
	}
}

func TestOpeningWithNoResolver(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	localSchema(t, dir, "minimalist", "minimalist")

	m := newLocalScreen([]string{dir}, config.Recents{Path: filepath.Join(t.TempDir(), "r.json"), Cap: 20}, nil)

	if _, cmd := m.Update(keyPress("enter")); cmd != nil {
		t.Error("enter produced a command with no resolver")
	}
}

func TestTheDetailTakesMessagesInTheLocalScreen(t *testing.T) {
	t.Parallel()

	type odd struct{}

	dir := t.TempDir()
	localSchema(t, dir, "minimalist", "minimalist")

	m := localWith(t, []string{dir})

	_, cmd := m.Update(keyPress("enter"))
	_, _ = m.Update(cmd())

	if _, cmd := m.Update(odd{}); cmd != nil {
		t.Error("an unknown message in the detail produced a command")
	}
	if !strings.Contains(localView(m), "apply gate") {
		t.Error("the detail closed on an unknown message")
	}
}

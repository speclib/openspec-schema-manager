package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/speclib/openspec-schema-manager/internal/ansi"
	"github.com/speclib/openspec-schema-manager/internal/compose"
	"github.com/speclib/openspec-schema-manager/internal/schema"
)

func sourceSchema(t *testing.T, dir, name string, artifacts []string, templates map[string]string) (schema.Schema, string) {
	t.Helper()

	var b strings.Builder
	b.WriteString("name: " + name + "\nversion: 1\ndescription: a schema for testing\nartifacts:\n")

	for i, id := range artifacts {
		b.WriteString("  - id: " + id + "\n")
		b.WriteString("    generates: " + id + ".md\n")
		b.WriteString("    description: what " + id + " produces\n")
		b.WriteString("    template: " + id + ".md\n")
		if i > 0 {
			b.WriteString("    requires: [" + artifacts[i-1] + "]\n")
		}
	}

	last := artifacts[len(artifacts)-1]
	b.WriteString("apply:\n  requires: [" + last + "]\n  tracks: " + last + ".md\n")

	writeFile(t, filepath.Join(dir, "schema.yaml"), b.String())

	for _, id := range artifacts {
		body := templates[id]
		if body == "" {
			body = "# " + id + "\n"
		}
		writeFile(t, filepath.Join(dir, "templates", id+".md"), body)
	}

	s, err := schema.Parse("schema.yaml", []byte(b.String()))
	if err != nil {
		t.Fatalf("parsing the fixture: %v", err)
	}

	return s, dir
}

func newComposer(t *testing.T, dirs []string) *composerModel {
	t.Helper()

	return newComposerScreen(dirs, compose.Drafts{Dir: filepath.Join(t.TempDir(), "drafts")}, true)
}

func composerView(c *composerModel) string { return flat(ansi.Strip(c.View(140, 30))) }

func withSource(t *testing.T, c *composerModel, name string, artifacts []string, templates map[string]string) string {
	t.Helper()

	s, dir := sourceSchema(t, t.TempDir(), name, artifacts, templates)

	_, _ = c.Update(SourceAdded{Schema: s, Dir: dir, Ref: "v1"})

	return dir
}

func keys(c *composerModel, keys ...string) {
	for _, k := range keys {
		_, _ = c.Update(keyPress(k))
	}
}

func TestAnEmptyComposerSaysWhatToDo(t *testing.T) {
	t.Parallel()

	got := composerView(newComposer(t, nil))

	for _, want := range []string{"0 artifact(s)", "no gate", "nothing tracked", "press p on a schema", "canvas is empty"} {
		if !strings.Contains(got, want) {
			t.Errorf("the composer does not carry %q:\n%s", want, got)
		}
	}
}

func TestAddingASourceFillsThePalette(t *testing.T) {
	t.Parallel()

	c := newComposer(t, nil)
	withSource(t, c, "research-first", []string{"research", "proposal", "tasks"}, nil)

	got := composerView(c)
	for _, want := range []string{"Palette", "research-first", "research", "proposal", "tasks", "added to the palette"} {
		if !strings.Contains(got, want) {
			t.Errorf("the palette does not carry %q:\n%s", want, got)
		}
	}
}

func TestAddingAnArtifactToTheCanvas(t *testing.T) {
	t.Parallel()

	c := newComposer(t, nil)
	withSource(t, c, "research-first", []string{"research", "proposal", "tasks"}, nil)

	keys(c, "a")

	got := composerView(c)
	if !strings.Contains(got, "1 artifact(s)") {
		t.Errorf("the artifact was not added:\n%s", got)
	}
	if !strings.Contains(got, "research added") {
		t.Errorf("the add is not reported:\n%s", got)
	}
	if !strings.Contains(got, "(on canvas)") {
		t.Errorf("the palette does not mark what is on the canvas:\n%s", got)
	}
}

func TestACollisionAsksForANewID(t *testing.T) {
	t.Parallel()

	c := newComposer(t, nil)
	withSource(t, c, "one", []string{"tasks"}, nil)
	withSource(t, c, "two", []string{"tasks"}, nil)

	keys(c, "a")
	keys(c, "down", "down", "a")

	if !c.Capturing() {
		t.Fatal("a collision did not open a prompt")
	}
	if got := composerView(c); !strings.Contains(got, "already on the canvas") {
		t.Errorf("the collision is not explained:\n%s", got)
	}

	keys(c, "t", "w", "o", "enter")

	got := composerView(c)
	if !strings.Contains(got, "2 artifact(s)") {
		t.Errorf("the renamed artifact was not added:\n%s", got)
	}
	if !strings.Contains(got, "two added") {
		t.Errorf("the add is not reported:\n%s", got)
	}
}

func TestACollisionThatCollidesAgain(t *testing.T) {
	t.Parallel()

	c := newComposer(t, nil)
	withSource(t, c, "one", []string{"tasks", "specs"}, nil)
	withSource(t, c, "two", []string{"tasks"}, nil)

	keys(c, "a", "down", "a")
	keys(c, "down", "down", "a")
	keys(c, "s", "p", "e", "c", "s", "enter")

	if got := composerView(c); !strings.Contains(got, "already on the canvas") {
		t.Errorf("a second collision was not refused:\n%s", got)
	}
}

func TestBuildingAWholeComposition(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	c := newComposer(t, []string{dir})

	withSource(t, c, "research-first", []string{"research", "proposal"}, nil)
	withSource(t, c, "team-review", []string{"review"}, nil)

	keys(c, "a")
	keys(c, "down", "a")
	keys(c, "down", "down", "a")

	if len(c.canvas.Artifacts) != 3 {
		t.Fatalf("got %d artifacts, want 3", len(c.canvas.Artifacts))
	}

	keys(c, "right")
	keys(c, "down", "down")
	keys(c, "l")

	if !c.Capturing() {
		t.Fatal("l did not open the link chooser")
	}
	if got := composerView(c); !strings.Contains(got, "requires:") {
		t.Errorf("the link chooser is not drawn:\n%s", got)
	}

	keys(c, "enter")
	keys(c, "g")
	keys(c, "t")
	keys(c, "r", "e", "v", "i", "e", "w", ".", "m", "d", "enter")

	got := composerView(c)
	if !strings.Contains(got, "gate: review") {
		t.Errorf("the gate was not set:\n%s", got)
	}
	if !strings.Contains(got, "tracks: review.md") {
		t.Errorf("the tracked file was not set:\n%s", got)
	}
	if !strings.Contains(got, "ready to write") {
		t.Errorf("the composition is not reported as ready:\n%s", got)
	}

	keys(c, "w")
	keys(c, "m", "i", "x", "enter")

	if got := composerView(c); !strings.Contains(got, "written to") {
		t.Errorf("the write is not reported:\n%s", got)
	}

	written, err := schema.Load(filepath.Join(dir, "mix"))
	if err != nil {
		t.Fatalf("the written schema does not parse: %v", err)
	}
	if len(written.Artifacts) != 3 {
		t.Errorf("the written schema holds %d artifacts", len(written.Artifacts))
	}
	if findings := schema.ValidateDir(written, filepath.Join(dir, "mix")); !findings.Valid() {
		t.Errorf("the written schema does not validate: %v", findings.Fatal())
	}
}

func TestACycleIsReportedAsItIsMade(t *testing.T) {
	t.Parallel()

	c := newComposer(t, nil)
	withSource(t, c, "chain", []string{"a", "b"}, nil)

	keys(c, "a", "down", "a")
	keys(c, "right")
	keys(c, "l", "enter")

	if got := composerView(c); !strings.Contains(got, "cycle") {
		t.Errorf("the cycle is not reported:\n%s", got)
	}
}

func TestRemovingTakesTheEdges(t *testing.T) {
	t.Parallel()

	c := newComposer(t, nil)
	withSource(t, c, "chain", []string{"a", "b"}, nil)

	keys(c, "a", "down", "a")
	keys(c, "right", "down")

	if got := composerView(c); !strings.Contains(got, "← a") {
		t.Errorf("the edge is not drawn:\n%s", got)
	}

	keys(c, "up", "x")

	got := composerView(c)
	if !strings.Contains(got, "a removed") {
		t.Errorf("the removal is not reported:\n%s", got)
	}
	if strings.Contains(got, "← a") {
		t.Errorf("the edge survived the removal:\n%s", got)
	}
}

func TestComposerUnlinking(t *testing.T) {
	t.Parallel()

	c := newComposer(t, nil)
	withSource(t, c, "chain", []string{"a", "b"}, nil)

	keys(c, "a", "down", "a")
	keys(c, "right", "down", "u")

	got := composerView(c)
	if !strings.Contains(got, "no longer requires a") {
		t.Errorf("the unlink is not reported:\n%s", got)
	}
	if strings.Contains(got, "← a") {
		t.Errorf("the edge survived:\n%s", got)
	}
}

func TestRenamingFromTheComposer(t *testing.T) {
	t.Parallel()

	c := newComposer(t, nil)
	withSource(t, c, "chain", []string{"a", "b"}, nil)

	keys(c, "a", "down", "a")
	keys(c, "right", "R")

	if !c.Capturing() {
		t.Fatal("R did not open a prompt")
	}

	keys(c, "backspace", "f", "i", "r", "s", "t", "enter")

	got := composerView(c)
	if !strings.Contains(got, "renamed to first") {
		t.Errorf("the rename is not reported:\n%s", got)
	}
	if !strings.Contains(got, "← first") {
		t.Errorf("the dependent's requirement was not renamed:\n%s", got)
	}
}

func TestARenameThatCollides(t *testing.T) {
	t.Parallel()

	c := newComposer(t, nil)
	withSource(t, c, "chain", []string{"a", "b"}, nil)

	keys(c, "a", "down", "a")
	keys(c, "right", "R")
	keys(c, "backspace", "b", "enter")

	if got := composerView(c); !strings.Contains(got, "already on the canvas") {
		t.Errorf("a colliding rename is not reported:\n%s", got)
	}
}

func TestATemplateWarning(t *testing.T) {
	t.Parallel()

	c := newComposer(t, nil)
	withSource(t, c, "research-first", []string{"research", "tasks"}, map[string]string{
		"tasks": "# Tasks\n\nCheck the research first.\n",
	})

	keys(c, "down", "a")

	got := composerView(c)
	if !strings.Contains(got, "template warning") {
		t.Errorf("the dangling reference is not warned about:\n%s", got)
	}
	if !strings.Contains(got, "research") {
		t.Errorf("the warning does not name the reference:\n%s", got)
	}

	keys(c, "up", "a")

	if got := composerView(c); strings.Contains(got, "template warning") {
		t.Errorf("the warning survived adding the artifact:\n%s", got)
	}
}

func TestTheComposerDiagramToggles(t *testing.T) {
	t.Parallel()

	c := newComposer(t, nil)
	withSource(t, c, "chain", []string{"a", "b"}, nil)

	keys(c, "a", "down", "a")
	keys(c, "d")

	got := composerView(c)
	if !strings.Contains(got, "d closes the diagram") {
		t.Errorf("the diagram did not open:\n%s", got)
	}
	if !strings.Contains(got, "a") || !strings.Contains(got, "b") {
		t.Errorf("the diagram does not carry the artifacts:\n%s", got)
	}

	keys(c, "d")
	if got := composerView(c); strings.Contains(got, "d closes the diagram") {
		t.Errorf("the diagram did not close:\n%s", got)
	}
}

func TestTheComposerDiagramReportsACycle(t *testing.T) {
	t.Parallel()

	c := newComposer(t, nil)
	withSource(t, c, "chain", []string{"a", "b"}, nil)

	keys(c, "a", "down", "a")
	keys(c, "right", "l", "enter")
	keys(c, "d")

	if got := composerView(c); !strings.Contains(got, "cannot be drawn") {
		t.Errorf("a cyclic composition draws:\n%s", got)
	}
}

func TestSavingAndResumingADraftFromTheComposer(t *testing.T) {
	t.Parallel()

	c := newComposer(t, nil)
	withSource(t, c, "chain", []string{"a", "b"}, nil)

	keys(c, "a", "down", "a")
	keys(c, "s")
	keys(c, "m", "i", "x", "enter")

	if got := composerView(c); !strings.Contains(got, "saved as a draft") {
		t.Errorf("the save is not reported:\n%s", got)
	}

	keys(c, "right", "x", "x")
	if len(c.canvas.Artifacts) != 0 {
		t.Fatalf("the canvas still holds %d artifacts", len(c.canvas.Artifacts))
	}

	keys(c, "o")
	if !c.Capturing() {
		t.Fatal("o did not open the draft list")
	}
	if got := composerView(c); !strings.Contains(got, "mix") {
		t.Errorf("the draft is not listed:\n%s", got)
	}

	keys(c, "enter")

	got := composerView(c)
	if !strings.Contains(got, "resumed mix") {
		t.Errorf("the resume is not reported:\n%s", got)
	}
	if len(c.canvas.Artifacts) != 2 {
		t.Errorf("the canvas holds %d artifacts after resuming", len(c.canvas.Artifacts))
	}
}

func TestOpeningDraftsWithNoneSaved(t *testing.T) {
	t.Parallel()

	c := newComposer(t, nil)

	keys(c, "o")

	if c.Capturing() {
		t.Error("the draft list opened with no drafts")
	}
	if got := composerView(c); !strings.Contains(got, "no saved drafts") {
		t.Errorf("the absence is not reported:\n%s", got)
	}
}

func TestTheComposerResumesADraftWhoseSourceHasGone(t *testing.T) {
	t.Parallel()

	c := newComposer(t, nil)
	dir := withSource(t, c, "chain", []string{"a", "b"}, nil)

	keys(c, "a")
	keys(c, "s", "m", "i", "x", "enter")

	if err := os.RemoveAll(dir); err != nil {
		t.Fatalf("removing the source: %v", err)
	}

	keys(c, "o", "enter")

	if got := composerView(c); !strings.Contains(got, "no longer there") {
		t.Errorf("the missing source is not reported:\n%s", got)
	}
}

func TestComposerWritingRefusals(t *testing.T) {
	t.Parallel()

	t.Run("no configured directory", func(t *testing.T) {
		t.Parallel()

		c := newComposer(t, nil)
		withSource(t, c, "chain", []string{"a"}, nil)

		keys(c, "a", "right", "g", "t")
		keys(c, "a", ".", "m", "d", "enter")
		keys(c, "w", "m", "i", "x", "enter")

		if got := composerView(c); !strings.Contains(got, "no local schemas directory is configured") {
			t.Errorf("the refusal is not reported:\n%s", got)
		}
	})

	t.Run("an invalid composition", func(t *testing.T) {
		t.Parallel()

		c := newComposer(t, []string{t.TempDir()})
		withSource(t, c, "chain", []string{"a"}, nil)

		keys(c, "a")
		keys(c, "w", "m", "i", "x", "enter")

		if got := composerView(c); !strings.Contains(got, "could not write") {
			t.Errorf("the refusal is not reported:\n%s", got)
		}
	})
}

func TestCancellingAPrompt(t *testing.T) {
	t.Parallel()

	c := newComposer(t, nil)
	withSource(t, c, "chain", []string{"a", "b"}, nil)

	keys(c, "a", "right", "R")
	keys(c, "x", "esc")

	if c.Capturing() {
		t.Error("esc left the prompt open")
	}
	if !c.canvas.Has("a") {
		t.Error("cancelling renamed the artifact anyway")
	}
}

func TestCancellingTheLinkChooser(t *testing.T) {
	t.Parallel()

	c := newComposer(t, nil)
	withSource(t, c, "chain", []string{"a", "b"}, nil)

	keys(c, "a", "down", "a")
	keys(c, "right", "down", "u")
	keys(c, "l", "esc")

	if c.Capturing() {
		t.Error("esc left the chooser open")
	}
	for _, artifact := range c.canvas.Artifacts {
		if len(artifact.Requires) != 0 {
			t.Errorf("cancelling added an edge: %v", artifact.Requires)
		}
	}
}

func TestLinkingWithNothingElseOnTheCanvas(t *testing.T) {
	t.Parallel()

	c := newComposer(t, nil)
	withSource(t, c, "chain", []string{"a"}, nil)

	keys(c, "a", "right", "l")

	if c.Capturing() {
		t.Error("the chooser opened with nothing to link to")
	}
	if got := composerView(c); !strings.Contains(got, "nothing else on the canvas") {
		t.Errorf("the reason is not reported:\n%s", got)
	}
}

func TestOperationsWithAnEmptyCanvas(t *testing.T) {
	t.Parallel()

	c := newComposer(t, nil)

	keys(c, "right", "x", "l", "u", "g", "R", "a")

	if len(c.canvas.Artifacts) != 0 {
		t.Errorf("something was added to an empty canvas: %v", c.canvas.IDs())
	}
	if c.Capturing() {
		t.Error("a prompt opened with nothing selected")
	}
}

func TestComposerMovementStaysInBounds(t *testing.T) {
	t.Parallel()

	c := newComposer(t, nil)
	withSource(t, c, "chain", []string{"a", "b", "c"}, nil)

	for range 20 {
		keys(c, "down")
	}
	if entry, ok := c.selectedPalette(); !ok || entry.artifact.ID != "c" {
		t.Errorf("moving down past the end left the palette at %+v", entry)
	}

	for range 20 {
		keys(c, "up")
	}
	if entry, ok := c.selectedPalette(); !ok || entry.artifact.ID != "a" {
		t.Errorf("moving up past the start left the palette at %+v", entry)
	}

	keys(c, "a", "right")
	for range 20 {
		keys(c, "down")
	}
	if c.canvasIndex != 0 {
		t.Errorf("the canvas cursor is at %d with one artifact", c.canvasIndex)
	}
}

func TestTheComposerReportsItsKeys(t *testing.T) {
	t.Parallel()

	c := newComposer(t, nil)

	var found []string
	for _, k := range c.Keys() {
		found = append(found, k.Key)
	}
	for _, want := range []string{"a", "x", "l", "g", "t", "R", "w", "s"} {
		if !contains(found, want) {
			t.Errorf("the composer does not report %q among %v", want, found)
		}
	}

	c.mode = composerLinking
	found = nil
	for _, k := range c.Keys() {
		found = append(found, k.Key)
	}
	if !contains(found, "enter") {
		t.Errorf("the link chooser does not report enter among %v", found)
	}

	c.mode = composerNamingWrite
	found = nil
	for _, k := range c.Keys() {
		found = append(found, k.Key)
	}
	if !contains(found, "esc") {
		t.Errorf("a prompt does not report esc among %v", found)
	}

	if c.Title() != "Composer" {
		t.Errorf("title = %q", c.Title())
	}
}

func TestAnUnknownMessageIsIgnoredByTheComposer(t *testing.T) {
	t.Parallel()

	type odd struct{}

	c := newComposer(t, nil)

	if _, cmd := c.Update(odd{}); cmd != nil {
		t.Error("an unknown message produced a command")
	}
}

func TestMovingInsideTheChoosers(t *testing.T) {
	t.Parallel()

	c := newComposer(t, nil)
	withSource(t, c, "chain", []string{"a", "b", "d"}, nil)

	keys(c, "a", "down", "a", "down", "a")
	keys(c, "right", "down", "down")
	keys(c, "l")

	targets := c.linkTargets()
	if len(targets) != 2 {
		t.Fatalf("got %d link targets, want 2", len(targets))
	}

	for range 10 {
		keys(c, "down")
	}
	if c.linkIndex != len(targets)-1 {
		t.Errorf("moving down past the end left the chooser at %d", c.linkIndex)
	}

	for range 10 {
		keys(c, "up")
	}
	if c.linkIndex != 0 {
		t.Errorf("moving up past the start left the chooser at %d", c.linkIndex)
	}

	keys(c, "enter")

	keys(c, "s", "o", "n", "e", "enter")
	keys(c, "s", "t", "w", "o", "enter")
	keys(c, "o")

	for range 10 {
		keys(c, "down")
	}
	if c.draftIndex != len(c.draftList)-1 {
		t.Errorf("moving down past the end left the draft list at %d", c.draftIndex)
	}

	for range 10 {
		keys(c, "up")
	}
	if c.draftIndex != 0 {
		t.Errorf("moving up past the start left the draft list at %d", c.draftIndex)
	}
}

func TestOperationsOnAnArtifactThatIsNotThere(t *testing.T) {
	t.Parallel()

	c := newComposer(t, nil)
	withSource(t, c, "chain", []string{"a"}, nil)

	keys(c, "a", "right")

	c.canvasIndex = 5

	c.remove()
	c.unlink()
	c.toggleGate()
	c.startRename()
	c.startLink()

	if c.Capturing() {
		t.Error("a prompt opened for an artifact that is not selected")
	}
}

func TestSavingADraftWithABadName(t *testing.T) {
	t.Parallel()

	c := newComposer(t, nil)
	withSource(t, c, "chain", []string{"a"}, nil)

	keys(c, "a")
	keys(c, "s", "enter")

	if got := composerView(c); !strings.Contains(got, "could not save the draft") {
		t.Errorf("a bad draft name is not reported:\n%s", got)
	}
}

func TestOpeningACorruptDraft(t *testing.T) {
	t.Parallel()

	drafts := compose.Drafts{Dir: filepath.Join(t.TempDir(), "drafts")}
	if err := os.MkdirAll(drafts.Dir, 0o755); err != nil {
		t.Fatalf("creating the drafts directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(drafts.Dir, "bad.json"), []byte("{oops"), 0o644); err != nil {
		t.Fatalf("writing the draft: %v", err)
	}

	c := newComposerScreen(nil, drafts, true)

	keys(c, "o")
	if !c.Capturing() {
		t.Fatal("the draft list did not open")
	}
	if got := composerView(c); !strings.Contains(got, "could not be read") {
		t.Errorf("the corrupt draft is not marked:\n%s", got)
	}

	keys(c, "enter")

	if got := composerView(c); !strings.Contains(got, "could not open that draft") {
		t.Errorf("the failure is not reported:\n%s", got)
	}
}

func TestTheCanvasListScrolls(t *testing.T) {
	t.Parallel()

	var artifacts []string
	for i := range 20 {
		artifacts = append(artifacts, "artifact-"+string(rune('a'+i)))
	}

	c := newComposer(t, nil)
	withSource(t, c, "many", artifacts, nil)

	for range len(artifacts) {
		keys(c, "a", "down")
	}

	keys(c, "right")

	before := ansi.Strip(c.View(120, 12))
	for range 15 {
		keys(c, "down")
	}

	if after := ansi.Strip(c.View(120, 12)); after == before {
		t.Errorf("the canvas did not scroll:\n%s", before)
	}

	keys(c, "left")
	for range 25 {
		keys(c, "up")
	}

	before = ansi.Strip(c.View(120, 12))
	for range 15 {
		keys(c, "down")
	}
	if after := ansi.Strip(c.View(120, 12)); after == before {
		t.Errorf("the palette did not scroll:\n%s", before)
	}
}

func TestManyProblemsAreTruncated(t *testing.T) {
	t.Parallel()

	var artifacts []string
	for i := range 8 {
		artifacts = append(artifacts, "artifact-"+string(rune('a'+i)))
	}

	c := newComposer(t, nil)
	withSource(t, c, "many", artifacts, nil)

	for range len(artifacts) {
		keys(c, "a", "down")
	}

	for i := range c.canvas.Artifacts {
		c.canvas.Artifacts[i].Generates = "same.md"
	}

	if got := composerView(c); !strings.Contains(got, "problem(s)") {
		t.Errorf("the problems are not reported:\n%s", got)
	}
}

func TestManyWarningsAreTruncated(t *testing.T) {
	t.Parallel()

	templates := map[string]string{}
	var artifacts []string
	for i := range 8 {
		id := "artifact-" + string(rune('a'+i))
		artifacts = append(artifacts, id)
		templates[id] = "# mentions artifact-x artifact-y artifact-z artifact-w artifact-v\n"
	}
	for _, extra := range []string{"artifact-x", "artifact-y", "artifact-z", "artifact-w", "artifact-v"} {
		artifacts = append(artifacts, extra)
	}

	c := newComposer(t, nil)
	withSource(t, c, "many", artifacts, templates)

	for range 8 {
		keys(c, "a", "down")
	}

	got := composerView(c)
	if !strings.Contains(got, "template warning") {
		t.Errorf("the warnings are not reported:\n%s", got)
	}
	if !strings.Contains(got, "and") {
		t.Errorf("a long warning list is not truncated:\n%s", got)
	}
}

func TestATemplateThatCannotBeReadIsShown(t *testing.T) {
	t.Parallel()

	c := newComposer(t, nil)
	dir := withSource(t, c, "chain", []string{"a", "b"}, nil)

	keys(c, "down", "a")

	if err := os.Remove(filepath.Join(dir, "templates", "b.md")); err != nil {
		t.Fatalf("removing the template: %v", err)
	}

	if got := composerView(c); !strings.Contains(got, "could not be read") {
		t.Errorf("the unreadable template is not reported:\n%s", got)
	}
}

func TestTheComposerReportsAnOperationThatFails(t *testing.T) {
	t.Parallel()

	c := newComposer(t, nil)
	withSource(t, c, "chain", []string{"a", "b"}, nil)

	keys(c, "a", "down", "a")
	keys(c, "right")

	// Point the cursor at an artifact the canvas does not hold, which is what
	// happens if a draft is resumed underneath a stale selection.
	c.canvas.Artifacts = append(c.canvas.Artifacts, compose.Artifact{ID: "ghost"})
	c.canvasIndex = len(c.canvas.Artifacts) - 1
	c.canvas.Artifacts = c.canvas.Artifacts[:len(c.canvas.Artifacts)-1]

	c.remove()
	c.unlink()
	c.toggleGate()

	if got := composerView(c); got == "" {
		t.Error("the composer drew nothing after an operation on a stale selection")
	}
}

func TestConfirmingAChoiceOutOfRange(t *testing.T) {
	t.Parallel()

	c := newComposer(t, nil)
	withSource(t, c, "chain", []string{"a", "b"}, nil)

	keys(c, "a", "down", "a")
	keys(c, "right", "l")

	c.linkIndex = 99
	keys(c, "enter")

	if c.Capturing() {
		t.Error("an out-of-range choice left the chooser open")
	}

	keys(c, "s", "m", "i", "x", "enter")
	keys(c, "o")

	c.draftIndex = 99
	keys(c, "enter")

	if c.Capturing() {
		t.Error("an out-of-range draft choice left the list open")
	}
}

func TestThePanesAtASmallWidth(t *testing.T) {
	t.Parallel()

	c := newComposer(t, nil)
	withSource(t, c, "chain", []string{"a", "b"}, nil)

	keys(c, "a")

	for _, width := range []int{10, 30, 60, 200} {
		if got := ansi.Strip(c.View(width, 20)); got == "" {
			t.Errorf("the composer drew nothing at width %d", width)
		}
	}
	if got := ansi.Strip(c.View(80, 2)); got == "" {
		t.Error("the composer drew nothing at a tiny height")
	}
}

func TestAPaletteWithOnlyHeadings(t *testing.T) {
	t.Parallel()

	c := newComposer(t, nil)

	_, _ = c.Update(SourceAdded{
		Schema: schema.Schema{Name: "empty"},
		Dir:    t.TempDir(),
		Ref:    "v1",
	})

	if _, ok := c.selectedPalette(); ok {
		t.Error("a palette holding only a heading has a selection")
	}

	keys(c, "a", "down", "up")

	if len(c.canvas.Artifacts) != 0 {
		t.Errorf("something was added from a heading: %v", c.canvas.IDs())
	}
}

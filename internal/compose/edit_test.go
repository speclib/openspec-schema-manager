package compose

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/speclib/openspec-schema-manager/internal/schema"
)

func chain(t *testing.T) (*Composition, Source, schema.Schema) {
	t.Helper()

	src, s := sourceDir(t, researchFirst, nil)

	c := New()
	c.AddSource(s, src.Dir, src.Ref)

	for _, id := range []string{"research", "proposal", "tasks"} {
		if err := c.Add(src, artifact(s, id), ""); err != nil {
			t.Fatalf("adding %s: %v", id, err)
		}
	}

	if err := c.ToggleGate("tasks"); err != nil {
		t.Fatalf("setting the gate: %v", err)
	}
	c.SetTracks("tasks.md")

	return c, src, s
}

func TestRemoveTakesItsEdgesAndItsGate(t *testing.T) {
	t.Parallel()

	c, _, _ := chain(t)

	if err := c.Remove("proposal"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.Has("proposal") {
		t.Error("the artifact is still on the canvas")
	}
	for _, a := range c.Artifacts {
		for _, req := range a.Requires {
			if req == "proposal" {
				t.Errorf("%s still requires the removed artifact", a.ID)
			}
		}
	}

	if err := c.Remove("tasks"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.IsGate("tasks") {
		t.Error("the removed artifact still gates apply")
	}
	if len(c.Gates) != 0 {
		t.Errorf("gates = %v", c.Gates)
	}

	if err := c.Remove("absent"); !errors.Is(err, ErrNotOnCanvas) {
		t.Errorf("error = %v, want ErrNotOnCanvas", err)
	}
}

func TestLinking(t *testing.T) {
	t.Parallel()

	c, _, _ := chain(t)

	if err := c.Unlink("tasks", "proposal"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := c.Link("tasks", "research"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, a := range c.Artifacts {
		if a.ID == "tasks" && !reflect.DeepEqual(a.Requires, []string{"research"}) {
			t.Errorf("tasks requires %v", a.Requires)
		}
	}

	if err := c.Link("tasks", "research"); err != nil {
		t.Fatalf("a duplicate link errored: %v", err)
	}
	for _, a := range c.Artifacts {
		if a.ID == "tasks" && len(a.Requires) != 1 {
			t.Errorf("a duplicate link added an edge: %v", a.Requires)
		}
	}

	if err := c.Link("tasks", "tasks"); !errors.Is(err, ErrSelfLink) {
		t.Errorf("error = %v, want ErrSelfLink", err)
	}
	if err := c.Link("tasks", "absent"); !errors.Is(err, ErrNotOnCanvas) {
		t.Errorf("error = %v, want ErrNotOnCanvas", err)
	}
	if err := c.Link("absent", "tasks"); !errors.Is(err, ErrNotOnCanvas) {
		t.Errorf("error = %v, want ErrNotOnCanvas", err)
	}
}

func TestUnlinking(t *testing.T) {
	t.Parallel()

	c, _, _ := chain(t)

	if err := c.Unlink("proposal", "research"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, a := range c.Artifacts {
		if a.ID == "proposal" && len(a.Requires) != 0 {
			t.Errorf("proposal still requires %v", a.Requires)
		}
	}

	if err := c.Unlink("proposal", "research"); err != nil {
		t.Errorf("unlinking twice errored: %v", err)
	}
	if err := c.Unlink("absent", "x"); !errors.Is(err, ErrNotOnCanvas) {
		t.Errorf("error = %v, want ErrNotOnCanvas", err)
	}
}

func TestGatesAndTracks(t *testing.T) {
	t.Parallel()

	c, _, _ := chain(t)

	if !c.IsGate("tasks") {
		t.Fatal("tasks does not gate apply")
	}

	if err := c.ToggleGate("tasks"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.IsGate("tasks") {
		t.Error("toggling did not remove the gate")
	}

	if err := c.ToggleGate("research"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !c.IsGate("research") {
		t.Error("toggling did not add the gate")
	}

	if err := c.ToggleGate("absent"); !errors.Is(err, ErrNotOnCanvas) {
		t.Errorf("error = %v, want ErrNotOnCanvas", err)
	}

	c.SetTracks("  notes.md  ")
	if c.Tracks != "notes.md" {
		t.Errorf("tracks = %q", c.Tracks)
	}
}

func TestRenaming(t *testing.T) {
	t.Parallel()

	c, _, _ := chain(t)

	if err := c.Rename("tasks", "work"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !c.Has("work") || c.Has("tasks") {
		t.Errorf("ids = %v", c.IDs())
	}
	if !reflect.DeepEqual(c.Gates, []string{"work"}) {
		t.Errorf("the gate was not renamed: %v", c.Gates)
	}

	if err := c.Rename("proposal", "plan"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, a := range c.Artifacts {
		if a.ID == "work" && !reflect.DeepEqual(a.Requires, []string{"plan"}) {
			t.Errorf("the requirement was not renamed: %v", a.Requires)
		}
	}
}

func TestRenamingRefusals(t *testing.T) {
	t.Parallel()

	c, _, _ := chain(t)

	if err := c.Rename("tasks", "tasks"); err != nil {
		t.Errorf("renaming to the same id errored: %v", err)
	}
	if err := c.Rename("tasks", "research"); !errors.Is(err, ErrCollision) {
		t.Errorf("error = %v, want ErrCollision", err)
	}
	if err := c.Rename("tasks", "   "); err == nil {
		t.Error("an empty id was accepted")
	}
	if err := c.Rename("absent", "x"); !errors.Is(err, ErrNotOnCanvas) {
		t.Errorf("error = %v, want ErrNotOnCanvas", err)
	}

	if got := c.IDs(); !reflect.DeepEqual(got, []string{"research", "proposal", "tasks"}) {
		t.Errorf("a refused rename changed the canvas: %v", got)
	}
}

func TestValidationFollowsTheEdits(t *testing.T) {
	t.Parallel()

	c, _, _ := chain(t)

	if !c.Valid("mix") {
		t.Fatalf("a well-formed composition is invalid: %v", c.Findings("mix"))
	}

	if err := c.Link("research", "tasks"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	findings := c.Findings("mix")
	if findings.Valid() {
		t.Fatal("a cycle was not reported")
	}

	var reported bool
	for _, f := range findings.Fatal() {
		if strings.Contains(f.Message, "cycle") {
			reported = true
			if !strings.Contains(f.Message, "research") || !strings.Contains(f.Message, "tasks") {
				t.Errorf("the cycle does not name its members: %s", f)
			}
		}
	}
	if !reported {
		t.Errorf("the findings do not mention a cycle: %v", findings)
	}

	if err := c.Unlink("research", "tasks"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !c.Valid("mix") {
		t.Errorf("resolving the cycle did not clear it: %v", c.Findings("mix"))
	}
}

func TestSeveralProblemsAreReportedAtOnce(t *testing.T) {
	t.Parallel()

	src, s := sourceDir(t, researchFirst, nil)

	c := New()
	c.AddSource(s, src.Dir, src.Ref)

	if err := c.Add(src, artifact(s, "research"), ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	collider := artifact(s, "proposal")
	collider.Generates = "research.md"
	if err := c.Add(src, collider, "also"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := c.Findings("mix")

	var text strings.Builder
	for _, f := range got {
		text.WriteString(f.String())
		text.WriteString("\n")
	}

	for _, want := range []string{"apply.requires names no artifact", "apply.tracks names no file", "generate"} {
		if !strings.Contains(text.String(), want) {
			t.Errorf("the findings do not mention %q:\n%s", want, text.String())
		}
	}
}

func TestTheComposerAndTheSchemaValidatorAgree(t *testing.T) {
	t.Parallel()

	c, _, _ := chain(t)

	fromComposition := c.Findings("mix")
	fromSchema := schema.Validate(c.Schema("mix"))

	if len(fromComposition) != len(fromSchema) {
		t.Errorf("the composer reported %d findings and the validator %d", len(fromComposition), len(fromSchema))
	}

	if err := c.Link("research", "tasks"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if c.Valid("mix") != schema.Validate(c.Schema("mix")).Valid() {
		t.Error("the composer and the validator disagree about validity")
	}
}

func TestTheBuiltSchemaCarriesEverything(t *testing.T) {
	t.Parallel()

	c, _, _ := chain(t)

	s := c.Schema("mix")

	if s.Name != "mix" || s.Version != 1 {
		t.Errorf("schema = %+v", s)
	}
	if len(s.Artifacts) != 3 {
		t.Fatalf("got %d artifacts, want 3", len(s.Artifacts))
	}
	if s.Artifacts[0].Description == "" {
		t.Error("the description was lost")
	}
	if !reflect.DeepEqual(s.Apply.Requires, []string{"tasks"}) {
		t.Errorf("apply.requires = %v", s.Apply.Requires)
	}
	if s.Apply.Tracks != "tasks.md" {
		t.Errorf("apply.tracks = %q", s.Apply.Tracks)
	}
}

func TestLinkingAnArtifactThatIsNotThere(t *testing.T) {
	t.Parallel()

	c, _, _ := chain(t)

	before := len(c.Artifacts)

	if err := c.Link("research", "absent"); !errors.Is(err, ErrNotOnCanvas) {
		t.Errorf("error = %v, want ErrNotOnCanvas", err)
	}
	if len(c.Artifacts) != before {
		t.Error("a failed link changed the canvas")
	}
}

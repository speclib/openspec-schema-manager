package compose

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/speclib/openspec-schema-manager/internal/schema"
)

func write(t *testing.T, path, body string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("creating %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
}

const researchFirst = `name: research-first
version: 1
description: Research before proposing
artifacts:
  - id: research
    generates: research.md
    description: What is already known
    template: research.md
  - id: proposal
    generates: proposal.md
    description: What to do about it
    template: proposal.md
    requires: [research]
  - id: tasks
    generates: tasks.md
    description: The work
    template: tasks.md
    requires: [proposal]
apply:
  requires: [tasks]
  tracks: tasks.md
`

const teamReview = `name: team-review
version: 1
description: A review gate before the specs
artifacts:
  - id: review
    generates: review.md
    description: What the team said
    template: review.md
  - id: tasks
    generates: tasks.md
    description: The work, again
    template: tasks.md
    requires: [review]
apply:
  requires: [tasks]
  tracks: tasks.md
`

func sourceDir(t *testing.T, body string, templates map[string]string) (Source, schema.Schema) {
	t.Helper()

	dir := t.TempDir()
	write(t, filepath.Join(dir, "schema.yaml"), body)

	s, err := schema.Parse("schema.yaml", []byte(body))
	if err != nil {
		t.Fatalf("parsing the fixture: %v", err)
	}

	for _, a := range s.Artifacts {
		text := templates[a.Template]
		if text == "" {
			text = "# " + a.ID + "\n"
		}
		write(t, filepath.Join(dir, "templates", filepath.FromSlash(a.Template)), text)
	}

	return Source{Dir: dir, Name: s.Name, Ref: "v1", Artifacts: s.Artifacts}, s
}

func artifact(s schema.Schema, id string) schema.Artifact {
	for _, a := range s.Artifacts {
		if a.ID == id {
			return a
		}
	}
	return schema.Artifact{}
}

func TestAddingAnArtifact(t *testing.T) {
	t.Parallel()

	src, s := sourceDir(t, researchFirst, nil)

	c := New()
	c.AddSource(s, src.Dir, src.Ref)

	if err := c.Add(src, artifact(s, "research"), ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(c.Artifacts) != 1 {
		t.Fatalf("got %d artifacts, want 1", len(c.Artifacts))
	}

	a := c.Artifacts[0]
	if a.ID != "research" || a.Generates != "research.md" || a.Template != "research.md" {
		t.Errorf("the artifact lost what it declared: %+v", a)
	}
	if a.SourceName != "research-first" || a.SourceRef != "v1" || a.SourceDir != src.Dir {
		t.Errorf("the artifact does not remember its source: %+v", a)
	}
	if a.SourceID != "research" || a.Renamed() {
		t.Errorf("an artifact added under its own id reports itself as renamed")
	}
}

func TestAnArtifactArrivesWithoutUnresolvedRequirements(t *testing.T) {
	t.Parallel()

	src, s := sourceDir(t, researchFirst, nil)

	c := New()
	c.AddSource(s, src.Dir, src.Ref)

	if err := c.Add(src, artifact(s, "tasks"), ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(c.Artifacts[0].Requires) != 0 {
		t.Errorf("tasks arrived requiring %v, none of which is on the canvas", c.Artifacts[0].Requires)
	}
	if !c.Valid("mix") {
		findings := c.Findings("mix")
		for _, f := range findings.Fatal() {
			if strings.Contains(f.Message, "no artifact declares") {
				t.Errorf("adding one artifact made the composition invalid: %s", f)
			}
		}
	}
}

func TestAnArtifactKeepsRequirementsAlreadyOnTheCanvas(t *testing.T) {
	t.Parallel()

	src, s := sourceDir(t, researchFirst, nil)

	c := New()
	c.AddSource(s, src.Dir, src.Ref)

	if err := c.Add(src, artifact(s, "research"), ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := c.Add(src, artifact(s, "proposal"), ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := c.Artifacts[1].Requires; !reflect.DeepEqual(got, []string{"research"}) {
		t.Errorf("proposal requires %v, want [research]", got)
	}
}

func TestAddingARequirementLaterDoesNotRestoreAnEdge(t *testing.T) {
	t.Parallel()

	src, s := sourceDir(t, researchFirst, nil)

	c := New()
	c.AddSource(s, src.Dir, src.Ref)

	if err := c.Add(src, artifact(s, "proposal"), ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := c.Add(src, artifact(s, "research"), ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(c.Artifacts[0].Requires) != 0 {
		t.Errorf("an edge came back on its own: %v", c.Artifacts[0].Requires)
	}
}

func TestACollisionIsReportedRatherThanRenamed(t *testing.T) {
	t.Parallel()

	first, one := sourceDir(t, researchFirst, nil)
	second, two := sourceDir(t, teamReview, nil)

	c := New()
	c.AddSource(one, first.Dir, first.Ref)
	c.AddSource(two, second.Dir, second.Ref)

	if err := c.Add(first, artifact(one, "tasks"), ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err := c.Add(second, artifact(two, "tasks"), "")
	if !errors.Is(err, ErrCollision) {
		t.Fatalf("error = %v, want ErrCollision", err)
	}
	if len(c.Artifacts) != 1 {
		t.Errorf("the colliding artifact was added anyway: %v", c.IDs())
	}

	if err := c.Add(second, artifact(two, "tasks"), "team-tasks"); err != nil {
		t.Fatalf("adding under a free id failed: %v", err)
	}

	added := c.Artifacts[1]
	if added.ID != "team-tasks" {
		t.Errorf("id = %q", added.ID)
	}
	if added.SourceID != "tasks" {
		t.Errorf("SourceID = %q, want the id it had in its source", added.SourceID)
	}
	if !added.Renamed() {
		t.Error("an artifact added under another id does not report itself as renamed")
	}

	if err := c.Add(second, artifact(two, "review"), "team-tasks"); !errors.Is(err, ErrCollision) {
		t.Errorf("a second collision was not refused: %v", err)
	}
}

func TestASourceIsRecordedOnce(t *testing.T) {
	t.Parallel()

	src, s := sourceDir(t, researchFirst, nil)

	c := New()
	c.AddSource(s, src.Dir, src.Ref)
	c.AddSource(s, src.Dir, src.Ref)

	if len(c.Sources) != 1 {
		t.Errorf("got %d sources, want 1", len(c.Sources))
	}
}

func TestFromSchema(t *testing.T) {
	t.Parallel()

	src, s := sourceDir(t, researchFirst, nil)

	c := FromSchema(s, src.Dir, "v1")

	if got := c.IDs(); !reflect.DeepEqual(got, []string{"research", "proposal", "tasks"}) {
		t.Errorf("ids = %v", got)
	}
	if !reflect.DeepEqual(c.Gates, []string{"tasks"}) {
		t.Errorf("gates = %v", c.Gates)
	}
	if c.Tracks != "tasks.md" {
		t.Errorf("tracks = %q", c.Tracks)
	}
	if got := c.Artifacts[2].Requires; !reflect.DeepEqual(got, []string{"proposal"}) {
		t.Errorf("tasks requires %v", got)
	}
	if !c.Valid("research-first") {
		t.Errorf("a composition from a valid schema is invalid: %v", c.Findings("research-first"))
	}
}

func TestAnEmptyCompositionReportsWhatItNeeds(t *testing.T) {
	t.Parallel()

	findings := New().Findings("mix")

	if findings.Valid() {
		t.Fatal("an empty composition is valid")
	}
	if !strings.Contains(findings.Fatal()[0].Message, "canvas is empty") {
		t.Errorf("the finding does not say the canvas is empty: %s", findings.Fatal()[0])
	}
}

package compose

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScanWarnsAboutAnArtifactNotOnTheCanvas(t *testing.T) {
	t.Parallel()

	src, s := sourceDir(t, researchFirst, map[string]string{
		"tasks.md": "# Tasks\n\nWork through what the proposal asked for.\n",
	})

	c := New()
	c.AddSource(s, src.Dir, src.Ref)

	if err := c.Add(src, artifact(s, "tasks"), ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	warnings := c.Scan()
	if len(warnings) == 0 {
		t.Fatal("a template mentioning a missing artifact produced no warning")
	}

	var found bool
	for _, w := range warnings {
		if w.Reference == "proposal" {
			found = true
			if w.Artifact != "tasks" {
				t.Errorf("the warning names %q as the artifact", w.Artifact)
			}
			if w.File != "tasks.md" {
				t.Errorf("the warning names %q as the file", w.File)
			}
			if !strings.Contains(w.String(), "not on the canvas") {
				t.Errorf("the warning reads %q", w.String())
			}
		}
	}
	if !found {
		t.Errorf("the reference was not reported: %+v", warnings)
	}
}

func TestScanWarnsAboutAGeneratedFile(t *testing.T) {
	t.Parallel()

	src, s := sourceDir(t, researchFirst, map[string]string{
		"tasks.md": "# Tasks\n\nCheck research.md before starting.\n",
	})

	c := New()
	c.AddSource(s, src.Dir, src.Ref)

	if err := c.Add(src, artifact(s, "tasks"), ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var found bool
	for _, w := range c.Scan() {
		if w.Reference == "research.md" {
			found = true
		}
	}
	if !found {
		t.Errorf("a generated file that nothing produces was not reported: %+v", c.Scan())
	}
}

func TestScanSaysNothingWhenTheReferenceResolves(t *testing.T) {
	t.Parallel()

	src, s := sourceDir(t, researchFirst, map[string]string{
		"tasks.md": "# Tasks\n\nWork through what the proposal asked for.\n",
	})

	c := New()
	c.AddSource(s, src.Dir, src.Ref)

	for _, id := range []string{"research", "proposal", "tasks"} {
		if err := c.Add(src, artifact(s, id), ""); err != nil {
			t.Fatalf("adding %s: %v", id, err)
		}
	}

	if got := c.Scan(); len(got) != 0 {
		t.Errorf("a complete canvas produced warnings: %+v", got)
	}
}

func TestAddingTheMissingArtifactClearsTheWarning(t *testing.T) {
	t.Parallel()

	src, s := sourceDir(t, researchFirst, map[string]string{
		"tasks.md": "# Tasks\n\nWork through what the proposal asked for.\n",
	})

	c := New()
	c.AddSource(s, src.Dir, src.Ref)

	if err := c.Add(src, artifact(s, "tasks"), ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(c.Scan()) == 0 {
		t.Fatal("no warning to clear")
	}

	if err := c.Add(src, artifact(s, "proposal"), ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, w := range c.Scan() {
		if w.Reference == "proposal" {
			t.Errorf("the warning survived adding the artifact: %+v", w)
		}
	}
}

func TestScanMatchesOnWordBoundaries(t *testing.T) {
	t.Parallel()

	src, s := sourceDir(t, researchFirst, map[string]string{
		"tasks.md": "# Tasks\n\nAsk the proposaller about subtasks.md and researching.\n",
	})

	c := New()
	c.AddSource(s, src.Dir, src.Ref)

	if err := c.Add(src, artifact(s, "tasks"), ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, w := range c.Scan() {
		t.Errorf("a substring inside a longer word was reported: %+v", w)
	}
}

func TestScanReadsTheInstructionToo(t *testing.T) {
	t.Parallel()

	src, s := sourceDir(t, researchFirst, nil)

	c := New()
	c.AddSource(s, src.Dir, src.Ref)

	a := artifact(s, "tasks")
	a.Instruction = "Read the proposal first."

	if err := c.Add(src, a, ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var found bool
	for _, w := range c.Scan() {
		if w.Reference == "proposal" {
			found = true
		}
	}
	if !found {
		t.Errorf("the instruction was not scanned: %+v", c.Scan())
	}
}

func TestATemplateThatCannotBeReadIsReportedOnce(t *testing.T) {
	t.Parallel()

	src, s := sourceDir(t, researchFirst, nil)

	c := New()
	c.AddSource(s, src.Dir, src.Ref)

	if err := c.Add(src, artifact(s, "tasks"), ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := os.Remove(filepath.Join(src.Dir, "templates", "tasks.md")); err != nil {
		t.Fatalf("removing the template: %v", err)
	}

	var reads int
	for _, w := range c.Scan() {
		if strings.Contains(w.Message, "could not be read") {
			reads++
		}
	}
	if reads != 1 {
		t.Errorf("a missing template was reported %d times", reads)
	}
}

func TestScanWithNoSourcesSaysNothing(t *testing.T) {
	t.Parallel()

	if got := New().Scan(); len(got) != 0 {
		t.Errorf("an empty composition produced warnings: %+v", got)
	}
}

func TestAnArtifactWithNoTemplateIsNotRead(t *testing.T) {
	t.Parallel()

	src, s := sourceDir(t, researchFirst, nil)

	c := New()
	c.AddSource(s, src.Dir, src.Ref)

	a := artifact(s, "tasks")
	a.Template = ""

	if err := c.Add(src, a, ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, w := range c.Scan() {
		if strings.Contains(w.Message, "could not be read") {
			t.Errorf("an artifact with no template was read anyway: %+v", w)
		}
	}
}

func TestAGlobIsNotScannedFor(t *testing.T) {
	t.Parallel()

	body := `name: globby
version: 1
description: a schema whose artifact generates a glob
artifacts:
  - id: specs
    generates: specs/**/*.md
    description: the specifications
    template: spec.md
  - id: tasks
    generates: tasks.md
    description: the work
    template: tasks.md
apply:
  requires: [tasks]
  tracks: tasks.md
`

	src, s := sourceDir(t, body, map[string]string{
		"tasks.md": "# Tasks\n\nWrite into specs/**/*.md as you go.\n",
	})

	c := New()
	c.AddSource(s, src.Dir, src.Ref)

	if err := c.Add(src, artifact(s, "tasks"), ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, w := range c.Scan() {
		if strings.Contains(w.Reference, "*") {
			t.Errorf("a glob was scanned for: %+v", w)
		}
	}
}

package compose

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/speclib/openspec-schema-manager/internal/schema"
	"github.com/speclib/openspec-schema-manager/internal/source"
)

func mixed(t *testing.T) *Composition {
	t.Helper()

	first, one := sourceDir(t, researchFirst, nil)
	second, two := sourceDir(t, teamReview, nil)

	c := New()
	c.AddSource(one, first.Dir, first.Ref)
	c.AddSource(two, second.Dir, second.Ref)

	if err := c.Add(first, artifact(one, "research"), ""); err != nil {
		t.Fatalf("adding research: %v", err)
	}
	if err := c.Add(second, artifact(two, "review"), ""); err != nil {
		t.Fatalf("adding review: %v", err)
	}
	if err := c.Add(first, artifact(one, "tasks"), "work"); err != nil {
		t.Fatalf("adding tasks: %v", err)
	}

	if err := c.Link("review", "research"); err != nil {
		t.Fatalf("linking: %v", err)
	}
	if err := c.Link("work", "review"); err != nil {
		t.Fatalf("linking: %v", err)
	}
	if err := c.ToggleGate("work"); err != nil {
		t.Fatalf("gating: %v", err)
	}
	c.SetTracks("tasks.md")

	return c
}

func TestWriteProducesASchemaThatValidates(t *testing.T) {
	t.Parallel()

	c := mixed(t)
	into := filepath.Join(t.TempDir(), "mine")

	destination, err := c.Write(into, "my-mix")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	loaded, err := schema.Load(destination)
	if err != nil {
		t.Fatalf("the written schema does not parse: %v", err)
	}

	if loaded.Name != "my-mix" {
		t.Errorf("name = %q", loaded.Name)
	}
	if got := loaded.IDs(); !reflect.DeepEqual(got, []string{"research", "review", "work"}) {
		t.Errorf("ids = %v", got)
	}
	if !reflect.DeepEqual(loaded.Apply.Requires, []string{"work"}) {
		t.Errorf("apply.requires = %v", loaded.Apply.Requires)
	}
	if loaded.Apply.Tracks != "tasks.md" {
		t.Errorf("apply.tracks = %q", loaded.Apply.Tracks)
	}

	findings := schema.ValidateDir(loaded, destination)
	if !findings.Valid() {
		t.Errorf("the written schema does not validate:\n%v", findings.Fatal())
	}
}

func TestWriteRecordsWhereEachArtifactCameFrom(t *testing.T) {
	t.Parallel()

	c := mixed(t)

	destination, err := c.Write(filepath.Join(t.TempDir(), "mine"), "my-mix")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(destination, "schema.yaml"))
	if err != nil {
		t.Fatalf("reading the written schema: %v", err)
	}
	got := string(raw)

	if !strings.HasPrefix(got, "# Composed with ossm from:") {
		t.Errorf("the file does not open with the provenance block:\n%s", got)
	}
	for _, want := range []string{"research-first @ v1", "team-review @ v1", "tasks as work"} {
		if !strings.Contains(got, want) {
			t.Errorf("the provenance block does not carry %q:\n%s", want, got)
		}
	}
}

func TestTemplatesAreCopiedByteForByte(t *testing.T) {
	t.Parallel()

	src, s := sourceDir(t, researchFirst, map[string]string{
		"research.md": "# Research\n\nExactly these bytes, trailing space and all.   \n",
	})

	c := FromSchema(s, src.Dir, "v1")

	destination, err := c.Write(filepath.Join(t.TempDir(), "mine"), "copy")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want, err := os.ReadFile(filepath.Join(src.Dir, "templates", "research.md"))
	if err != nil {
		t.Fatalf("reading the source template: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(destination, "templates", "research.md"))
	if err != nil {
		t.Fatalf("reading the copied template: %v", err)
	}

	if string(got) != string(want) {
		t.Errorf("the template changed:\n%q\nwant\n%q", got, want)
	}
}

func TestTwoSourcesSharingATemplatePath(t *testing.T) {
	t.Parallel()

	first, one := sourceDir(t, researchFirst, map[string]string{"tasks.md": "# from research-first\n"})
	second, two := sourceDir(t, teamReview, map[string]string{"tasks.md": "# from team-review\n"})

	c := New()
	c.AddSource(one, first.Dir, first.Ref)
	c.AddSource(two, second.Dir, second.Ref)

	if err := c.Add(first, artifact(one, "tasks"), ""); err != nil {
		t.Fatalf("adding: %v", err)
	}
	if err := c.Add(second, artifact(two, "tasks"), "team-tasks"); err != nil {
		t.Fatalf("adding: %v", err)
	}

	c.Artifacts[1].Generates = "team-tasks.md"

	if err := c.ToggleGate("tasks"); err != nil {
		t.Fatalf("gating: %v", err)
	}
	if err := c.ToggleGate("team-tasks"); err != nil {
		t.Fatalf("gating: %v", err)
	}
	c.SetTracks("tasks.md")

	destination, err := c.Write(filepath.Join(t.TempDir(), "mine"), "both")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	loaded, err := schema.Load(destination)
	if err != nil {
		t.Fatalf("the written schema does not parse: %v", err)
	}

	if loaded.Artifacts[0].Template == loaded.Artifacts[1].Template {
		t.Fatalf("both artifacts declare the same template %q", loaded.Artifacts[0].Template)
	}

	for i, want := range []string{"# from research-first\n", "# from team-review\n"} {
		path := filepath.Join(destination, "templates", filepath.FromSlash(loaded.Artifacts[i].Template))
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("reading %s: %v", path, err)
		}
		if string(raw) != want {
			t.Errorf("%s holds %q, want %q", path, raw, want)
		}
	}

	if findings := schema.ValidateDir(loaded, destination); !findings.Valid() {
		t.Errorf("the written schema does not validate:\n%v", findings.Fatal())
	}
}

func TestWriteRefusals(t *testing.T) {
	t.Parallel()

	into := filepath.Join(t.TempDir(), "mine")

	t.Run("a bad name", func(t *testing.T) {
		t.Parallel()

		for _, name := range []string{"", "..", "a/b"} {
			if _, err := mixed(t).Write(into, name); err == nil {
				t.Errorf("Write accepted the name %q", name)
			}
		}
	})

	t.Run("an invalid composition", func(t *testing.T) {
		t.Parallel()

		c := mixed(t)
		if err := c.Link("research", "work"); err != nil {
			t.Fatalf("linking: %v", err)
		}

		_, err := c.Write(filepath.Join(t.TempDir(), "mine"), "cyclic")
		if !errors.Is(err, ErrNotValid) {
			t.Fatalf("error = %v, want ErrNotValid", err)
		}
		if !strings.Contains(err.Error(), "cycle") {
			t.Errorf("error %q does not name the problem", err)
		}
	})

	t.Run("an empty composition", func(t *testing.T) {
		t.Parallel()

		if _, err := New().Write(filepath.Join(t.TempDir(), "mine"), "empty"); !errors.Is(err, ErrNotValid) {
			t.Errorf("error = %v, want ErrNotValid", err)
		}
	})

	t.Run("an occupied destination", func(t *testing.T) {
		t.Parallel()

		dir := filepath.Join(t.TempDir(), "mine")
		c := mixed(t)

		if _, err := c.Write(dir, "my-mix"); err != nil {
			t.Fatalf("the first write failed: %v", err)
		}
		if _, err := c.Write(dir, "my-mix"); !errors.Is(err, source.ErrDestinationBusy) {
			t.Errorf("error = %v, want ErrDestinationBusy", err)
		}
	})

	t.Run("a template that has gone", func(t *testing.T) {
		t.Parallel()

		c := mixed(t)

		if err := os.Remove(filepath.Join(c.Artifacts[0].SourceDir, "templates", "research.md")); err != nil {
			t.Fatalf("removing the template: %v", err)
		}

		dir := filepath.Join(t.TempDir(), "mine")
		if _, err := c.Write(dir, "my-mix"); err == nil {
			t.Fatal("a missing template was written over")
		}
		if _, err := os.Stat(filepath.Join(dir, "my-mix")); err == nil {
			t.Error("a failed write left a destination behind")
		}
	})
}

func TestWriteLeavesNoStagingDirectory(t *testing.T) {
	t.Parallel()

	into := filepath.Join(t.TempDir(), "mine")

	if _, err := mixed(t).Write(into, "my-mix"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries, err := os.ReadDir(into)
	if err != nil {
		t.Fatalf("reading %s: %v", into, err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".compose-") {
			t.Errorf("a write left %s behind", entry.Name())
		}
	}
}

func TestWriteReportsAnUnwritableDirectory(t *testing.T) {
	t.Parallel()

	if os.Geteuid() == 0 {
		t.Skip("running as root, so an unwritable directory is still writable")
	}

	root := t.TempDir()
	blocked := filepath.Join(root, "blocked")
	if err := os.Mkdir(blocked, 0o500); err != nil {
		t.Fatalf("creating the blocked directory: %v", err)
	}

	if _, err := mixed(t).Write(filepath.Join(blocked, "mine"), "my-mix"); err == nil {
		t.Fatal("expected an error writing into an unwritable directory")
	}
}

func TestAnInstructionSurvivesTheWrite(t *testing.T) {
	t.Parallel()

	src, s := sourceDir(t, researchFirst, nil)

	c := FromSchema(s, src.Dir, "v1")
	c.Artifacts[0].Instruction = "Line one.\nLine two.\n"

	destination, err := c.Write(filepath.Join(t.TempDir(), "mine"), "instructed")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	loaded, err := schema.Load(destination)
	if err != nil {
		t.Fatalf("the written schema does not parse: %v", err)
	}

	if !strings.Contains(loaded.Artifacts[0].Instruction, "Line one.") {
		t.Errorf("the instruction was lost: %q", loaded.Artifacts[0].Instruction)
	}
	if !strings.Contains(loaded.Artifacts[0].Instruction, "Line two.") {
		t.Errorf("the instruction was truncated: %q", loaded.Artifacts[0].Instruction)
	}
}

func TestQuotingAwkwardValues(t *testing.T) {
	t.Parallel()

	src, s := sourceDir(t, researchFirst, nil)

	c := FromSchema(s, src.Dir, "v1")
	c.Artifacts[0].Description = "Research: what is known, and what is not"

	destination, err := c.Write(filepath.Join(t.TempDir(), "mine"), "quoted")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	loaded, err := schema.Load(destination)
	if err != nil {
		t.Fatalf("a description holding a colon broke the written schema: %v", err)
	}
	if loaded.Artifacts[0].Description != "Research: what is known, and what is not" {
		t.Errorf("description = %q", loaded.Artifacts[0].Description)
	}
}

func TestQuote(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"plain":          "plain",
		"":               `""`,
		"holds: a colon": `"holds: a colon"`,
		`holds "quotes"`: `"holds \"quotes\""`,
		"  padded  ":     `"  padded  "`,
		"specs/**/*.md":  `"specs/**/*.md"`,
		"has-hyphens_ok": "has-hyphens_ok",
	}

	for in, want := range cases {
		if got := quote(in); got != want {
			t.Errorf("quote(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestProvenanceSkipsASourceNothingCameFrom(t *testing.T) {
	t.Parallel()

	first, one := sourceDir(t, researchFirst, nil)
	second, two := sourceDir(t, teamReview, nil)

	c := New()
	c.AddSource(one, first.Dir, first.Ref)
	c.AddSource(two, second.Dir, second.Ref)

	if err := c.Add(first, artifact(one, "research"), ""); err != nil {
		t.Fatalf("adding: %v", err)
	}

	got := c.provenance()
	if !strings.Contains(got, "research-first") {
		t.Errorf("the source that contributed is missing:\n%s", got)
	}
	if strings.Contains(got, "team-review") {
		t.Errorf("a source nothing came from is listed:\n%s", got)
	}
}

func TestProvenanceWithNoRef(t *testing.T) {
	t.Parallel()

	src, s := sourceDir(t, researchFirst, nil)

	c := FromSchema(s, src.Dir, "")

	if got := c.provenance(); !strings.Contains(got, "default branch") {
		t.Errorf("an absent ref is not described:\n%s", got)
	}
}

package source

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const richSchema = `# a schema with things a YAML round trip would lose
name: minimalist
version: 1
description: Lightweight schema for well-scoped, low-risk changes
artifacts:
  - id: specs
    generates: specs/**/*.md
    description: Specifications first
    template: specs/spec.md
    instruction: |
      Write the specs before anything else.
  - id: tasks
    generates: tasks.md
    description: Implementation checklist
    template: tasks.md
    requires: [specs]
apply:
  requires: [tasks]
  tracks: tasks.md
`

func richFolder(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	write(t, filepath.Join(dir, "schema.yaml"), richSchema)
	write(t, filepath.Join(dir, "templates", "tasks.md"), "# tasks\n")
	write(t, filepath.Join(dir, "templates", "specs", "spec.md"), "# spec\n")

	return dir
}

func TestValidName(t *testing.T) {
	t.Parallel()

	cases := map[string]error{
		"team-review": nil,
		"Team Review": nil,
		"":            ErrEmptyName,
		"   ":         ErrEmptyName,
		".":           ErrNameIsRelative,
		"..":          ErrNameIsRelative,
		"a/b":         ErrNameHasSeparator,
		`a\b`:         ErrNameHasSeparator,
		"../escape":   ErrNameHasSeparator,
	}

	for name, want := range cases {
		got := ValidName(name)
		if want == nil && got != nil {
			t.Errorf("ValidName(%q) = %v, want nil", name, got)
		}
		if want != nil && !errors.Is(got, want) {
			t.Errorf("ValidName(%q) = %v, want %v", name, got, want)
		}
	}
}

func TestDuplicateCopiesAndRenames(t *testing.T) {
	t.Parallel()

	from := richFolder(t)
	into := filepath.Join(t.TempDir(), "my-schemas")

	destination, err := Duplicate(from, into, "team-review")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if destination != filepath.Join(into, "team-review") {
		t.Errorf("destination = %q", destination)
	}

	raw, err := os.ReadFile(filepath.Join(destination, "schema.yaml"))
	if err != nil {
		t.Fatalf("reading the copy: %v", err)
	}
	got := string(raw)

	if !strings.Contains(got, "name: team-review") {
		t.Errorf("the name was not rewritten:\n%s", got)
	}
	if strings.Contains(got, "name: minimalist") {
		t.Errorf("the old name survived:\n%s", got)
	}
	for _, want := range []string{
		"# a schema with things a YAML round trip would lose",
		"Write the specs before anything else.",
		"description: Lightweight schema",
		"generates: specs/**/*.md",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("the copy lost %q:\n%s", want, got)
		}
	}

	if strings.Index(got, "- id: specs") > strings.Index(got, "- id: tasks") {
		t.Errorf("the artifact order changed:\n%s", got)
	}

	for _, want := range []string{filepath.Join("templates", "tasks.md"), filepath.Join("templates", "specs", "spec.md")} {
		if _, err := os.Stat(filepath.Join(destination, want)); err != nil {
			t.Errorf("%s is missing: %v", want, err)
		}
	}
}

func TestDuplicateMakesTheCopyWritable(t *testing.T) {
	t.Parallel()

	if os.Geteuid() == 0 {
		t.Skip("running as root, so a read-only file is still writable")
	}

	from := richFolder(t)
	for _, rel := range []string{"schema.yaml", filepath.Join("templates", "tasks.md")} {
		if err := os.Chmod(filepath.Join(from, rel), 0o444); err != nil {
			t.Fatalf("making %s read only: %v", rel, err)
		}
	}
	t.Cleanup(func() {
		_ = filepath.Walk(from, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				_ = os.Chmod(path, 0o644)
			}
			return nil
		})
	})

	destination, err := Duplicate(from, filepath.Join(t.TempDir(), "mine"), "team-review")
	if err != nil {
		t.Fatalf("duplicating a read-only schema failed: %v", err)
	}

	for _, rel := range []string{"schema.yaml", filepath.Join("templates", "tasks.md")} {
		info, err := os.Stat(filepath.Join(destination, rel))
		if err != nil {
			t.Fatalf("stat %s: %v", rel, err)
		}
		if info.Mode().Perm()&0o200 == 0 {
			t.Errorf("%s is not writable in the copy: %v", rel, info.Mode())
		}
	}
}

func TestDuplicateRefusesAnOccupiedDestination(t *testing.T) {
	t.Parallel()

	from := richFolder(t)
	into := filepath.Join(t.TempDir(), "mine")

	if _, err := Duplicate(from, into, "team-review"); err != nil {
		t.Fatalf("the first duplicate failed: %v", err)
	}

	marker := filepath.Join(into, "team-review", "leftover.md")
	if err := os.WriteFile(marker, []byte("old\n"), 0o644); err != nil {
		t.Fatalf("writing the marker: %v", err)
	}

	_, err := Duplicate(from, into, "team-review")
	if !errors.Is(err, ErrDestinationBusy) {
		t.Fatalf("error = %v, want ErrDestinationBusy", err)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Error("the refused duplicate removed what was there")
	}
}

func TestDuplicateRefusesABadName(t *testing.T) {
	t.Parallel()

	from := richFolder(t)
	into := filepath.Join(t.TempDir(), "mine")

	for _, name := range []string{"", "  ", ".", "..", "a/b"} {
		if _, err := Duplicate(from, into, name); err == nil {
			t.Errorf("Duplicate accepted the name %q", name)
		}
	}

	if _, err := os.Stat(into); err == nil {
		t.Error("a refused name created the destination directory")
	}
}

func TestDuplicateRefusesASourceWithNoSchema(t *testing.T) {
	t.Parallel()

	if _, err := Duplicate(t.TempDir(), filepath.Join(t.TempDir(), "mine"), "x"); err == nil {
		t.Fatal("a folder with no schema was duplicated")
	}
}

func TestDuplicateAddsANameWhenThereIsNone(t *testing.T) {
	t.Parallel()

	from := t.TempDir()
	write(t, filepath.Join(from, "schema.yaml"), "version: 1\nartifacts: []\napply:\n  requires: []\n")

	destination, err := Duplicate(from, filepath.Join(t.TempDir(), "mine"), "named")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(destination, "schema.yaml"))
	if err != nil {
		t.Fatalf("reading the copy: %v", err)
	}
	if !strings.HasPrefix(string(raw), "name: named\n") {
		t.Errorf("the name was not added:\n%s", raw)
	}
}

func TestDuplicateLeavesNoStagingDirectory(t *testing.T) {
	t.Parallel()

	into := filepath.Join(t.TempDir(), "mine")

	if _, err := Duplicate(richFolder(t), into, "team-review"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries, err := os.ReadDir(into)
	if err != nil {
		t.Fatalf("reading %s: %v", into, err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".duplicate-") {
			t.Errorf("a duplicate left %s behind", entry.Name())
		}
	}
}

func TestDuplicateReportsAnUnwritableDestination(t *testing.T) {
	t.Parallel()

	if os.Geteuid() == 0 {
		t.Skip("running as root, so an unwritable directory is still writable")
	}

	root := t.TempDir()
	blocked := filepath.Join(root, "blocked")
	if err := os.Mkdir(blocked, 0o500); err != nil {
		t.Fatalf("creating the blocked directory: %v", err)
	}

	if _, err := Duplicate(richFolder(t), filepath.Join(blocked, "mine"), "x"); err == nil {
		t.Fatal("expected an error duplicating into an unwritable directory")
	}
}

func TestFilesListsTheSchemaAndItsTemplates(t *testing.T) {
	t.Parallel()

	dir := richFolder(t)
	write(t, filepath.Join(dir, "README.md"), "# not part of the schema\n")

	got, err := Files(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var rels []string
	for _, f := range got {
		rels = append(rels, f.Rel)
	}

	want := []string{"schema.yaml", "templates/specs/spec.md", "templates/tasks.md"}
	if !reflect.DeepEqual(rels, want) {
		t.Errorf("files = %v, want %v", rels, want)
	}

	if got[0].Depth != 0 {
		t.Errorf("schema.yaml is at depth %d", got[0].Depth)
	}
	if got[1].Depth != 2 {
		t.Errorf("a nested template is at depth %d, want 2", got[1].Depth)
	}
	if got[0].Path != filepath.Join(dir, "schema.yaml") {
		t.Errorf("path = %q", got[0].Path)
	}
}

func TestFilesOnASchemaWithNoTemplates(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	write(t, filepath.Join(dir, "schema.yaml"), richSchema)

	got, err := Files(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].Rel != "schema.yaml" {
		t.Errorf("files = %v", got)
	}
}

func TestFilesIsStable(t *testing.T) {
	t.Parallel()

	dir := richFolder(t)

	first, err := Files(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for range 10 {
		again, err := Files(dir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(first, again) {
			t.Fatalf("the order changed between runs:\n%v\n%v", first, again)
		}
	}
}

func TestFilesReportsADirectoryThatIsNotThere(t *testing.T) {
	t.Parallel()

	if _, err := Files(filepath.Join(t.TempDir(), "absent")); err == nil {
		t.Fatal("expected an error for a directory that does not exist")
	}
}

package source

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func schemaFolder(t *testing.T, parent, folder, name string) string {
	t.Helper()

	dir := filepath.Join(parent, folder)

	write(t, filepath.Join(dir, "schema.yaml"), "name: "+name+`
version: 1
description: a schema
artifacts:
  - id: tasks
    generates: tasks.md
    description: the tasks
    template: tasks.md
apply:
  requires: [tasks]
  tracks: tasks.md
`)
	write(t, filepath.Join(dir, "templates", "tasks.md"), "# tasks\n")

	return dir
}

func TestScanDirFindsSchemaFolders(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	schemaFolder(t, root, "team-review", "team-review")
	schemaFolder(t, root, "minimalist", "minimalist")
	write(t, filepath.Join(root, "notes", "readme.md"), "# not a schema\n")
	write(t, filepath.Join(root, "loose.md"), "# not a directory\n")

	scan := ScanDir(root)

	if scan.Problem != "" {
		t.Errorf("problem = %q", scan.Problem)
	}
	if len(scan.Schemas) != 2 {
		t.Fatalf("got %d schemas, want 2", len(scan.Schemas))
	}

	var names []string
	for _, s := range scan.Schemas {
		names = append(names, s.Name)
	}
	if !reflect.DeepEqual(names, []string{"minimalist", "team-review"}) {
		t.Errorf("names = %v, want them sorted", names)
	}

	if scan.Schemas[0].Artifacts != 1 {
		t.Errorf("artifact count = %d", scan.Schemas[0].Artifacts)
	}
	if scan.Schemas[0].Directory != root {
		t.Errorf("directory = %q", scan.Schemas[0].Directory)
	}
}

func TestScanDirDoesNotDescend(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	schemaFolder(t, filepath.Join(root, "deep", "deeper"), "buried", "buried")

	if scan := ScanDir(root); len(scan.Schemas) != 0 {
		t.Errorf("scanning descended past one level: %v", scan.Schemas)
	}
}

func TestScanDirReportsADirectoryThatIsNotThere(t *testing.T) {
	t.Parallel()

	scan := ScanDir(filepath.Join(t.TempDir(), "absent"))

	if scan.Problem == "" {
		t.Error("a missing directory produced no problem")
	}
	if !strings.Contains(scan.Problem, "absent") {
		t.Errorf("problem %q does not name the directory", scan.Problem)
	}
	if len(scan.Schemas) != 0 {
		t.Errorf("schemas = %v", scan.Schemas)
	}
}

func TestScanDirOnADirectoryHoldingNoSchemas(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "empty"), 0o755); err != nil {
		t.Fatalf("creating the directory: %v", err)
	}

	scan := ScanDir(root)

	if scan.Problem != "" {
		t.Errorf("problem = %q", scan.Problem)
	}
	if len(scan.Schemas) != 0 {
		t.Errorf("schemas = %v", scan.Schemas)
	}
	if scan.Directory != root {
		t.Errorf("directory = %q", scan.Directory)
	}
}

func TestAnUnreadableSchemaIsListed(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	write(t, filepath.Join(root, "broken", "schema.yaml"), "name: [unclosed\n")

	scan := ScanDir(root)

	if len(scan.Schemas) != 1 {
		t.Fatalf("got %d schemas, want the unreadable one listed", len(scan.Schemas))
	}

	broken := scan.Schemas[0]
	if broken.Readable() {
		t.Error("an unparseable schema reports itself as readable")
	}
	if broken.Unreadable == "" {
		t.Error("no reason was given")
	}
	if broken.Label() != "broken" {
		t.Errorf("label = %q, want the folder name", broken.Label())
	}
}

func TestASchemaWithNoNameFallsBackToItsFolder(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	write(t, filepath.Join(root, "nameless", "schema.yaml"), "version: 1\nartifacts: []\napply:\n  requires: []\n")

	scan := ScanDir(root)

	if len(scan.Schemas) != 1 {
		t.Fatalf("got %d schemas, want 1", len(scan.Schemas))
	}
	if got := scan.Schemas[0].Label(); got != "nameless" {
		t.Errorf("label = %q, want the folder name", got)
	}
}

func TestReadOnAFolderWithNoSchema(t *testing.T) {
	t.Parallel()

	if _, ok := Read(t.TempDir(), ""); ok {
		t.Error("a folder with no schema was read as one")
	}
}

func TestScanAll(t *testing.T) {
	t.Parallel()

	first := t.TempDir()
	second := t.TempDir()
	schemaFolder(t, first, "a", "a")
	schemaFolder(t, second, "b", "b")

	scans := ScanAll([]string{first, second, filepath.Join(t.TempDir(), "absent")})

	if len(scans) != 3 {
		t.Fatalf("got %d scans, want 3", len(scans))
	}
	if len(scans[0].Schemas) != 1 || len(scans[1].Schemas) != 1 {
		t.Errorf("the configured directories were not both scanned")
	}
	if scans[2].Problem == "" {
		t.Error("the missing directory was not reported")
	}
}

package openspec

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func schemaSource(t *testing.T, name string) string {
	t.Helper()

	dir := t.TempDir()

	files := map[string]string{
		"schema.yaml":             "name: " + name + "\nversion: 1\ndescription: a schema\nartifacts:\n  - id: tasks\n    generates: tasks.md\n    description: the tasks\n    template: tasks.md\napply:\n  requires: [tasks]\n  tracks: tasks.md\n",
		"templates/tasks.md":      "# tasks\n",
		"templates/specs/spec.md": "# spec\n",
	}

	for rel, body := range files {
		path := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("creating %s: %v", filepath.Dir(path), err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatalf("writing %s: %v", path, err)
		}
	}

	return dir
}

func TestDestinationFollowsTheDeclaredName(t *testing.T) {
	t.Parallel()

	want := filepath.Join("/p", "openspec", "schemas", "SuperSpec")
	if got := Destination("/p", "SuperSpec"); got != want {
		t.Errorf("Destination = %q, want %q", got, want)
	}
}

func TestPlanInstall(t *testing.T) {
	t.Parallel()

	root := project(t, "schema: spec-driven\n")
	src := schemaSource(t, "minimalist")

	plan, err := PlanInstall(root, src, "minimalist")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if plan.Destination != filepath.Join(root, "openspec", "schemas", "minimalist") {
		t.Errorf("destination = %q", plan.Destination)
	}
	if plan.Occupied {
		t.Error("an empty destination is reported as occupied")
	}

	want := []string{"schema.yaml", "templates/specs/spec.md", "templates/tasks.md"}
	if !reflect.DeepEqual(plan.Files, want) {
		t.Errorf("files = %v, want %v", plan.Files, want)
	}
}

func TestPlanInstallForANameThatIsNotThePath(t *testing.T) {
	t.Parallel()

	root := project(t, "")
	src := schemaSource(t, "SuperSpec")

	plan, err := PlanInstall(root, src, "SuperSpec")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasSuffix(plan.Destination, filepath.Join("schemas", "SuperSpec")) {
		t.Errorf("destination = %q; it must follow the declared name", plan.Destination)
	}
}

func TestPlanInstallRefusesWhatItCannotPlan(t *testing.T) {
	t.Parallel()

	root := project(t, "")

	if _, err := PlanInstall(root, schemaSource(t, "x"), "  "); !errors.Is(err, ErrNoSchemaName) {
		t.Errorf("error = %v, want ErrNoSchemaName", err)
	}
	if _, err := PlanInstall(root, t.TempDir(), "x"); !errors.Is(err, ErrEmptySource) {
		t.Errorf("error = %v, want ErrEmptySource", err)
	}
	if _, err := PlanInstall(root, filepath.Join(t.TempDir(), "absent"), "x"); err == nil {
		t.Error("expected an error for a source that does not exist")
	}
}

func TestInstallWritesTheSchema(t *testing.T) {
	t.Parallel()

	root := project(t, "schema: spec-driven\n")
	src := schemaSource(t, "minimalist")

	plan, err := PlanInstall(root, src, "minimalist")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := Install(plan, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, want := range []string{"schema.yaml", filepath.Join("templates", "tasks.md"), filepath.Join("templates", "specs", "spec.md")} {
		if _, err := os.Stat(filepath.Join(plan.Destination, want)); err != nil {
			t.Errorf("%s is missing: %v", want, err)
		}
	}
}

func TestInstallRefusesAnOccupiedDestination(t *testing.T) {
	t.Parallel()

	root := project(t, "")
	src := schemaSource(t, "minimalist")

	plan, _ := PlanInstall(root, src, "minimalist")
	if err := Install(plan, false); err != nil {
		t.Fatalf("the first install failed: %v", err)
	}

	again, err := PlanInstall(root, src, "minimalist")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !again.Occupied {
		t.Fatal("an occupied destination is not reported as occupied")
	}
	if len(again.Existing) == 0 {
		t.Error("the plan does not say what is already there")
	}

	if err := Install(again, false); !errors.Is(err, ErrAlreadyThere) {
		t.Errorf("error = %v, want ErrAlreadyThere", err)
	}
}

func TestOverwritingReplacesWhatWasThere(t *testing.T) {
	t.Parallel()

	root := project(t, "")

	first := schemaSource(t, "minimalist")
	plan, _ := PlanInstall(root, first, "minimalist")
	if err := Install(plan, false); err != nil {
		t.Fatalf("the first install failed: %v", err)
	}

	if err := os.WriteFile(filepath.Join(plan.Destination, "leftover.md"), []byte("old\n"), 0o644); err != nil {
		t.Fatalf("writing the leftover: %v", err)
	}

	second := schemaSource(t, "minimalist")
	if err := os.WriteFile(filepath.Join(second, "templates", "new.md"), []byte("new\n"), 0o644); err != nil {
		t.Fatalf("writing the new template: %v", err)
	}

	again, _ := PlanInstall(root, second, "minimalist")
	if err := Install(again, true); err != nil {
		t.Fatalf("overwriting failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(plan.Destination, "leftover.md")); err == nil {
		t.Error("the old directory survived the overwrite")
	}
	if _, err := os.Stat(filepath.Join(plan.Destination, "templates", "new.md")); err != nil {
		t.Errorf("the new file is missing: %v", err)
	}
}

func TestAFailedInstallLeavesTheProjectAsItWas(t *testing.T) {
	t.Parallel()

	if os.Geteuid() == 0 {
		t.Skip("running as root, so an unwritable directory is still writable")
	}

	root := project(t, "")
	src := schemaSource(t, "minimalist")

	schemas := filepath.Join(root, "openspec", "schemas")
	if err := os.MkdirAll(schemas, 0o755); err != nil {
		t.Fatalf("creating %s: %v", schemas, err)
	}

	plan, err := PlanInstall(root, src, "minimalist")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := os.Chmod(schemas, 0o500); err != nil {
		t.Fatalf("making %s read only: %v", schemas, err)
	}
	t.Cleanup(func() { _ = os.Chmod(schemas, 0o755) })

	if err := Install(plan, false); err == nil {
		t.Fatal("expected an error installing into a read-only directory")
	}

	if _, err := os.Stat(plan.Destination); err == nil {
		t.Error("a failed install left a destination behind")
	}
}

func TestInstallLeavesNoStagingDirectory(t *testing.T) {
	t.Parallel()

	root := project(t, "")
	plan, _ := PlanInstall(root, schemaSource(t, "minimalist"), "minimalist")

	if err := Install(plan, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries, err := os.ReadDir(filepath.Dir(plan.Destination))
	if err != nil {
		t.Fatalf("reading the schemas directory: %v", err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".install-") {
			t.Errorf("an install left %s behind", entry.Name())
		}
	}
}

func TestSetDefaultSchemaKeepsEverythingElse(t *testing.T) {
	t.Parallel()

	config := `schema: spec-driven

# Project context (optional)
# This is shown to AI when creating artifacts.
context: |
  a project
`

	root := project(t, config)

	if err := SetDefaultSchema(root, "minimalist"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(root, "openspec", "config.yaml"))
	if err != nil {
		t.Fatalf("reading the config: %v", err)
	}
	got := string(raw)

	if !strings.Contains(got, "schema: minimalist") {
		t.Errorf("the default was not set:\n%s", got)
	}
	if strings.Contains(got, "schema: spec-driven") {
		t.Errorf("the old default survived:\n%s", got)
	}
	if !strings.Contains(got, "# Project context (optional)") {
		t.Errorf("the comments were lost:\n%s", got)
	}
	if !strings.Contains(got, "context: |") {
		t.Errorf("another key was lost:\n%s", got)
	}

	if got := DefaultSchema(root); got != "minimalist" {
		t.Errorf("the rewritten config reads back as %q", got)
	}
}

func TestSetDefaultSchemaAddsTheKeyWhenItIsMissing(t *testing.T) {
	t.Parallel()

	root := project(t, "context: |\n  a project\n")

	if err := SetDefaultSchema(root, "minimalist"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := DefaultSchema(root); got != "minimalist" {
		t.Errorf("DefaultSchema = %q", got)
	}
}

func TestSetDefaultSchemaReportsAMissingConfig(t *testing.T) {
	t.Parallel()

	if err := SetDefaultSchema(t.TempDir(), "minimalist"); err == nil {
		t.Fatal("expected an error for a project with no config")
	}
}

func TestInstallingDoesNotTouchTheConfig(t *testing.T) {
	t.Parallel()

	config := "schema: spec-driven\n\n# a comment\n"
	root := project(t, config)

	plan, _ := PlanInstall(root, schemaSource(t, "minimalist"), "minimalist")
	if err := Install(plan, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(root, "openspec", "config.yaml"))
	if err != nil {
		t.Fatalf("reading the config: %v", err)
	}
	if string(raw) != config {
		t.Errorf("an install changed the config:\n%s", raw)
	}
}

func TestValidateInstalled(t *testing.T) {
	t.Parallel()

	fake := &Fake{Validations: map[string]Validation{
		"broken": {Name: "broken", Valid: false, Issues: []ValidationIssue{{Message: "template missing"}}},
	}}

	good, err := ValidateInstalled(context.Background(), fake, "/p", "minimalist")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !good.Valid {
		t.Error("a valid schema was reported as invalid")
	}

	bad, err := ValidateInstalled(context.Background(), fake, "/p", "broken")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bad.Valid {
		t.Error("an invalid schema was reported as valid")
	}
	if len(bad.Issues) != 1 {
		t.Errorf("issues = %v", bad.Issues)
	}

	if _, err := ValidateInstalled(context.Background(), nil, "/p", "x"); err == nil {
		t.Error("a missing adapter was not reported")
	}
}

func TestInstallReportsASourceThatVanished(t *testing.T) {
	t.Parallel()

	root := project(t, "schema: spec-driven\n")
	src := schemaSource(t, "minimalist")

	plan, err := PlanInstall(root, src, "minimalist")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := os.RemoveAll(src); err != nil {
		t.Fatalf("removing the source: %v", err)
	}

	if err := Install(plan, false); err == nil {
		t.Fatal("installing from a source that no longer exists succeeded")
	}
	if _, err := os.Stat(plan.Destination); err == nil {
		t.Error("a failed install left a destination behind")
	}
}

func TestPlanInstallRefusesAFileWhereADirectoryIsExpected(t *testing.T) {
	t.Parallel()

	root := project(t, "")

	file := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(file, []byte("x\n"), 0o644); err != nil {
		t.Fatalf("writing the file: %v", err)
	}

	if _, err := PlanInstall(root, file, "x"); err == nil {
		t.Fatal("a file was accepted as a schema source")
	}
}

func TestSetDefaultSchemaReportsAnUnwritableConfig(t *testing.T) {
	t.Parallel()

	if os.Geteuid() == 0 {
		t.Skip("running as root, so a read-only file is still writable")
	}

	root := project(t, "schema: spec-driven\n")
	path := filepath.Join(root, "openspec", "config.yaml")

	if err := os.Chmod(path, 0o400); err != nil {
		t.Fatalf("making the config read only: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o644) })

	if err := SetDefaultSchema(root, "minimalist"); err == nil {
		t.Fatal("expected an error writing a read-only config")
	}
}

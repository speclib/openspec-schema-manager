package openspec

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func project(t *testing.T, config string) string {
	t.Helper()

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "openspec", "changes"), 0o755); err != nil {
		t.Fatalf("creating the project: %v", err)
	}
	if config != "" {
		if err := os.WriteFile(filepath.Join(root, "openspec", "config.yaml"), []byte(config), 0o644); err != nil {
			t.Fatalf("writing the config: %v", err)
		}
	}

	return root
}

func TestDefaultSchema(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		config string
		want   string
	}{
		{name: "a schema is named", config: "schema: minimalist\n", want: "minimalist"},
		{name: "with surrounding content", config: "# a comment\nschema: spec-driven\n\ncontext: |\n  words\n", want: "spec-driven"},
		{name: "padded", config: "schema:   minimalist   \n", want: "minimalist"},
		{name: "no schema key", config: "context: |\n  words\n", want: UnknownSchema},
		{name: "empty schema key", config: "schema: \n", want: UnknownSchema},
		{name: "not yaml", config: "schema: [unclosed\n", want: UnknownSchema},
		{name: "no file", config: "", want: UnknownSchema},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := DefaultSchema(project(t, tc.config)); got != tc.want {
				t.Errorf("DefaultSchema = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDefaultKnown(t *testing.T) {
	t.Parallel()

	if !(Project{Default: "minimalist"}).DefaultKnown() {
		t.Error("a named default is reported as unknown")
	}
	if (Project{Default: UnknownSchema}).DefaultKnown() {
		t.Error("an unknown default is reported as known")
	}
	if (Project{}).DefaultKnown() {
		t.Error("an empty default is reported as known")
	}
}

func TestReadProject(t *testing.T) {
	t.Parallel()

	root := project(t, "schema: minimalist\n")

	fake := &Fake{
		SchemasResult: []ResolvedSchema{
			{Name: "minimalist", Source: SourceProject, Path: "/p/openspec/schemas/minimalist"},
			{Name: "spec-driven", Source: SourcePackage, Path: "/nix/store/x"},
		},
		ChangesResult: []Change{
			{Name: "add-auth", CompletedTasks: 3, TotalTasks: 7},
			{Name: "fix-export", CompletedTasks: 0, TotalTasks: 4},
		},
		ChangeSchemas: map[string]string{"add-auth": "spec-driven", "fix-export": "minimalist"},
	}

	got := ReadProject(context.Background(), fake, root)

	if got.Root != root {
		t.Errorf("root = %q", got.Root)
	}
	if got.Default != "minimalist" {
		t.Errorf("default = %q", got.Default)
	}
	if len(got.Schemas) != 2 {
		t.Errorf("got %d schemas, want 2", len(got.Schemas))
	}
	if len(got.Changes) != 2 {
		t.Fatalf("got %d changes, want 2", len(got.Changes))
	}
	if len(got.Problems) != 0 {
		t.Errorf("problems = %v", got.Problems)
	}

	byName := map[string]ProjectChange{}
	for _, c := range got.Changes {
		byName[c.Name] = c
	}

	if byName["add-auth"].Schema != "spec-driven" {
		t.Errorf("add-auth uses %q", byName["add-auth"].Schema)
	}
	if byName["fix-export"].Schema != "minimalist" {
		t.Errorf("fix-export uses %q", byName["fix-export"].Schema)
	}
	if byName["add-auth"].CompletedTasks != 3 || byName["add-auth"].TotalTasks != 7 {
		t.Errorf("add-auth progress = %d/%d", byName["add-auth"].CompletedTasks, byName["add-auth"].TotalTasks)
	}
}

func TestAChangeWhoseSchemaCannotBeReadIsStillListed(t *testing.T) {
	t.Parallel()

	fake := &Fake{
		ChangesResult:   []Change{{Name: "add-auth"}},
		ChangeSchemaErr: errors.New("exit status 1"),
	}

	got := ReadProject(context.Background(), fake, project(t, "schema: spec-driven\n"))

	if len(got.Changes) != 1 {
		t.Fatalf("got %d changes, want 1", len(got.Changes))
	}
	if got.Changes[0].Schema != UnknownSchema {
		t.Errorf("schema = %q, want %q", got.Changes[0].Schema, UnknownSchema)
	}
}

func TestAProjectWithNoChanges(t *testing.T) {
	t.Parallel()

	got := ReadProject(context.Background(), &Fake{}, project(t, "schema: spec-driven\n"))

	if len(got.Changes) != 0 {
		t.Errorf("got %d changes, want none", len(got.Changes))
	}
	if got.Default != "spec-driven" {
		t.Errorf("the default is not reported for a project with no changes: %q", got.Default)
	}
}

func TestAMissingCLIIsReportedNotFatal(t *testing.T) {
	t.Parallel()

	fake := &Fake{SchemasErr: ErrNotInstalled, ChangesErr: ErrNotInstalled}

	got := ReadProject(context.Background(), fake, project(t, "schema: spec-driven\n"))

	if len(got.Problems) != 2 {
		t.Fatalf("problems = %v, want two", got.Problems)
	}
	for _, p := range got.Problems {
		if !strings.Contains(p, "not on PATH") {
			t.Errorf("problem %q does not say the CLI is missing", p)
		}
	}
	if got.Default != "spec-driven" {
		t.Error("the default was lost because the CLI is missing")
	}
}

func TestAFailingCommandCarriesItsMessage(t *testing.T) {
	t.Parallel()

	fake := &Fake{ChangesErr: errors.New("exit status 3: something broke")}

	got := ReadProject(context.Background(), fake, project(t, ""))

	if len(got.Problems) != 1 {
		t.Fatalf("problems = %v, want one", got.Problems)
	}
	if !strings.Contains(got.Problems[0], "something broke") {
		t.Errorf("problem %q does not carry what the CLI said", got.Problems[0])
	}
}

func TestNoAdapterAtAllIsReported(t *testing.T) {
	t.Parallel()

	got := ReadProject(context.Background(), nil, project(t, "schema: spec-driven\n"))

	if len(got.Problems) != 1 || !strings.Contains(got.Problems[0], "no OpenSpec adapter") {
		t.Errorf("problems = %v", got.Problems)
	}
}

func TestResolvedSchemaLabels(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		SourcePackage: "built-in",
		SourceProject: "project",
		SourceUser:    "user",
		"something":   "something",
	}

	for source, want := range cases {
		if got := (ResolvedSchema{Source: source}).OriginLabel(); got != want {
			t.Errorf("OriginLabel(%q) = %q, want %q", source, got, want)
		}
	}

	if (ResolvedSchema{}).Shadowing() {
		t.Error("a schema shadowing nothing reports that it shadows")
	}
	if !(ResolvedSchema{Shadows: []string{"user"}}).Shadowing() {
		t.Error("a schema shadowing something does not report it")
	}
}

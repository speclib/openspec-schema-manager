package registry

import (
	"reflect"
	"testing"
)

func fixtureEntries(t *testing.T) []Entry {
	t.Helper()

	doc, err := Parse("fixture", fixture(t, "registry.json"))
	if err != nil {
		t.Fatalf("parsing the fixture: %v", err)
	}

	return doc.Schemas
}

func names(rows []Row) []string {
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		out = append(out, r.Name)
	}
	return out
}

func TestMergeSortsByNameCaseInsensitively(t *testing.T) {
	t.Parallel()

	rows := Merge(fixtureEntries(t), []Row{
		RowFromInstalled("spec-driven", "package", "/nix/store/openspec/schemas/spec-driven", []string{"proposal", "specs", "design", "tasks"}),
	})

	want := []string{"e2e-runbooks", "retired", "spec-driven", "SuperSpec", "tinychange"}
	if got := names(rows); !reflect.DeepEqual(got, want) {
		t.Errorf("order = %v, want %v", got, want)
	}
}

func TestMergeIsStable(t *testing.T) {
	t.Parallel()

	entries := fixtureEntries(t)
	installed := []Row{RowFromInstalled("spec-driven", "package", "/p", nil)}

	first := names(Merge(entries, installed))
	second := names(Merge(entries, installed))

	if !reflect.DeepEqual(first, second) {
		t.Errorf("the order changed between runs: %v then %v", first, second)
	}
}

func TestTwoRowsSharingANameKeepAStableOrder(t *testing.T) {
	t.Parallel()

	entries := []Entry{
		{ID: "b/shared", Name: "shared", Artifacts: []string{"x"}},
		{ID: "a/shared", Name: "shared", Artifacts: []string{"x"}},
	}
	installed := []Row{RowFromInstalled("shared", "package", "/p", nil)}

	rows := Merge(entries, installed)

	if len(rows) != 3 {
		t.Fatalf("got %d rows, want 3", len(rows))
	}
	if rows[0].Origin != OriginBuiltIn {
		t.Errorf("the built-in row is at position %d, want first among equals", 0)
	}
	if rows[1].Entry.ID != "a/shared" || rows[2].Entry.ID != "b/shared" {
		t.Errorf("registry rows sharing a name are ordered %q then %q", rows[1].Entry.ID, rows[2].Entry.ID)
	}
}

func TestARowCarriesWhatIsNeededToCompare(t *testing.T) {
	t.Parallel()

	rows := Merge(fixtureEntries(t), nil)

	byName := map[string]Row{}
	for _, r := range rows {
		byName[r.Name] = r
	}

	super := byName["SuperSpec"]
	if super.ArtifactCount() != 9 {
		t.Errorf("SuperSpec artifact count = %d, want 9", super.ArtifactCount())
	}
	if got := super.Shape(); got != "brainstorm → … → finalize" {
		t.Errorf("SuperSpec shape = %q", got)
	}
	if got := super.RefLabel(); got != "default branch" {
		t.Errorf("SuperSpec ref label = %q, want the default branch", got)
	}
	if !super.Installable {
		t.Error("a registry row is not installable")
	}

	runbooks := byName["e2e-runbooks"]
	if got := runbooks.RefLabel(); got != "v0.2.0" {
		t.Errorf("e2e-runbooks ref label = %q, want v0.2.0", got)
	}

	tiny := byName["tinychange"]
	if got := tiny.Shape(); got != "specs → tasks" {
		t.Errorf("tinychange shape = %q", got)
	}

	retired := byName["retired"]
	if !retired.Deprecated {
		t.Error("the deprecated row does not say so")
	}
	if retired.SupersededBy != "speclib/tinychange" {
		t.Errorf("the deprecated row names %q as its replacement", retired.SupersededBy)
	}
}

func TestABuiltInRowIsNotInstallableAndCarriesNoRef(t *testing.T) {
	t.Parallel()

	row := RowFromInstalled("spec-driven", "package", "/nix/store/x/spec-driven", []string{"proposal", "specs", "design", "tasks"})

	if row.Origin != OriginBuiltIn {
		t.Errorf("origin = %q, want %q", row.Origin, OriginBuiltIn)
	}
	if row.Installable {
		t.Error("a built-in schema is offered as installable")
	}
	if got := row.RefLabel(); got != "" {
		t.Errorf("a built-in row carries ref %q; there is no source to pin", got)
	}
	if row.Path == "" {
		t.Error("a built-in row does not say where it resolves from")
	}
}

func TestInstalledOriginsAreDistinguished(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"package": OriginBuiltIn,
		"project": OriginProject,
		"user":    OriginUser,
		"unknown": OriginBuiltIn,
	}

	for source, want := range cases {
		if got := RowFromInstalled("x", source, "/p", nil).Origin; got != want {
			t.Errorf("source %q gave origin %q, want %q", source, got, want)
		}
	}
}

func TestShapeOfShortAndEmptyWorkflows(t *testing.T) {
	t.Parallel()

	cases := []struct {
		artifacts []string
		want      string
	}{
		{artifacts: nil, want: ""},
		{artifacts: []string{"tasks"}, want: "tasks"},
		{artifacts: []string{"specs", "tasks"}, want: "specs → tasks"},
		{artifacts: []string{"a", "b", "c"}, want: "a → b → c"},
		{artifacts: []string{"a", "b", "c", "d"}, want: "a → … → d"},
	}

	for _, tc := range cases {
		if got := (Row{Artifacts: tc.artifacts}).Shape(); got != tc.want {
			t.Errorf("Shape(%v) = %q, want %q", tc.artifacts, got, tc.want)
		}
	}
}

func TestFilterMatchesNameDescriptionAndID(t *testing.T) {
	t.Parallel()

	rows := Merge(fixtureEntries(t), []Row{RowFromInstalled("spec-driven", "package", "/p", nil)})

	cases := []struct {
		name  string
		query string
		want  []string
	}{
		{name: "empty matches everything", query: "", want: []string{"e2e-runbooks", "retired", "spec-driven", "SuperSpec", "tinychange"}},
		{name: "whitespace matches everything", query: "   ", want: []string{"e2e-runbooks", "retired", "spec-driven", "SuperSpec", "tinychange"}},
		{name: "by name", query: "tinychange", want: []string{"tinychange"}},
		{name: "by name ignoring case", query: "SUPERSPEC", want: []string{"SuperSpec"}},
		{name: "by description", query: "worktrees", want: []string{"SuperSpec"}},
		{name: "by id owner", query: "lukk17", want: []string{"e2e-runbooks"}},
		{name: "matching several", query: "workflow", want: []string{"retired", "SuperSpec", "tinychange"}},
		{name: "matching nothing", query: "nonexistent", want: []string{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := names(Filter(rows, tc.query))
			if len(got) == 0 && len(tc.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Filter(%q) = %v, want %v", tc.query, got, tc.want)
			}
		})
	}
}

func TestFilterDoesNotMatchABuiltInRowByAnAbsentID(t *testing.T) {
	t.Parallel()

	rows := []Row{RowFromInstalled("spec-driven", "package", "/p", nil)}

	if got := Filter(rows, "speclib"); len(got) != 0 {
		t.Errorf("a built-in row matched an owner it does not have: %v", names(got))
	}
}

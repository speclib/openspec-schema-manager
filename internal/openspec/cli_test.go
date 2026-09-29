package openspec

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func stubBinary(t *testing.T, script string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "openspec")

	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+script), 0o755); err != nil {
		t.Fatalf("writing the stub: %v", err)
	}

	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))

	return path
}

func TestSchemasReadsTheCLIOutput(t *testing.T) {
	stubBinary(t, `
echo "Note: Schema commands are experimental and may change." >&2
cat <<'JSON'
[
  {"name":"spec-driven","source":"package","path":"/nix/store/x/spec-driven","shadows":[]},
  {"name":"minimalist","source":"project","path":"/home/t/demo/openspec/schemas/minimalist","shadows":["user"]}
]
JSON
`)

	got, err := NewExec().Schemas(t.Context(), t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d schemas, want 2", len(got))
	}

	if !got[0].BuiltIn() {
		t.Error("a package schema does not report itself as built in")
	}
	if got[1].BuiltIn() {
		t.Error("a project schema reports itself as built in")
	}
	if got[1].Source != SourceProject {
		t.Errorf("source = %q, want %q", got[1].Source, SourceProject)
	}
	if len(got[1].Shadows) != 1 {
		t.Errorf("shadows = %v", got[1].Shadows)
	}
}

func TestSchemasIgnoresTheExperimentalNoteOnStderr(t *testing.T) {
	stubBinary(t, `
echo "Note: Schema commands are experimental and may change." >&2
echo '[]'
`)

	got, err := NewExec().Schemas(t.Context(), t.TempDir())
	if err != nil {
		t.Fatalf("the note on stderr was treated as output: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %d schemas, want none", len(got))
	}
}

func TestSchemasReportsOutputItCannotParse(t *testing.T) {
	stubBinary(t, `echo 'not json'`)

	_, err := NewExec().Schemas(t.Context(), t.TempDir())
	if err == nil {
		t.Fatal("expected an error for output that is not JSON")
	}
	if !strings.Contains(err.Error(), "schema which") {
		t.Errorf("error %q does not name the command", err)
	}
}

func TestSchemasReportsAFailingCommand(t *testing.T) {
	stubBinary(t, `
echo "something went wrong" >&2
exit 3
`)

	_, err := NewExec().Schemas(t.Context(), t.TempDir())
	if err == nil {
		t.Fatal("expected an error for a failing command")
	}
	if !strings.Contains(err.Error(), "something went wrong") {
		t.Errorf("error %q does not carry what the CLI said", err)
	}
}

func TestSchemasReportsAMissingBinary(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	_, err := Exec{Binary: "definitely-not-openspec"}.Schemas(t.Context(), t.TempDir())
	if !errors.Is(err, ErrNotInstalled) {
		t.Fatalf("error = %v, want ErrNotInstalled", err)
	}
}

func TestExecDefaultsToTheOpenSpecBinary(t *testing.T) {
	t.Parallel()

	if got := (Exec{}).binary(); got != "openspec" {
		t.Errorf("binary = %q, want openspec", got)
	}
	if got := (Exec{Binary: "other"}).binary(); got != "other" {
		t.Errorf("binary = %q, want other", got)
	}
}

func TestSchemasHonoursItsTimeout(t *testing.T) {
	stubBinary(t, `sleep 5`)

	e := NewExec()
	e.Timeout = 100 * time.Millisecond

	if _, err := e.Schemas(t.Context(), t.TempDir()); err == nil {
		t.Fatal("expected a timeout")
	}
}

func TestSchemasRunsInTheGivenDirectory(t *testing.T) {
	stubBinary(t, `echo "[{\"name\":\"$(basename "$PWD")\",\"source\":\"project\",\"path\":\"$PWD\"}]"`)

	dir := filepath.Join(t.TempDir(), "demo-app")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("creating the directory: %v", err)
	}

	got, err := NewExec().Schemas(t.Context(), dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].Name != "demo-app" {
		t.Errorf("the command did not run in %s: %+v", dir, got)
	}
}

func TestTheFakeRecordsItsCalls(t *testing.T) {
	t.Parallel()

	fake := &Fake{SchemasResult: []ResolvedSchema{{Name: "spec-driven", Source: SourcePackage}}}

	var cli CLI = fake

	got, err := cli.Schemas(t.Context(), "/somewhere")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d schemas, want 1", len(got))
	}
	if len(fake.SchemasCalls) != 1 || fake.SchemasCalls[0] != "/somewhere" {
		t.Errorf("the fake recorded %v", fake.SchemasCalls)
	}
}

func TestTheFakeCanFail(t *testing.T) {
	t.Parallel()

	fake := &Fake{SchemasErr: ErrNotInstalled}

	if _, err := fake.Schemas(t.Context(), "/x"); !errors.Is(err, ErrNotInstalled) {
		t.Errorf("error = %v, want ErrNotInstalled", err)
	}
}

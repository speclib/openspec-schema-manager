package source

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixtureDir(t *testing.T, fixture string) string {
	t.Helper()

	dir := t.TempDir()

	raw, err := os.ReadFile(filepath.Join("..", "schema", "testdata", fixture))
	if err != nil {
		t.Fatalf("reading the fixture: %v", err)
	}
	write(t, filepath.Join(dir, "schema.yaml"), string(raw))

	for _, template := range []string{"specs/spec.md", "tasks.md", "proposal.md", "design.md", "spec.md", "sketch.md"} {
		write(t, filepath.Join(dir, "templates", filepath.FromSlash(template)), "# template\n")
	}

	return dir
}

type stubFetcher struct {
	dir     string
	err     error
	refetch bool
	calls   int
}

func (s *stubFetcher) Fetch(_ context.Context, _ Source, refetch bool) (string, error) {
	s.calls++
	s.refetch = refetch

	if s.err != nil {
		return "", s.err
	}

	return s.dir, nil
}

func TestResolveDir(t *testing.T) {
	t.Parallel()

	dir := fixtureDir(t, "chain.yaml")

	got, err := ResolveDir(dir, "minimalist")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Origin != OriginLocal {
		t.Errorf("origin = %q, want %q", got.Origin, OriginLocal)
	}
	if got.Dir != dir {
		t.Errorf("dir = %q, want %q", got.Dir, dir)
	}
	if got.Schema.Name != "minimalist" {
		t.Errorf("name = %q", got.Schema.Name)
	}
	if got.Graph.Len() != 2 {
		t.Errorf("the graph holds %d nodes, want 2", got.Graph.Len())
	}
	if got.Metrics.LongestChain != 2 {
		t.Errorf("longest chain = %d, want 2", got.Metrics.LongestChain)
	}
	if !got.Findings.Valid() {
		t.Errorf("a complete fixture does not validate: %v", got.Findings)
	}
	if got.Cyclic() {
		t.Error("a chain reports itself as cyclic")
	}
	if got.Label != "minimalist" {
		t.Errorf("label = %q", got.Label)
	}
}

func TestResolveDirReportsAMissingSchema(t *testing.T) {
	t.Parallel()

	if _, err := ResolveDir(t.TempDir(), "absent"); err == nil {
		t.Fatal("expected an error for a directory with no schema")
	}
}

func TestResolveFetchesFirst(t *testing.T) {
	t.Parallel()

	stub := &stubFetcher{dir: fixtureDir(t, "chain.yaml")}
	src := Source{Repo: "https://e.test/r", Path: "p", Ref: "v1"}

	got, err := Resolve(context.Background(), stub, src, "minimalist", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stub.calls != 1 {
		t.Errorf("the fetcher was called %d times", stub.calls)
	}
	if stub.refetch {
		t.Error("a first resolve asked for a refetch")
	}
	if got.Origin != OriginRegistry {
		t.Errorf("origin = %q, want %q", got.Origin, OriginRegistry)
	}
	if got.Source != src {
		t.Errorf("source = %+v, want %+v", got.Source, src)
	}
}

func TestResolveAsksForARefetch(t *testing.T) {
	t.Parallel()

	stub := &stubFetcher{dir: fixtureDir(t, "chain.yaml")}

	if _, err := Resolve(context.Background(), stub, Source{}, "x", true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !stub.refetch {
		t.Error("the refetch flag did not reach the fetcher")
	}
}

func TestResolveReportsAFailedFetch(t *testing.T) {
	t.Parallel()

	stub := &stubFetcher{err: ErrNoGit}

	_, err := Resolve(context.Background(), stub, Source{}, "x", false)
	if !errors.Is(err, ErrNoGit) {
		t.Fatalf("error = %v, want ErrNoGit", err)
	}
}

func TestResolveReportsASchemaThatWillNotLoad(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	write(t, filepath.Join(dir, "schema.yaml"), "name: [unclosed\n")

	stub := &stubFetcher{dir: dir}

	_, err := Resolve(context.Background(), stub, Source{}, "x", false)
	if err == nil {
		t.Fatal("expected an error for a schema that does not parse")
	}
	if !strings.Contains(err.Error(), "the fetched schema") {
		t.Errorf("error %q does not say what it was reading", err)
	}
}

func TestACyclicSchemaResolvesAndSaysSo(t *testing.T) {
	t.Parallel()

	got, err := ResolveDir(fixtureDir(t, "cycle.yaml"), "cycle")
	if err != nil {
		t.Fatalf("a cyclic schema did not resolve: %v", err)
	}

	if !got.Cyclic() {
		t.Error("a cyclic schema does not report itself as cyclic")
	}
	if got.Metrics.Artifacts != 3 {
		t.Errorf("a cyclic schema reported %d artifacts, want 3", got.Metrics.Artifacts)
	}
	if got.Metrics.LongestChain != 0 {
		t.Errorf("a cyclic schema reported a chain of %d; it has no order to measure", got.Metrics.LongestChain)
	}
	if got.Findings.Valid() {
		t.Error("a cyclic schema validates")
	}
}

func TestFindingsComeFromTheDirectory(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	raw, err := os.ReadFile(filepath.Join("..", "schema", "testdata", "chain.yaml"))
	if err != nil {
		t.Fatalf("reading the fixture: %v", err)
	}
	write(t, filepath.Join(dir, "schema.yaml"), string(raw))
	write(t, filepath.Join(dir, "templates", "tasks.md"), "# tasks\n")

	got, err := ResolveDir(dir, "minimalist")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Findings.Valid() {
		t.Error("a missing template file was not reported, so ValidateDir was not used")
	}
}

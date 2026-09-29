package source

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const chainSchema = `name: minimalist
version: 1
description: Lightweight schema for well-scoped, low-risk changes
artifacts:
  - id: specs
    generates: specs/**/*.md
    template: specs/spec.md
  - id: tasks
    generates: tasks.md
    template: tasks.md
    requires: [specs]
apply:
  requires: [tasks]
  tracks: tasks.md
`

func git(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_CONFIG_SYSTEM=/dev/null",
		"GIT_AUTHOR_NAME=Test", "GIT_AUTHOR_EMAIL=test@example.test",
		"GIT_COMMITTER_NAME=Test", "GIT_COMMITTER_EMAIL=test@example.test",
	)

	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

func write(t *testing.T, path, body string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("creating %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
}

type repoOptions struct {
	branch   string
	path     string
	schema   string
	tag      string
	extraDir string
}

func newRepo(t *testing.T, opts repoOptions) string {
	t.Helper()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not on PATH")
	}

	if opts.branch == "" {
		opts.branch = "main"
	}
	if opts.path == "" {
		opts.path = "openspec/schemas/minimalist"
	}
	if opts.schema == "" {
		opts.schema = chainSchema
	}

	dir := t.TempDir()

	git(t, dir, "init", "--quiet", "--initial-branch", opts.branch)

	write(t, filepath.Join(dir, filepath.FromSlash(opts.path), "schema.yaml"), opts.schema)
	write(t, filepath.Join(dir, filepath.FromSlash(opts.path), "templates", "tasks.md"), "# tasks\n")
	write(t, filepath.Join(dir, filepath.FromSlash(opts.path), "templates", "specs", "spec.md"), "# spec\n")
	write(t, filepath.Join(dir, "README.md"), "# a repository\n")

	if opts.extraDir != "" {
		write(t, filepath.Join(dir, filepath.FromSlash(opts.extraDir), "schema.yaml"), strings.Replace(opts.schema, "name: minimalist", "name: other", 1))
		write(t, filepath.Join(dir, filepath.FromSlash(opts.extraDir), "templates", "tasks.md"), "# other tasks\n")
		write(t, filepath.Join(dir, filepath.FromSlash(opts.extraDir), "templates", "specs", "spec.md"), "# other spec\n")
	}

	git(t, dir, "add", "-A")
	git(t, dir, "commit", "--quiet", "-m", "the schema")

	if opts.tag != "" {
		git(t, dir, "tag", opts.tag)
	}

	return dir
}

func newFetcher(t *testing.T) Git {
	t.Helper()

	g := NewGit(filepath.Join(t.TempDir(), "schemas"))
	g.Timeout = 60 * time.Second

	return g
}

func TestFetchReadsTheNamedPath(t *testing.T) {
	t.Parallel()

	repo := newRepo(t, repoOptions{})
	g := newFetcher(t)

	dir, err := g.Fetch(t.Context(), Source{Repo: repo, Path: "openspec/schemas/minimalist"}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, want := range []string{"schema.yaml", filepath.Join("templates", "tasks.md"), filepath.Join("templates", "specs", "spec.md")} {
		if _, err := os.Stat(filepath.Join(dir, want)); err != nil {
			t.Errorf("%s is missing from the fetched folder: %v", want, err)
		}
	}

	if _, err := os.Stat(filepath.Join(dir, "README.md")); err == nil {
		t.Error("the whole repository was copied, not just the named path")
	}
}

func TestThePathIsNotDerivedFromTheName(t *testing.T) {
	t.Parallel()

	repo := newRepo(t, repoOptions{path: "schemas/tucked-away"})
	g := newFetcher(t)

	dir, err := g.Fetch(t.Context(), Source{Repo: repo, Path: "schemas/tucked-away"}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "schema.yaml")); err != nil {
		t.Errorf("the schema was not found at the named path: %v", err)
	}
}

func TestADefaultBranchThatIsNotMain(t *testing.T) {
	t.Parallel()

	repo := newRepo(t, repoOptions{branch: "master"})
	g := newFetcher(t)

	if _, err := g.Fetch(t.Context(), Source{Repo: repo, Path: "openspec/schemas/minimalist"}, false); err != nil {
		t.Fatalf("a repository whose default branch is master could not be fetched: %v", err)
	}
}

func TestAnUnusualDefaultBranch(t *testing.T) {
	t.Parallel()

	repo := newRepo(t, repoOptions{branch: "trunk"})
	g := newFetcher(t)

	if _, err := g.Fetch(t.Context(), Source{Repo: repo, Path: "openspec/schemas/minimalist"}, false); err != nil {
		t.Fatalf("a repository whose default branch is trunk could not be fetched: %v", err)
	}
}

func TestANamedTag(t *testing.T) {
	t.Parallel()

	repo := newRepo(t, repoOptions{tag: "v0.2.0"})
	g := newFetcher(t)

	if _, err := g.Fetch(t.Context(), Source{Repo: repo, Path: "openspec/schemas/minimalist", Ref: "v0.2.0"}, false); err != nil {
		t.Fatalf("a pinned tag could not be fetched: %v", err)
	}
}

func TestARefThatDoesNotResolve(t *testing.T) {
	t.Parallel()

	repo := newRepo(t, repoOptions{})
	g := newFetcher(t)

	_, err := g.Fetch(t.Context(), Source{Repo: repo, Path: "openspec/schemas/minimalist", Ref: "v9.9.9"}, false)
	if err == nil {
		t.Fatal("expected an error for a ref that does not resolve")
	}
	if !strings.Contains(err.Error(), "git fetch") {
		t.Errorf("error %q does not say what failed", err)
	}
}

func TestAPathThatIsNotInTheRepository(t *testing.T) {
	t.Parallel()

	repo := newRepo(t, repoOptions{})
	g := newFetcher(t)

	_, err := g.Fetch(t.Context(), Source{Repo: repo, Path: "openspec/schemas/absent"}, false)
	if err == nil {
		t.Fatal("expected an error for a path that is not in the repository")
	}
	if !strings.Contains(err.Error(), "openspec/schemas/absent") {
		t.Errorf("error %q does not name the path", err)
	}
	if !strings.Contains(err.Error(), repo) {
		t.Errorf("error %q does not name the repository", err)
	}
}

func TestADirectoryHoldingNoSchema(t *testing.T) {
	t.Parallel()

	repo := newRepo(t, repoOptions{})
	g := newFetcher(t)

	_, err := g.Fetch(t.Context(), Source{Repo: repo, Path: "openspec/schemas/minimalist/templates"}, false)
	if !errors.Is(err, ErrNoSchema) {
		t.Fatalf("error = %v, want ErrNoSchema", err)
	}
}

func TestADirectoryHoldingAnUnparseableSchema(t *testing.T) {
	t.Parallel()

	repo := newRepo(t, repoOptions{schema: "name: [unclosed\n"})
	g := newFetcher(t)

	if _, err := g.Fetch(t.Context(), Source{Repo: repo, Path: "openspec/schemas/minimalist"}, false); err == nil {
		t.Fatal("expected an error for a schema that does not parse")
	}
}

func TestASecondFetchUsesTheCache(t *testing.T) {
	t.Parallel()

	repo := newRepo(t, repoOptions{})
	g := newFetcher(t)

	src := Source{Repo: repo, Path: "openspec/schemas/minimalist"}

	first, err := g.Fetch(t.Context(), src, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	broken := g
	broken.Binary = "definitely-not-git"

	second, err := broken.Fetch(t.Context(), src, false)
	if err != nil {
		t.Fatalf("a cached source was fetched again: %v", err)
	}
	if second != first {
		t.Errorf("the cached directory moved from %q to %q", first, second)
	}
}

func TestTwoSchemasFromOneRepository(t *testing.T) {
	t.Parallel()

	repo := newRepo(t, repoOptions{extraDir: "openspec/schemas/other"})
	g := newFetcher(t)

	one, err := g.Fetch(t.Context(), Source{Repo: repo, Path: "openspec/schemas/minimalist"}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	two, err := g.Fetch(t.Context(), Source{Repo: repo, Path: "openspec/schemas/other"}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if one == two {
		t.Fatal("two schemas from one repository share a cache directory")
	}

	for dir, want := range map[string]string{one: "name: minimalist", two: "name: other"} {
		raw, err := os.ReadFile(filepath.Join(dir, "schema.yaml"))
		if err != nil {
			t.Fatalf("reading %s: %v", dir, err)
		}
		if !strings.Contains(string(raw), want) {
			t.Errorf("%s does not hold %q", dir, want)
		}
	}
}

func TestOneSchemaAtTwoRefs(t *testing.T) {
	t.Parallel()

	repo := newRepo(t, repoOptions{tag: "v0.2.0"})
	g := newFetcher(t)

	tracking, err := g.Fetch(t.Context(), Source{Repo: repo, Path: "openspec/schemas/minimalist"}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pinned, err := g.Fetch(t.Context(), Source{Repo: repo, Path: "openspec/schemas/minimalist", Ref: "v0.2.0"}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tracking == pinned {
		t.Error("the same schema at two refs shares a cache directory")
	}
	if _, err := os.Stat(filepath.Join(tracking, "schema.yaml")); err != nil {
		t.Errorf("the first fetch was overwritten: %v", err)
	}
}

func TestARefetchReplacesTheCachedCopy(t *testing.T) {
	t.Parallel()

	repo := newRepo(t, repoOptions{})
	g := newFetcher(t)

	src := Source{Repo: repo, Path: "openspec/schemas/minimalist"}

	dir, err := g.Fetch(t.Context(), src, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	write(t, filepath.Join(repo, "openspec/schemas/minimalist/templates/extra.md"), "# extra\n")
	git(t, repo, "add", "-A")
	git(t, repo, "commit", "--quiet", "-m", "another template")

	again, err := g.Fetch(t.Context(), src, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if again != dir {
		t.Errorf("the refetch moved the cache directory from %q to %q", dir, again)
	}
	if _, err := os.Stat(filepath.Join(again, "templates", "extra.md")); err != nil {
		t.Errorf("the refetch did not bring the new file: %v", err)
	}
}

func TestAFailedFetchLeavesTheCachedCopyAlone(t *testing.T) {
	t.Parallel()

	repo := newRepo(t, repoOptions{})
	g := newFetcher(t)

	src := Source{Repo: repo, Path: "openspec/schemas/minimalist"}

	dir, err := g.Fetch(t.Context(), src, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	unreachable := src
	unreachable.Repo = filepath.Join(t.TempDir(), "not-a-repository")

	failing := Source{Repo: unreachable.Repo, Path: src.Path}
	if _, err := g.Fetch(t.Context(), failing, true); err == nil {
		t.Fatal("expected an error fetching from a directory that is not a repository")
	}

	if _, err := os.Stat(filepath.Join(dir, "schema.yaml")); err != nil {
		t.Errorf("a failed fetch of another source damaged the cache: %v", err)
	}

	if _, err := g.Fetch(t.Context(), src, true); err != nil {
		t.Errorf("refetching after a failure elsewhere failed: %v", err)
	}
}

func TestAFailedRefetchLeavesNothingPartial(t *testing.T) {
	t.Parallel()

	repo := newRepo(t, repoOptions{})
	g := newFetcher(t)

	src := Source{Repo: repo, Path: "openspec/schemas/minimalist"}
	if _, err := g.Fetch(t.Context(), src, false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	broken := src
	broken.Ref = "v9.9.9"
	if _, err := g.Fetch(t.Context(), broken, true); err == nil {
		t.Fatal("expected an error for a ref that does not resolve")
	}

	entries, err := os.ReadDir(g.CacheRoot)
	if err != nil {
		t.Fatalf("reading the cache root: %v", err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".fetch-") {
			t.Errorf("a failed fetch left %s behind", entry.Name())
		}
	}
}

func TestAMissingGitIsReportedAsSuch(t *testing.T) {
	t.Parallel()

	g := newFetcher(t)
	g.Binary = "definitely-not-git"

	_, err := g.Fetch(t.Context(), Source{Repo: "/nowhere", Path: "p"}, false)
	if !errors.Is(err, ErrNoGit) {
		t.Fatalf("error = %v, want ErrNoGit", err)
	}
}

func TestCachedReportsWhatIsThere(t *testing.T) {
	t.Parallel()

	repo := newRepo(t, repoOptions{})
	g := newFetcher(t)

	src := Source{Repo: repo, Path: "openspec/schemas/minimalist"}

	if _, err := g.Cached(src); !errors.Is(err, ErrNotCached) {
		t.Fatalf("error = %v, want ErrNotCached", err)
	}

	dir, err := g.Fetch(t.Context(), src, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cached, err := g.Cached(src)
	if err != nil {
		t.Fatalf("a fetched source is not reported as cached: %v", err)
	}
	if cached != dir {
		t.Errorf("Cached = %q, Fetch = %q", cached, dir)
	}
}

func TestTheSourceRecordIsWritten(t *testing.T) {
	t.Parallel()

	repo := newRepo(t, repoOptions{tag: "v0.2.0"})
	g := newFetcher(t)

	src := Source{Repo: repo, Path: "openspec/schemas/minimalist", Ref: "v0.2.0"}

	dir, err := g.Fetch(t.Context(), src, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := ReadSourceRecord(dir)
	if err != nil {
		t.Fatalf("reading the source record: %v", err)
	}
	if got != src {
		t.Errorf("the record holds %+v, want %+v", got, src)
	}
}

func TestReadSourceRecordReportsWhatItCannotRead(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	if _, err := ReadSourceRecord(dir); err == nil {
		t.Fatal("expected an error for a directory with no record")
	}

	write(t, filepath.Join(dir, "source.json"), "{oops")
	if _, err := ReadSourceRecord(dir); err == nil {
		t.Fatal("expected an error for a record that is not JSON")
	}
}

func TestTheKeyDependsOnAllThreeFields(t *testing.T) {
	t.Parallel()

	base := Source{Repo: "https://e.test/r", Path: "p", Ref: "v1"}

	seen := map[string]string{base.Key(): "base"}

	for name, changed := range map[string]Source{
		"a different repo": {Repo: "https://e.test/other", Path: "p", Ref: "v1"},
		"a different path": {Repo: "https://e.test/r", Path: "q", Ref: "v1"},
		"a different ref":  {Repo: "https://e.test/r", Path: "p", Ref: "v2"},
		"no ref":           {Repo: "https://e.test/r", Path: "p"},
	} {
		key := changed.Key()
		if other, clash := seen[key]; clash {
			t.Errorf("%s shares a key with %s", name, other)
		}
		seen[key] = name
	}

	if base.Key() != base.Key() {
		t.Error("the key is not stable")
	}
}

func TestRefLabel(t *testing.T) {
	t.Parallel()

	if got := (Source{}).RefLabel(); got != "default branch" {
		t.Errorf("RefLabel with no ref = %q", got)
	}
	if got := (Source{Ref: "v1"}).RefLabel(); got != "v1" {
		t.Errorf("RefLabel = %q", got)
	}
}

func TestGitDefaultsToTheGitBinary(t *testing.T) {
	t.Parallel()

	if got := (Git{}).binary(); got != "git" {
		t.Errorf("binary = %q", got)
	}
}

func TestAnUnwritableCacheRootIsReported(t *testing.T) {
	t.Parallel()

	if os.Geteuid() == 0 {
		t.Skip("running as root, so an unwritable directory is still writable")
	}

	root := t.TempDir()
	blocked := filepath.Join(root, "blocked")
	if err := os.Mkdir(blocked, 0o500); err != nil {
		t.Fatalf("creating the blocked directory: %v", err)
	}

	g := NewGit(filepath.Join(blocked, "schemas"))

	if _, err := g.Fetch(t.Context(), Source{Repo: newRepo(t, repoOptions{}), Path: "openspec/schemas/minimalist"}, false); err == nil {
		t.Fatal("expected an error for an unwritable cache root")
	}
}

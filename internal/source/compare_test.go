package source

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/speclib/openspec-schema-manager/internal/registry"
)

func entry(id, name string, ref string) registry.Entry {
	return registry.Entry{
		ID:          id,
		Name:        name,
		Description: "a schema",
		Artifacts:   []string{"tasks"},
		Source: registry.Source{
			Repo: "https://example.test/" + id,
			Path: "openspec/schemas/" + name,
			Ref:  ref,
		},
	}
}

func folder(t *testing.T, files map[string]string) string {
	t.Helper()

	dir := t.TempDir()
	for rel, body := range files {
		write(t, filepath.Join(dir, filepath.FromSlash(rel)), body)
	}

	return dir
}

func TestCompareIdentical(t *testing.T) {
	t.Parallel()

	files := map[string]string{
		"schema.yaml":        "name: minimalist\n",
		"templates/tasks.md": "# tasks\n",
	}

	installed := folder(t, files)
	fetched := folder(t, files)

	got := Compare(context.Background(), &stubFetcher{dir: fetched}, installed, false, "minimalist", []registry.Entry{entry("a/minimalist", "minimalist", "v1")})

	if got.Verdict != VerdictIdentical {
		t.Fatalf("verdict = %v, want identical", got.Verdict)
	}
	if got.Ref != "v1" {
		t.Errorf("ref = %q", got.Ref)
	}
	if explained := got.Explain("minimalist"); !strings.Contains(explained, "matches what the registry offers at v1") {
		t.Errorf("Explain = %q", explained)
	}
}

func TestCompareDiffers(t *testing.T) {
	t.Parallel()

	installed := folder(t, map[string]string{
		"schema.yaml":        "name: minimalist\n",
		"templates/tasks.md": "# tasks, edited\n",
		"templates/extra.md": "# only here\n",
	})
	fetched := folder(t, map[string]string{
		"schema.yaml":         "name: minimalist\n",
		"templates/tasks.md":  "# tasks\n",
		"templates/theirs.md": "# only there\n",
	})

	got := Compare(context.Background(), &stubFetcher{dir: fetched}, installed, false, "minimalist", []registry.Entry{entry("a/minimalist", "minimalist", "")})

	if got.Verdict != VerdictDiffers {
		t.Fatalf("verdict = %v, want differs", got.Verdict)
	}
	if !reflect.DeepEqual(got.Changed, []string{"templates/tasks.md"}) {
		t.Errorf("changed = %v", got.Changed)
	}
	if !reflect.DeepEqual(got.Added, []string{"templates/extra.md"}) {
		t.Errorf("added = %v", got.Added)
	}
	if !reflect.DeepEqual(got.Removed, []string{"templates/theirs.md"}) {
		t.Errorf("removed = %v", got.Removed)
	}
	if got.Ref != "default branch" {
		t.Errorf("ref = %q, want the default branch", got.Ref)
	}

	explained := got.Explain("minimalist")
	if strings.Contains(strings.ToLower(explained), "update") {
		t.Errorf("a difference was described as an update: %q", explained)
	}
	if !strings.Contains(explained, "cannot tell an upstream change from a local edit") {
		t.Errorf("the limitation is not stated: %q", explained)
	}
	if !strings.Contains(explained, "OpenSpec records nothing") {
		t.Errorf("the reason is not stated: %q", explained)
	}
}

func TestCompareIgnoresTheSourceRecord(t *testing.T) {
	t.Parallel()

	installed := folder(t, map[string]string{"schema.yaml": "name: x\n"})
	fetched := folder(t, map[string]string{
		"schema.yaml": "name: x\n",
		"source.json": `{"repo":"https://example.test/a","path":"p"}`,
	})

	got := Compare(context.Background(), &stubFetcher{dir: fetched}, installed, false, "x", []registry.Entry{entry("a/x", "x", "v1")})

	if got.Verdict != VerdictIdentical {
		t.Errorf("ossm's own cache note was counted as a difference: %+v", got)
	}
}

func TestCompareWithNothingInTheRegistry(t *testing.T) {
	t.Parallel()

	got := Compare(context.Background(), &stubFetcher{}, t.TempDir(), false, "nowhere", []registry.Entry{entry("a/other", "other", "")})

	if got.Verdict != VerdictNoEntry {
		t.Fatalf("verdict = %v, want no entry", got.Verdict)
	}
	if explained := got.Explain("nowhere"); !strings.Contains(explained, "nothing to compare against") {
		t.Errorf("Explain = %q", explained)
	}
}

func TestCompareWithSeveralCandidates(t *testing.T) {
	t.Parallel()

	entries := []registry.Entry{
		entry("b/shared", "shared", ""),
		entry("a/shared", "shared", ""),
	}

	got := Compare(context.Background(), &stubFetcher{}, t.TempDir(), false, "shared", entries)

	if got.Verdict != VerdictSeveralEntries {
		t.Fatalf("verdict = %v, want several entries", got.Verdict)
	}
	if !reflect.DeepEqual(got.Candidates, []string{"a/shared", "b/shared"}) {
		t.Errorf("candidates = %v", got.Candidates)
	}

	explained := got.Explain("shared")
	for _, want := range []string{"a/shared", "b/shared", "not unique across the registry"} {
		if !strings.Contains(explained, want) {
			t.Errorf("Explain does not carry %q: %q", want, explained)
		}
	}
}

func TestCompareWhenTheSourceCannotBeReached(t *testing.T) {
	t.Parallel()

	stub := &stubFetcher{err: errors.New("no route to host")}

	got := Compare(context.Background(), stub, t.TempDir(), false, "minimalist", []registry.Entry{entry("a/minimalist", "minimalist", "v1")})

	if got.Verdict != VerdictUnreachable {
		t.Fatalf("verdict = %v, want unreachable", got.Verdict)
	}
	if explained := got.Explain("minimalist"); !strings.Contains(explained, "no route to host") {
		t.Errorf("Explain = %q", explained)
	}
}

func TestCompareABuiltInSchema(t *testing.T) {
	t.Parallel()

	got := Compare(context.Background(), &stubFetcher{}, "/nix/store/x", true, "spec-driven", nil)

	if got.Verdict != VerdictBuiltIn {
		t.Fatalf("verdict = %v, want built in", got.Verdict)
	}
	if explained := got.Explain("spec-driven"); !strings.Contains(explained, "ships with OpenSpec") {
		t.Errorf("Explain = %q", explained)
	}
}

func TestCompareRefetches(t *testing.T) {
	t.Parallel()

	stub := &stubFetcher{dir: folder(t, map[string]string{"schema.yaml": "name: x\n"})}

	Compare(context.Background(), stub, folder(t, map[string]string{"schema.yaml": "name: x\n"}), false, "x", []registry.Entry{entry("a/x", "x", "")})

	if !stub.refetch {
		t.Error("the comparison used a cached copy rather than fetching fresh")
	}
}

func TestCompareReportsAnInstalledFolderItCannotRead(t *testing.T) {
	t.Parallel()

	got := Compare(context.Background(), &stubFetcher{dir: t.TempDir()}, filepath.Join(t.TempDir(), "absent"), false, "x", []registry.Entry{entry("a/x", "x", "")})

	if got.Verdict != VerdictUnreachable {
		t.Fatalf("verdict = %v, want unreachable", got.Verdict)
	}
}

func TestExplainOfAnUnknownVerdict(t *testing.T) {
	t.Parallel()

	if got := (Comparison{Verdict: Verdict(99)}).Explain("x"); got != "" {
		t.Errorf("Explain = %q, want empty", got)
	}
}

func TestReadTreeReportsWhatItCannotRead(t *testing.T) {
	t.Parallel()

	if _, err := readTree(filepath.Join(t.TempDir(), "absent")); err == nil {
		t.Fatal("reading a missing directory succeeded")
	}

	if os.Geteuid() == 0 {
		t.Skip("running as root, so an unreadable file is still readable")
	}

	dir := folder(t, map[string]string{"schema.yaml": "name: x\n"})
	blocked := filepath.Join(dir, "blocked.md")
	if err := os.WriteFile(blocked, []byte("x\n"), 0o000); err != nil {
		t.Fatalf("writing the blocked file: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(blocked, 0o644) })

	if _, err := readTree(dir); err == nil {
		t.Fatal("reading an unreadable file succeeded")
	}
}

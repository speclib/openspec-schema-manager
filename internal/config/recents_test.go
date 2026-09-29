package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func newRecents(t *testing.T, cap int) Recents {
	t.Helper()

	return Recents{Path: filepath.Join(t.TempDir(), "state", "recents.json"), Cap: cap}
}

func TestRecentsStartEmpty(t *testing.T) {
	t.Parallel()

	if got := newRecents(t, 20).Read(); len(got) != 0 {
		t.Errorf("a missing recents file read as %v", got)
	}
}

func TestAddingPutsAPathOnTop(t *testing.T) {
	t.Parallel()

	r := newRecents(t, 20)

	if _, err := r.Add("/a"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := r.Add("/b")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(got, []string{"/b", "/a"}) {
		t.Errorf("recents = %v, want [/b /a]", got)
	}
	if !reflect.DeepEqual(r.Read(), []string{"/b", "/a"}) {
		t.Errorf("the file reads back as %v", r.Read())
	}
}

func TestAddingTheSamePathAgainMovesItUp(t *testing.T) {
	t.Parallel()

	r := newRecents(t, 20)

	for _, path := range []string{"/a", "/b", "/c"} {
		if _, err := r.Add(path); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	got, err := r.Add("/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(got, []string{"/a", "/c", "/b"}) {
		t.Errorf("recents = %v, want [/a /c /b]", got)
	}

	var seen int
	for _, path := range got {
		if path == "/a" {
			seen++
		}
	}
	if seen != 1 {
		t.Errorf("/a appears %d times", seen)
	}
}

func TestTheCapIsHonoured(t *testing.T) {
	t.Parallel()

	r := newRecents(t, 3)

	for _, path := range []string{"/a", "/b", "/c", "/d", "/e"} {
		if _, err := r.Add(path); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	got := r.Read()
	if !reflect.DeepEqual(got, []string{"/e", "/d", "/c"}) {
		t.Errorf("recents = %v, want the three most recent", got)
	}
}

func TestACapOfZeroRemembersNothing(t *testing.T) {
	t.Parallel()

	r := newRecents(t, 0)

	got, err := r.Add("/a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("recents = %v, want none", got)
	}
	if len(r.Read()) != 0 {
		t.Errorf("the file reads back as %v", r.Read())
	}
}

func TestANegativeCapRemembersNothing(t *testing.T) {
	t.Parallel()

	r := newRecents(t, -5)

	if got, _ := r.Add("/a"); len(got) != 0 {
		t.Errorf("recents = %v, want none", got)
	}
}

func TestACorruptFileReadsAsEmptyAndStillWrites(t *testing.T) {
	t.Parallel()

	r := newRecents(t, 20)

	if err := EnsureParent(r.Path); err != nil {
		t.Fatalf("creating the state directory: %v", err)
	}
	if err := os.WriteFile(r.Path, []byte("{oops"), 0o644); err != nil {
		t.Fatalf("writing the corrupt file: %v", err)
	}

	if got := r.Read(); len(got) != 0 {
		t.Errorf("a corrupt file read as %v", got)
	}

	got, err := r.Add("/a")
	if err != nil {
		t.Fatalf("a corrupt file stopped a write: %v", err)
	}
	if !reflect.DeepEqual(got, []string{"/a"}) {
		t.Errorf("recents = %v", got)
	}
}

func TestARelativePathIsStoredAbsolute(t *testing.T) {
	t.Parallel()

	r := newRecents(t, 20)

	got, err := r.Add("relative/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || !filepath.IsAbs(got[0]) {
		t.Errorf("recents = %v, want one absolute path", got)
	}
}

func TestWritingLeavesNoTemporaryFile(t *testing.T) {
	t.Parallel()

	r := newRecents(t, 20)

	if _, err := r.Add("/a"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	entries, err := os.ReadDir(filepath.Dir(r.Path))
	if err != nil {
		t.Fatalf("reading the state directory: %v", err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".recents-") {
			t.Errorf("a write left %s behind", entry.Name())
		}
	}
}

func TestTheFileIsReadableJSON(t *testing.T) {
	t.Parallel()

	r := newRecents(t, 20)

	if _, err := r.Add("/a"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw, err := os.ReadFile(r.Path)
	if err != nil {
		t.Fatalf("reading the file: %v", err)
	}

	var file recentsFile
	if err := json.Unmarshal(raw, &file); err != nil {
		t.Fatalf("the file is not readable JSON: %v", err)
	}
	if !reflect.DeepEqual(file.Paths, []string{"/a"}) {
		t.Errorf("the file holds %v", file.Paths)
	}
}

func TestWritingIntoAnUnwritableDirectoryIsReported(t *testing.T) {
	t.Parallel()

	if os.Geteuid() == 0 {
		t.Skip("running as root, so an unwritable directory is still writable")
	}

	root := t.TempDir()
	blocked := filepath.Join(root, "blocked")
	if err := os.Mkdir(blocked, 0o500); err != nil {
		t.Fatalf("creating the blocked directory: %v", err)
	}

	r := Recents{Path: filepath.Join(blocked, "state", "recents.json"), Cap: 20}

	if _, err := r.Add("/a"); err == nil {
		t.Fatal("expected an error writing into an unwritable directory")
	}
}

func TestWritingIntoADirectoryThatIsAFile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	blocker := filepath.Join(root, "state")

	if err := os.WriteFile(blocker, []byte("not a directory\n"), 0o644); err != nil {
		t.Fatalf("writing the blocker: %v", err)
	}

	r := Recents{Path: filepath.Join(blocker, "recents.json"), Cap: 20}

	if _, err := r.Add("/a"); err == nil {
		t.Fatal("expected an error when the parent is a file")
	}
}

func TestReadingADirectoryAsTheRecentsFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	r := Recents{Path: dir, Cap: 20}

	if got := r.Read(); len(got) != 0 {
		t.Errorf("reading a directory gave %v", got)
	}
}

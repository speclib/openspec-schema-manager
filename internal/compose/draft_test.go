package compose

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

var draftNow = time.Date(2026, 9, 29, 20, 0, 0, 0, time.UTC)

func newDrafts(t *testing.T) Drafts {
	t.Helper()

	return Drafts{
		Dir: filepath.Join(t.TempDir(), "drafts"),
		Now: func() time.Time { return draftNow },
	}
}

func TestSavingAndResumingADraft(t *testing.T) {
	t.Parallel()

	d := newDrafts(t)
	c := mixed(t)

	if err := d.Save("my-mix", c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	listed := d.List()
	if len(listed) != 1 {
		t.Fatalf("got %d drafts, want 1", len(listed))
	}
	if listed[0].Name != "my-mix" {
		t.Errorf("name = %q", listed[0].Name)
	}
	if !listed[0].Saved.Equal(draftNow) {
		t.Errorf("saved = %v, want %v", listed[0].Saved, draftNow)
	}

	resumed, err := d.Resume(listed[0].Path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resumed.Missing) != 0 {
		t.Errorf("missing = %v", resumed.Missing)
	}

	got := resumed.Composition
	if !reflect.DeepEqual(got.IDs(), c.IDs()) {
		t.Errorf("ids = %v, want %v", got.IDs(), c.IDs())
	}
	if !reflect.DeepEqual(got.Gates, c.Gates) {
		t.Errorf("gates = %v, want %v", got.Gates, c.Gates)
	}
	if got.Tracks != c.Tracks {
		t.Errorf("tracks = %q, want %q", got.Tracks, c.Tracks)
	}

	for i, a := range got.Artifacts {
		if !reflect.DeepEqual(a.Requires, c.Artifacts[i].Requires) {
			t.Errorf("%s requires %v, want %v", a.ID, a.Requires, c.Artifacts[i].Requires)
		}
		if a.SourceDir != c.Artifacts[i].SourceDir {
			t.Errorf("%s came from %q, want %q", a.ID, a.SourceDir, c.Artifacts[i].SourceDir)
		}
		if a.SourceID != c.Artifacts[i].SourceID {
			t.Errorf("%s had id %q, want %q", a.ID, a.SourceID, c.Artifacts[i].SourceID)
		}
	}
}

func TestAResumedDraftCanBeWritten(t *testing.T) {
	t.Parallel()

	d := newDrafts(t)

	if err := d.Save("my-mix", mixed(t)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resumed, err := d.Resume(d.List()[0].Path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	destination, err := resumed.Composition.Write(filepath.Join(t.TempDir(), "mine"), "from-draft")
	if err != nil {
		t.Fatalf("writing a resumed draft failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(destination, "templates")); err != nil {
		t.Errorf("the templates were not copied: %v", err)
	}
}

func TestSavingAgainReplacesTheDraft(t *testing.T) {
	t.Parallel()

	d := newDrafts(t)
	c := mixed(t)

	if err := d.Save("my-mix", c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := c.Remove("research"); err != nil {
		t.Fatalf("removing: %v", err)
	}
	if err := d.Save("my-mix", c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	listed := d.List()
	if len(listed) != 1 {
		t.Fatalf("got %d drafts, want 1", len(listed))
	}

	resumed, err := d.Resume(listed[0].Path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resumed.Composition.Has("research") {
		t.Error("the replacement did not take")
	}
}

func TestDraftsAreListedMostRecentFirst(t *testing.T) {
	t.Parallel()

	d := newDrafts(t)

	times := []time.Time{
		draftNow.Add(-2 * time.Hour),
		draftNow.Add(-time.Hour),
		draftNow,
	}

	for i, name := range []string{"oldest", "middle", "newest"} {
		at := times[i]
		d.Now = func() time.Time { return at }
		if err := d.Save(name, mixed(t)); err != nil {
			t.Fatalf("saving %s: %v", name, err)
		}
	}

	var names []string
	for _, listed := range d.List() {
		names = append(names, listed.Name)
	}

	if !reflect.DeepEqual(names, []string{"newest", "middle", "oldest"}) {
		t.Errorf("drafts = %v, want the newest first", names)
	}
}

func TestNoDraftsAtAll(t *testing.T) {
	t.Parallel()

	if got := newDrafts(t).List(); len(got) != 0 {
		t.Errorf("an empty drafts directory listed %v", got)
	}
}

func TestACorruptDraftDoesNotHideTheOthers(t *testing.T) {
	t.Parallel()

	d := newDrafts(t)

	if err := d.Save("good", mixed(t)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(d.Dir, "bad.json"), []byte("{oops"), 0o644); err != nil {
		t.Fatalf("writing the corrupt draft: %v", err)
	}

	listed := d.List()
	if len(listed) != 2 {
		t.Fatalf("got %d drafts, want both listed", len(listed))
	}

	var good, bad int
	for _, l := range listed {
		if l.Problem != "" {
			bad++
			continue
		}
		good++
		if _, err := d.Resume(l.Path); err != nil {
			t.Errorf("the good draft could not be resumed: %v", err)
		}
	}

	if good != 1 || bad != 1 {
		t.Errorf("listed %d good and %d bad", good, bad)
	}
}

func TestResumingADraftWhoseSourceHasGone(t *testing.T) {
	t.Parallel()

	d := newDrafts(t)
	c := mixed(t)

	if err := d.Save("my-mix", c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	gone := c.Artifacts[0].SourceDir
	if err := os.RemoveAll(gone); err != nil {
		t.Fatalf("removing the source: %v", err)
	}

	resumed, err := d.Resume(d.List()[0].Path)
	if err != nil {
		t.Fatalf("resuming failed rather than reporting: %v", err)
	}
	if len(resumed.Missing) == 0 {
		t.Fatal("the missing source was not reported")
	}
	if !strings.Contains(resumed.Missing[0], gone) {
		t.Errorf("the report does not name the directory: %q", resumed.Missing[0])
	}

	if err := resumed.Composition.Remove(c.Artifacts[0].ID); err != nil {
		t.Errorf("the resumed composition is not editable: %v", err)
	}
}

func TestResumingWhatCannotBeRead(t *testing.T) {
	t.Parallel()

	d := newDrafts(t)

	if _, err := d.Resume(filepath.Join(d.Dir, "absent.json")); err == nil {
		t.Error("resuming a missing file succeeded")
	}

	if err := os.MkdirAll(d.Dir, 0o755); err != nil {
		t.Fatalf("creating the drafts directory: %v", err)
	}

	bad := filepath.Join(d.Dir, "bad.json")
	if err := os.WriteFile(bad, []byte("{oops"), 0o644); err != nil {
		t.Fatalf("writing the draft: %v", err)
	}
	if _, err := d.Resume(bad); err == nil {
		t.Error("resuming an unparseable file succeeded")
	}

	empty := filepath.Join(d.Dir, "empty.json")
	if err := os.WriteFile(empty, []byte(`{"name":"x"}`), 0o644); err != nil {
		t.Fatalf("writing the draft: %v", err)
	}
	if _, err := d.Resume(empty); err == nil {
		t.Error("resuming a file with no composition succeeded")
	}
}

func TestSavingRefusesABadName(t *testing.T) {
	t.Parallel()

	d := newDrafts(t)

	for _, name := range []string{"", "..", "a/b"} {
		if err := d.Save(name, mixed(t)); err == nil {
			t.Errorf("Save accepted the name %q", name)
		}
	}
}

func TestSavingIntoAnUnwritableDirectory(t *testing.T) {
	t.Parallel()

	if os.Geteuid() == 0 {
		t.Skip("running as root, so an unwritable directory is still writable")
	}

	root := t.TempDir()
	blocked := filepath.Join(root, "blocked")
	if err := os.Mkdir(blocked, 0o500); err != nil {
		t.Fatalf("creating the blocked directory: %v", err)
	}

	d := Drafts{Dir: filepath.Join(blocked, "drafts")}

	if err := d.Save("x", mixed(t)); err == nil {
		t.Fatal("expected an error saving into an unwritable directory")
	}
}

func TestDraftsUseTheRealClockByDefault(t *testing.T) {
	t.Parallel()

	d := Drafts{Dir: filepath.Join(t.TempDir(), "drafts")}

	before := time.Now().Add(-time.Second)
	if err := d.Save("x", mixed(t)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if listed := d.List(); len(listed) != 1 || listed[0].Saved.Before(before) {
		t.Errorf("the draft was saved at %v", listed)
	}
}

func TestListingIgnoresWhatIsNotADraft(t *testing.T) {
	t.Parallel()

	d := newDrafts(t)

	if err := d.Save("good", mixed(t)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(d.Dir, "notes.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatalf("writing the stray file: %v", err)
	}
	if err := os.Mkdir(filepath.Join(d.Dir, "a-directory"), 0o755); err != nil {
		t.Fatalf("creating the directory: %v", err)
	}

	if got := d.List(); len(got) != 1 {
		t.Errorf("listed %d entries, want only the draft", len(got))
	}
}

func TestListingADirectoryThatIsNotThere(t *testing.T) {
	t.Parallel()

	d := Drafts{Dir: filepath.Join(t.TempDir(), "absent")}

	if got := d.List(); got != nil {
		t.Errorf("listing a missing directory gave %v", got)
	}
}

func TestListingADraftThatCannotBeRead(t *testing.T) {
	t.Parallel()

	if os.Geteuid() == 0 {
		t.Skip("running as root, so an unreadable file is still readable")
	}

	d := newDrafts(t)

	if err := d.Save("good", mixed(t)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	blocked := filepath.Join(d.Dir, "blocked.json")
	if err := os.WriteFile(blocked, []byte("{}"), 0o000); err != nil {
		t.Fatalf("writing the draft: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(blocked, 0o644) })

	listed := d.List()
	if len(listed) != 2 {
		t.Fatalf("got %d drafts, want both listed", len(listed))
	}

	var problems int
	for _, l := range listed {
		if l.Problem != "" {
			problems++
		}
	}
	if problems != 1 {
		t.Errorf("%d drafts were reported as problems", problems)
	}
}

package registry

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const registryURL = "https://registry.speclib.test/api/v1/openspec-schemas.json"

var fixedNow = time.Date(2026, 9, 29, 18, 0, 0, 0, time.UTC)

func newStore(t *testing.T) Store {
	t.Helper()

	return Store{
		Path: filepath.Join(t.TempDir(), "cache", "registry.json"),
		URL:  registryURL,
		Now:  func() time.Time { return fixedNow },
	}
}

func TestAMissingCacheIsNotAnError(t *testing.T) {
	t.Parallel()

	_, err := newStore(t).Read()
	if !errors.Is(err, ErrNoCache) {
		t.Fatalf("error = %v, want ErrNoCache", err)
	}
}

func TestWriteThenRead(t *testing.T) {
	t.Parallel()

	store := newStore(t)

	doc, err := Parse("fixture", fixture(t, "registry.json"))
	if err != nil {
		t.Fatalf("parsing the fixture: %v", err)
	}

	written, err := store.Write(doc)
	if err != nil {
		t.Fatalf("writing the cache: %v", err)
	}
	if !written.FetchedAt.Equal(fixedNow) {
		t.Errorf("FetchedAt = %v, want %v", written.FetchedAt, fixedNow)
	}

	read, err := store.Read()
	if err != nil {
		t.Fatalf("reading the cache: %v", err)
	}
	if len(read.Registry.Schemas) != 4 {
		t.Errorf("got %d entries, want 4", len(read.Registry.Schemas))
	}
	if read.URL != registryURL {
		t.Errorf("URL = %q", read.URL)
	}
}

func TestAgeAndStaleness(t *testing.T) {
	t.Parallel()

	cached := Cached{FetchedAt: fixedNow}

	if got := cached.Age(fixedNow.Add(3 * time.Hour)); got != 3*time.Hour {
		t.Errorf("Age = %v, want 3h", got)
	}
	if cached.Stale(fixedNow.Add(time.Hour), 24*time.Hour) {
		t.Error("an hour-old cache is stale against a 24h time to live")
	}
	if !cached.Stale(fixedNow.Add(25*time.Hour), 24*time.Hour) {
		t.Error("a 25-hour-old cache is not stale against a 24h time to live")
	}
	if !cached.Stale(fixedNow, 0) {
		t.Error("a time to live of zero does not make every cache stale")
	}
}

func TestACorruptCacheReadsAsNoCache(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"not json":            `{oops`,
		"not a cache record":  `{"hello":"world"}`,
		"the registry is bad": `{"fetched_at":"2026-09-29T18:00:00Z","url":"` + registryURL + `","registry":{"schemas":[{"id":"bad"}]}}`,
		"an empty file":       ``,
	}

	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			store := newStore(t)
			if err := os.MkdirAll(filepath.Dir(store.Path), 0o755); err != nil {
				t.Fatalf("creating the cache directory: %v", err)
			}
			if err := os.WriteFile(store.Path, []byte(body), 0o644); err != nil {
				t.Fatalf("writing the cache: %v", err)
			}

			if _, err := store.Read(); !errors.Is(err, ErrNoCache) {
				t.Errorf("error = %v, want ErrNoCache", err)
			}
		})
	}
}

func TestACacheForADifferentURLIsIgnored(t *testing.T) {
	t.Parallel()

	store := newStore(t)

	doc, err := Parse("fixture", fixture(t, "empty.json"))
	if err != nil {
		t.Fatalf("parsing the fixture: %v", err)
	}
	if _, err := store.Write(doc); err != nil {
		t.Fatalf("writing the cache: %v", err)
	}

	moved := store
	moved.URL = "https://mirror.example.test/openspec-schemas.json"

	if _, err := moved.Read(); !errors.Is(err, ErrNoCache) {
		t.Errorf("a cache recorded against another URL was served; error = %v", err)
	}
}

func TestRefreshWritesOnlyWhatParses(t *testing.T) {
	t.Parallel()

	store := newStore(t)

	good, err := Parse("fixture", fixture(t, "registry.json"))
	if err != nil {
		t.Fatalf("parsing the fixture: %v", err)
	}
	if _, err := store.Write(good); err != nil {
		t.Fatalf("seeding the cache: %v", err)
	}

	before, err := os.ReadFile(store.Path)
	if err != nil {
		t.Fatalf("reading the cache: %v", err)
	}

	cases := map[string]Fetcher{
		"an unreachable host": FetcherFunc(func(context.Context, string) ([]byte, error) {
			return nil, errors.New("no route to host")
		}),
		"a body that is not json": FetcherFunc(func(context.Context, string) ([]byte, error) {
			return []byte(`{oops`), nil
		}),
		"a body that is not a registry": FetcherFunc(func(context.Context, string) ([]byte, error) {
			return []byte(`{"entries":[]}`), nil
		}),
		"a registry with a bad entry": FetcherFunc(func(context.Context, string) ([]byte, error) {
			return []byte(`{"schemas":[{"id":"a/b"}]}`), nil
		}),
	}

	for name, fetcher := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := store.Refresh(t.Context(), fetcher); err == nil {
				t.Fatal("expected an error")
			}

			after, err := os.ReadFile(store.Path)
			if err != nil {
				t.Fatalf("reading the cache: %v", err)
			}
			if string(after) != string(before) {
				t.Error("a failed refresh changed the cache")
			}
		})
	}
}

func TestRefreshReplacesTheCacheOnSuccess(t *testing.T) {
	t.Parallel()

	store := newStore(t)

	seed, err := Parse("fixture", fixture(t, "registry.json"))
	if err != nil {
		t.Fatalf("parsing the fixture: %v", err)
	}
	if _, err := store.Write(seed); err != nil {
		t.Fatalf("seeding the cache: %v", err)
	}

	fresh := FetcherFunc(func(context.Context, string) ([]byte, error) {
		return fixture(t, "empty.json"), nil
	})

	cached, err := store.Refresh(t.Context(), fresh)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cached.Registry.Schemas) != 0 {
		t.Errorf("got %d entries, want none", len(cached.Registry.Schemas))
	}

	read, err := store.Read()
	if err != nil {
		t.Fatalf("reading the cache: %v", err)
	}
	if len(read.Registry.Schemas) != 0 {
		t.Error("the replacement did not reach the cache file")
	}
}

func TestWriteLeavesNoTemporaryFileBehind(t *testing.T) {
	t.Parallel()

	store := newStore(t)

	doc, err := Parse("fixture", fixture(t, "empty.json"))
	if err != nil {
		t.Fatalf("parsing the fixture: %v", err)
	}
	if _, err := store.Write(doc); err != nil {
		t.Fatalf("writing the cache: %v", err)
	}

	entries, err := os.ReadDir(filepath.Dir(store.Path))
	if err != nil {
		t.Fatalf("reading the cache directory: %v", err)
	}
	if len(entries) != 1 {
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Errorf("the cache directory holds %s; want only the cache file", strings.Join(names, ", "))
	}
}

func TestWriteReportsADirectoryItCannotCreate(t *testing.T) {
	t.Parallel()

	if os.Geteuid() == 0 {
		t.Skip("running as root, so an unwritable directory is still writable")
	}

	root := t.TempDir()
	blocked := filepath.Join(root, "blocked")
	if err := os.Mkdir(blocked, 0o500); err != nil {
		t.Fatalf("creating the blocked directory: %v", err)
	}

	store := Store{Path: filepath.Join(blocked, "sub", "registry.json"), URL: registryURL}

	if _, err := store.Write(Document{}); err == nil {
		t.Fatal("expected an error writing under an unwritable directory")
	}
}

func TestStoreUsesTheRealClockByDefault(t *testing.T) {
	t.Parallel()

	store := Store{Path: filepath.Join(t.TempDir(), "registry.json"), URL: registryURL}

	before := time.Now().Add(-time.Second)
	cached, err := store.Write(Document{})
	if err != nil {
		t.Fatalf("writing the cache: %v", err)
	}

	if cached.FetchedAt.Before(before) {
		t.Errorf("FetchedAt = %v, want a time from the real clock", cached.FetchedAt)
	}
}

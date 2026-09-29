package registry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

type Cached struct {
	FetchedAt time.Time `json:"fetched_at"`
	URL       string    `json:"url"`
	Registry  Document  `json:"registry"`
}

func (c Cached) Age(now time.Time) time.Duration {
	return now.Sub(c.FetchedAt)
}

func (c Cached) Stale(now time.Time, ttl time.Duration) bool {
	return c.Age(now) >= ttl
}

type Store struct {
	Path string
	URL  string
	Now  func() time.Time
}

var ErrNoCache = errors.New("the registry has never been fetched")

func (s Store) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}

func (s Store) Read() (Cached, error) {
	raw, err := os.ReadFile(s.Path)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return Cached{}, ErrNoCache
	case err != nil:
		return Cached{}, ErrNoCache
	}

	var cached Cached
	if err := json.Unmarshal(raw, &cached); err != nil {
		return Cached{}, ErrNoCache
	}

	if cached.URL != s.URL {
		return Cached{}, ErrNoCache
	}

	if _, err := Parse(s.Path, mustMarshal(cached.Registry)); err != nil {
		return Cached{}, ErrNoCache
	}

	return cached, nil
}

func mustMarshal(doc Document) []byte {
	raw, err := json.Marshal(doc)
	if err != nil {
		return []byte(`{}`)
	}
	return raw
}

func (s Store) Write(doc Document) (Cached, error) {
	cached := Cached{FetchedAt: s.now().UTC(), URL: s.URL, Registry: doc}

	raw, err := json.Marshal(cached)
	if err != nil {
		return Cached{}, fmt.Errorf("writing %s: %w", s.Path, err)
	}

	if err := os.MkdirAll(filepath.Dir(s.Path), 0o755); err != nil {
		return Cached{}, fmt.Errorf("writing %s: %w", s.Path, err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(s.Path), ".registry-*")
	if err != nil {
		return Cached{}, fmt.Errorf("writing %s: %w", s.Path, err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(raw); err != nil {
		tmp.Close()
		return Cached{}, fmt.Errorf("writing %s: %w", s.Path, err)
	}
	if err := tmp.Close(); err != nil {
		return Cached{}, fmt.Errorf("writing %s: %w", s.Path, err)
	}

	if err := os.Rename(tmp.Name(), s.Path); err != nil {
		return Cached{}, fmt.Errorf("writing %s: %w", s.Path, err)
	}

	return cached, nil
}

func (s Store) Refresh(ctx context.Context, fetcher Fetcher) (Cached, error) {
	raw, err := fetcher.Fetch(ctx, s.URL)
	if err != nil {
		return Cached{}, err
	}

	doc, err := Parse(s.URL, raw)
	if err != nil {
		return Cached{}, err
	}

	return s.Write(doc)
}

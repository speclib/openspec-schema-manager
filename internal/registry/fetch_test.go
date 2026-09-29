package registry

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestHTTPFetcherReadsASuccessfulResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Errorf("Accept = %q, want application/json", got)
		}
		_, _ = w.Write([]byte(`{"schemas":[]}`))
	}))
	defer server.Close()

	got, err := NewHTTPFetcher().Fetch(t.Context(), server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != `{"schemas":[]}` {
		t.Errorf("body = %q", got)
	}
}

func TestHTTPFetcherReportsANonSuccessStatus(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "gone", http.StatusNotFound)
	}))
	defer server.Close()

	_, err := NewHTTPFetcher().Fetch(t.Context(), server.URL)
	if err == nil {
		t.Fatal("expected an error for a 404")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("error %q does not name the status", err)
	}
}

func TestHTTPFetcherRefusesAnOversizedBody(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("x", 100)))
	}))
	defer server.Close()

	fetcher := NewHTTPFetcher()
	fetcher.MaxSize = 10

	_, err := fetcher.Fetch(t.Context(), server.URL)
	if err == nil {
		t.Fatal("expected an error for an oversized body")
	}
	if !strings.Contains(err.Error(), "larger than") {
		t.Errorf("error %q does not say the response was too large", err)
	}
}

func TestHTTPFetcherHonoursItsTimeout(t *testing.T) {
	t.Parallel()

	block := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		<-block
	}))
	defer func() {
		close(block)
		server.Close()
	}()

	fetcher := NewHTTPFetcher()
	fetcher.Timeout = 50 * time.Millisecond
	fetcher.Client = &http.Client{}

	if _, err := fetcher.Fetch(t.Context(), server.URL); err == nil {
		t.Fatal("expected a timeout")
	}
}

func TestHTTPFetcherReportsAnUnreachableHost(t *testing.T) {
	t.Parallel()

	fetcher := NewHTTPFetcher()
	fetcher.Timeout = 200 * time.Millisecond

	if _, err := fetcher.Fetch(t.Context(), "http://127.0.0.1:1/openspec-schemas.json"); err == nil {
		t.Fatal("expected an error for an unreachable host")
	}
}

func TestHTTPFetcherReadsAFileURL(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "openspec-schemas.json")
	if err := os.WriteFile(path, []byte(`{"schemas":[]}`), 0o644); err != nil {
		t.Fatalf("writing the fixture: %v", err)
	}

	got, err := NewHTTPFetcher().Fetch(t.Context(), "file://"+path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != `{"schemas":[]}` {
		t.Errorf("body = %q", got)
	}
}

func TestHTTPFetcherReportsAMissingFile(t *testing.T) {
	t.Parallel()

	_, err := NewHTTPFetcher().Fetch(t.Context(), "file://"+filepath.Join(t.TempDir(), "absent.json"))
	if err == nil {
		t.Fatal("expected an error for a missing file")
	}
}

func TestHTTPFetcherRefusesAnOversizedFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "big.json")
	if err := os.WriteFile(path, []byte(strings.Repeat("x", 100)), 0o644); err != nil {
		t.Fatalf("writing the fixture: %v", err)
	}

	fetcher := NewHTTPFetcher()
	fetcher.MaxSize = 10

	if _, err := fetcher.Fetch(t.Context(), "file://"+path); err == nil {
		t.Fatal("expected an error for an oversized file")
	}
}

func TestHTTPFetcherReportsAMalformedURL(t *testing.T) {
	t.Parallel()

	if _, err := NewHTTPFetcher().Fetch(t.Context(), "://no-scheme"); err == nil {
		t.Fatal("expected an error for a malformed URL")
	}
}

func TestFetcherFuncIsAFetcher(t *testing.T) {
	t.Parallel()

	var called string
	var f Fetcher = FetcherFunc(func(_ context.Context, url string) ([]byte, error) {
		called = url
		return []byte("body"), nil
	})

	got, err := f.Fetch(t.Context(), "https://example.test/r.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != "body" || called != "https://example.test/r.json" {
		t.Errorf("got %q for %q", got, called)
	}
}

func TestAFetcherWithNoClientStillWorks(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"schemas":[]}`))
	}))
	defer server.Close()

	fetcher := &HTTPFetcher{}

	got, err := fetcher.Fetch(t.Context(), server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != `{"schemas":[]}` {
		t.Errorf("body = %q", got)
	}
	if fetcher.maxSize() != MaxBodySize {
		t.Errorf("maxSize = %d, want the default", fetcher.maxSize())
	}
}

func TestFetchErrorsAreNotSwallowed(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("no route to host")

	_, err := FetcherFunc(func(context.Context, string) ([]byte, error) {
		return nil, sentinel
	}).Fetch(t.Context(), "https://example.test/r.json")

	if !errors.Is(err, sentinel) {
		t.Errorf("error = %v, want the fetcher's own", err)
	}
}

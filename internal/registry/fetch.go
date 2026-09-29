package registry

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	MaxBodySize    = 8 << 20
	defaultTimeout = 20 * time.Second
)

type Fetcher interface {
	Fetch(ctx context.Context, url string) ([]byte, error)
}

type FetcherFunc func(ctx context.Context, url string) ([]byte, error)

func (f FetcherFunc) Fetch(ctx context.Context, url string) ([]byte, error) {
	return f(ctx, url)
}

type HTTPFetcher struct {
	Client  *http.Client
	Timeout time.Duration
	MaxSize int64
}

func NewHTTPFetcher() *HTTPFetcher {
	return &HTTPFetcher{
		Client:  &http.Client{Timeout: defaultTimeout},
		Timeout: defaultTimeout,
		MaxSize: MaxBodySize,
	}
}

func (f *HTTPFetcher) Fetch(ctx context.Context, raw string) ([]byte, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("fetching %s: %w", raw, err)
	}

	if parsed.Scheme == "file" {
		return f.fetchFile(parsed)
	}

	return f.fetchHTTP(ctx, raw)
}

func (f *HTTPFetcher) fetchFile(parsed *url.URL) ([]byte, error) {
	path := parsed.Path
	if parsed.Host != "" && parsed.Host != "localhost" {
		path = parsed.Host + path
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("fetching %s: %w", parsed, err)
	}

	if int64(len(raw)) > f.maxSize() {
		return nil, fmt.Errorf("fetching %s: the file is larger than %d bytes", parsed, f.maxSize())
	}

	return raw, nil
}

func (f *HTTPFetcher) fetchHTTP(ctx context.Context, raw string) ([]byte, error) {
	if f.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, f.Timeout)
		defer cancel()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return nil, fmt.Errorf("fetching %s: %w", raw, err)
	}
	req.Header.Set("Accept", "application/json")

	client := f.Client
	if client == nil {
		client = &http.Client{Timeout: f.Timeout}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching %s: %w", raw, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("fetching %s: the server answered %s", raw, resp.Status)
	}

	limit := f.maxSize()
	body, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, fmt.Errorf("fetching %s: %w", raw, err)
	}

	if int64(len(body)) > limit {
		return nil, fmt.Errorf("fetching %s: the response is larger than %d bytes", raw, limit)
	}

	return body, nil
}

func (f *HTTPFetcher) maxSize() int64 {
	if f.MaxSize > 0 {
		return f.MaxSize
	}
	return MaxBodySize
}

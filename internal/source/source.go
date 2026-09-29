package source

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/speclib/openspec-schema-manager/internal/schema"
)

var (
	ErrNoGit     = errors.New("git is needed to fetch a schema and is not on PATH")
	ErrNoSchema  = errors.New("the fetched directory holds no schema.yaml")
	ErrNotCached = errors.New("this source has not been fetched")
)

const defaultTimeout = 2 * time.Minute

type Source struct {
	Repo string `json:"repo"`
	Path string `json:"path"`
	Ref  string `json:"ref,omitempty"`
}

func (s Source) Key() string {
	sum := sha256.Sum256([]byte(s.Repo + "\x00" + s.Ref + "\x00" + s.Path))
	return hex.EncodeToString(sum[:])
}

func (s Source) RefLabel() string {
	if s.Ref == "" {
		return "default branch"
	}
	return s.Ref
}

type Fetcher interface {
	Fetch(ctx context.Context, src Source, refetch bool) (string, error)
}

type Git struct {
	CacheRoot string
	Binary    string
	Timeout   time.Duration
}

func NewGit(cacheRoot string) Git {
	return Git{CacheRoot: cacheRoot, Binary: "git", Timeout: defaultTimeout}
}

func (g Git) binary() string {
	if g.Binary != "" {
		return g.Binary
	}
	return "git"
}

func (g Git) Dir(src Source) string {
	return filepath.Join(g.CacheRoot, src.Key())
}

func (g Git) Cached(src Source) (string, error) {
	dir := g.Dir(src)

	if _, err := os.Stat(filepath.Join(dir, schema.SchemaFile)); err != nil {
		return "", ErrNotCached
	}

	return dir, nil
}

func (g Git) Fetch(ctx context.Context, src Source, refetch bool) (string, error) {
	if !refetch {
		if dir, err := g.Cached(src); err == nil {
			return dir, nil
		}
	}

	if _, err := exec.LookPath(g.binary()); err != nil {
		return "", ErrNoGit
	}

	if err := os.MkdirAll(g.CacheRoot, 0o755); err != nil {
		return "", fmt.Errorf("preparing the schema cache: %w", err)
	}

	work, err := os.MkdirTemp(g.CacheRoot, ".fetch-*")
	if err != nil {
		return "", fmt.Errorf("preparing the schema cache: %w", err)
	}
	defer os.RemoveAll(work)

	checkout := filepath.Join(work, "checkout")
	if err := os.MkdirAll(checkout, 0o755); err != nil {
		return "", fmt.Errorf("preparing the schema cache: %w", err)
	}

	if err := g.clone(ctx, checkout, src); err != nil {
		return "", err
	}

	wanted := filepath.Join(checkout, filepath.FromSlash(src.Path))
	info, err := os.Stat(wanted)
	if err != nil || !info.IsDir() {
		return "", fmt.Errorf("%s holds no directory %s at %s", src.Repo, src.Path, src.RefLabel())
	}

	staged := filepath.Join(work, "staged")
	if err := copyTree(wanted, staged); err != nil {
		return "", err
	}

	if err := verify(staged); err != nil {
		return "", err
	}

	if err := writeSourceRecord(staged, src); err != nil {
		return "", err
	}

	final := g.Dir(src)
	if err := os.RemoveAll(final); err != nil {
		return "", fmt.Errorf("replacing the cached schema: %w", err)
	}
	if err := os.Rename(staged, final); err != nil {
		return "", fmt.Errorf("replacing the cached schema: %w", err)
	}

	return final, nil
}

func (g Git) clone(ctx context.Context, dir string, src Source) error {
	ref := src.Ref
	if ref == "" {
		ref = "HEAD"
	}

	steps := [][]string{
		{"init", "--quiet"},
		{"remote", "add", "origin", src.Repo},
		{"fetch", "--quiet", "--depth", "1", "origin", ref},
		{"checkout", "--quiet", "--detach", "FETCH_HEAD"},
	}

	for _, args := range steps {
		if err := g.run(ctx, dir, args...); err != nil {
			return err
		}
	}

	return nil
}

func (g Git) run(ctx context.Context, dir string, args ...string) error {
	timeout := g.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, g.binary(), args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_TERMINAL_PROMPT=0",
		"GIT_ASKPASS=",
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_CONFIG_SYSTEM=/dev/null",
	)
	cmd.WaitDelay = 5 * time.Second

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git %s: %w: %s", args[0], err, bytes.TrimSpace(stderr.Bytes()))
	}

	return nil
}

func verify(dir string) error {
	raw, err := os.ReadFile(filepath.Join(dir, schema.SchemaFile))
	if err != nil {
		return ErrNoSchema
	}

	if _, err := schema.Parse(schema.SchemaFile, raw); err != nil {
		return err
	}

	return nil
}

func writeSourceRecord(dir string, src Source) error {
	raw, err := json.MarshalIndent(src, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, "source.json"), append(raw, '\n'), 0o644)
}

func ReadSourceRecord(dir string) (Source, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "source.json"))
	if err != nil {
		return Source{}, err
	}

	var src Source
	if err := json.Unmarshal(raw, &src); err != nil {
		return Source{}, err
	}

	return src, nil
}

func copyTree(from, to string) error {
	return filepath.WalkDir(from, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(from, path)
		if err != nil {
			return err
		}
		target := filepath.Join(to, rel)

		switch {
		case d.IsDir():
			return os.MkdirAll(target, 0o755)
		case !d.Type().IsRegular():
			return nil
		}

		return copyFile(path, target)
	})
}

func copyFile(from, to string) error {
	if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
		return err
	}

	in, err := os.Open(from)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(to)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	return out.Close()
}

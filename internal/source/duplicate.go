package source

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/speclib/openspec-schema-manager/internal/schema"
)

var (
	ErrEmptyName        = errors.New("a schema needs a name, because the name becomes its directory")
	ErrNameHasSeparator = errors.New("a schema name becomes one directory, so it cannot hold a path separator")
	ErrNameIsRelative   = errors.New("a schema cannot be named . or ..")
	ErrDestinationBusy  = errors.New("the destination already exists")
)

func ValidName(name string) error {
	trimmed := strings.TrimSpace(name)

	switch {
	case trimmed == "":
		return ErrEmptyName
	case trimmed == "." || trimmed == "..":
		return ErrNameIsRelative
	case strings.ContainsAny(trimmed, `/\`):
		return ErrNameHasSeparator
	}

	return nil
}

// Duplicate copies a schema folder under a new name.
//
// ossm does this rather than calling openspec schema fork. Fork takes a schema
// OpenSpec already resolves rather than a path, so it cannot copy a registry
// schema or a folder opened by path; and it preserves the source's mode, which
// makes it fail on any schema in the Nix store. NOTES.md records both.
func Duplicate(from, intoDir, name string) (string, error) {
	if err := ValidName(name); err != nil {
		return "", err
	}

	name = strings.TrimSpace(name)
	destination := filepath.Join(intoDir, name)

	if _, err := os.Stat(destination); err == nil {
		return "", fmt.Errorf("%w: %s", ErrDestinationBusy, destination)
	}

	if _, err := os.Stat(filepath.Join(from, schema.SchemaFile)); err != nil {
		return "", fmt.Errorf("%s holds no %s", from, schema.SchemaFile)
	}

	if err := os.MkdirAll(intoDir, 0o755); err != nil {
		return "", fmt.Errorf("preparing %s: %w", intoDir, err)
	}

	staged, err := os.MkdirTemp(intoDir, ".duplicate-*")
	if err != nil {
		return "", fmt.Errorf("preparing %s: %w", intoDir, err)
	}
	defer os.RemoveAll(staged)

	content := filepath.Join(staged, "content")
	if err := copyWritable(from, content); err != nil {
		return "", fmt.Errorf("copying the schema: %w", err)
	}

	if err := rewriteName(filepath.Join(content, schema.SchemaFile), name); err != nil {
		return "", err
	}

	if err := os.Rename(content, destination); err != nil {
		return "", fmt.Errorf("writing %s: %w", destination, err)
	}

	return destination, nil
}

// rewriteName edits the name: line as text rather than re-serialising the
// document. A schema carries comments, instruction blocks and a deliberate
// artifact order, and a YAML round trip loses the first and can reorder the
// rest.
func rewriteName(path, name string) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}

	lines := strings.Split(string(raw), "\n")

	replaced := false
	for i, line := range lines {
		if strings.HasPrefix(line, "name:") {
			lines[i] = "name: " + name
			replaced = true
			break
		}
	}

	if !replaced {
		lines = append([]string{"name: " + name}, lines...)
	}

	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644)
}

// copyWritable copies a tree writing every file mode 644, so a copy of a
// read-only source is editable. This is the half of openspec schema fork that
// fails on a Nix install.
func copyWritable(from, to string) error {
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

		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}

		return os.WriteFile(target, body, 0o644)
	})
}

type File struct {
	Rel   string
	Path  string
	Depth int
}

func Files(dir string) ([]File, error) {
	var files []File

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !d.Type().IsRegular() {
			return nil
		}

		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		if rel != schema.SchemaFile && !strings.HasPrefix(rel, "templates/") {
			return nil
		}

		files = append(files, File{Rel: rel, Path: path, Depth: strings.Count(rel, "/")})

		return nil
	})
	if err != nil {
		return nil, err
	}

	sortFiles(files)

	return files, nil
}

func sortFiles(files []File) {
	// schema.yaml first, then templates in path order. The order has to be the
	// same every time or the cursor lands on a different file between runs.
	rank := func(f File) int {
		if f.Rel == schema.SchemaFile {
			return 0
		}
		return 1
	}

	for i := 1; i < len(files); i++ {
		for j := i; j > 0; j-- {
			a, b := files[j-1], files[j]
			if rank(a) < rank(b) || (rank(a) == rank(b) && a.Rel <= b.Rel) {
				break
			}
			files[j-1], files[j] = b, a
		}
	}
}

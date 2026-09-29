package openspec

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var (
	ErrNotAProject  = errors.New("installing a schema needs an OpenSpec project; browsing works anywhere")
	ErrAlreadyThere = errors.New("the destination already holds a schema")
	ErrEmptySource  = errors.New("the source directory holds no files")
	ErrNoSchemaName = errors.New("the schema declares no name, so there is nowhere to install it")
)

type Plan struct {
	Name        string
	Source      string
	Destination string
	// Relative is the destination as a user reads it: relative to the project
	// root. The absolute path is often longer than a pane, and truncating it
	// cuts off the half that says which schema is being written.
	Relative string
	Root     string
	Files    []string
	Occupied bool
	Existing []string
}

func Destination(root, name string) string {
	return filepath.Join(root, "openspec", "schemas", name)
}

func PlanInstall(root, source, name string) (Plan, error) {
	if strings.TrimSpace(name) == "" {
		return Plan{}, ErrNoSchemaName
	}

	files, err := listFiles(source)
	if err != nil {
		return Plan{}, err
	}
	if len(files) == 0 {
		return Plan{}, ErrEmptySource
	}

	destination := Destination(root, name)

	relative, err := filepath.Rel(root, destination)
	if err != nil {
		relative = destination
	}

	plan := Plan{
		Name:        name,
		Source:      source,
		Destination: destination,
		Relative:    filepath.ToSlash(relative),
		Root:        root,
		Files:       files,
	}

	if existing, err := listFiles(destination); err == nil {
		plan.Occupied = true
		plan.Existing = existing
	}

	return plan, nil
}

func Install(plan Plan, overwrite bool) error {
	if plan.Occupied && !overwrite {
		return ErrAlreadyThere
	}

	parent := filepath.Dir(plan.Destination)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return fmt.Errorf("preparing %s: %w", parent, err)
	}

	staged, err := os.MkdirTemp(parent, ".install-*")
	if err != nil {
		return fmt.Errorf("preparing %s: %w", parent, err)
	}
	defer os.RemoveAll(staged)

	content := filepath.Join(staged, "content")
	if err := copyTree(plan.Source, content); err != nil {
		return fmt.Errorf("copying the schema: %w", err)
	}

	var displaced string
	if plan.Occupied {
		displaced = filepath.Join(staged, "displaced")
		if err := os.Rename(plan.Destination, displaced); err != nil {
			return fmt.Errorf("moving the existing schema aside: %w", err)
		}
	}

	if err := os.Rename(content, plan.Destination); err != nil {
		if displaced != "" {
			_ = os.Rename(displaced, plan.Destination)
		}
		return fmt.Errorf("writing %s: %w", plan.Destination, err)
	}

	return nil
}

func listFiles(dir string) ([]string, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", dir)
	}

	var files []string

	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
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
		files = append(files, filepath.ToSlash(rel))

		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Strings(files)

	return files, nil
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

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}

		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()

		out, err := os.Create(target)
		if err != nil {
			return err
		}
		defer out.Close()

		if _, err := io.Copy(out, in); err != nil {
			return err
		}

		return out.Close()
	})
}

// SetDefaultSchema rewrites only the schema: line in openspec/config.yaml.
// The file is full of commented guidance a user may have edited, so it is
// edited in place rather than re-serialised.
func SetDefaultSchema(root, name string) error {
	path := filepath.Join(root, "openspec", "config.yaml")

	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}

	lines := strings.Split(string(raw), "\n")

	replaced := false
	for i, line := range lines {
		if strings.HasPrefix(line, "schema:") {
			lines[i] = "schema: " + name
			replaced = true
			break
		}
	}

	if !replaced {
		lines = append([]string{"schema: " + name}, lines...)
	}

	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}

	return nil
}

func ValidateInstalled(ctx context.Context, cli CLI, root, name string) (Validation, error) {
	if cli == nil {
		return Validation{}, errors.New("no OpenSpec adapter is configured, so the install is unverified")
	}

	return cli.ValidateSchema(ctx, root, name)
}

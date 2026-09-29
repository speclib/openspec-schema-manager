package compose

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/speclib/openspec-schema-manager/internal/source"
)

type draftFile struct {
	Name        string       `json:"name"`
	Saved       time.Time    `json:"saved"`
	Composition *Composition `json:"composition"`
}

type Drafts struct {
	Dir string
	Now func() time.Time
}

func (d Drafts) now() time.Time {
	if d.Now != nil {
		return d.Now()
	}
	return time.Now()
}

func (d Drafts) path(name string) string {
	return filepath.Join(d.Dir, sanitise(name)+".json")
}

func (d Drafts) Save(name string, c *Composition) error {
	if err := source.ValidName(name); err != nil {
		return err
	}

	name = strings.TrimSpace(name)

	raw, err := json.MarshalIndent(draftFile{Name: name, Saved: d.now().UTC(), Composition: c}, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(d.Dir, 0o755); err != nil {
		return fmt.Errorf("preparing %s: %w", d.Dir, err)
	}

	tmp, err := os.CreateTemp(d.Dir, ".draft-*")
	if err != nil {
		return fmt.Errorf("preparing %s: %w", d.Dir, err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(raw); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	return os.Rename(tmp.Name(), d.path(name))
}

type Listed struct {
	Name    string
	Path    string
	Saved   time.Time
	Problem string
}

func (d Drafts) List() []Listed {
	entries, err := os.ReadDir(d.Dir)
	if err != nil {
		return nil
	}

	var listed []Listed

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		path := filepath.Join(d.Dir, entry.Name())

		raw, err := os.ReadFile(path)
		if err != nil {
			listed = append(listed, Listed{Name: entry.Name(), Path: path, Problem: err.Error()})
			continue
		}

		var file draftFile
		if err := json.Unmarshal(raw, &file); err != nil {
			listed = append(listed, Listed{Name: entry.Name(), Path: path, Problem: "this draft could not be read"})
			continue
		}

		listed = append(listed, Listed{Name: file.Name, Path: path, Saved: file.Saved})
	}

	sort.SliceStable(listed, func(i, j int) bool {
		return listed[i].Saved.After(listed[j].Saved)
	})

	return listed
}

type Resumed struct {
	Composition *Composition
	Missing     []string
}

func (d Drafts) Resume(path string) (Resumed, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Resumed{}, fmt.Errorf("reading %s: %w", path, err)
	}

	var file draftFile
	if err := json.Unmarshal(raw, &file); err != nil {
		return Resumed{}, fmt.Errorf("reading %s: %w", path, err)
	}

	if file.Composition == nil {
		return Resumed{}, fmt.Errorf("reading %s: it holds no composition", path)
	}

	resumed := Resumed{Composition: file.Composition}

	seen := make(map[string]bool)
	for _, a := range file.Composition.Artifacts {
		if a.SourceDir == "" || seen[a.SourceDir] {
			continue
		}
		if _, err := os.Stat(a.SourceDir); err != nil {
			seen[a.SourceDir] = true
			resumed.Missing = append(resumed.Missing, a.ID+" came from "+a.SourceDir+", which is no longer there")
		}
	}

	return resumed, nil
}

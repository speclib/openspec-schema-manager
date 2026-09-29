package source

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/speclib/openspec-schema-manager/internal/schema"
)

type Local struct {
	Dir        string
	Directory  string
	Name       string
	Artifacts  int
	Unreadable string
}

func (l Local) Readable() bool { return l.Unreadable == "" }

func (l Local) Label() string {
	if l.Readable() {
		return l.Name
	}
	return filepath.Base(l.Dir)
}

type Scan struct {
	Directory string
	Schemas   []Local
	Problem   string
}

func ScanDir(dir string) Scan {
	scan := Scan{Directory: dir}

	entries, err := os.ReadDir(dir)
	if err != nil {
		scan.Problem = fmt.Sprintf("%s could not be read: %v", dir, err)
		return scan
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		candidate := filepath.Join(dir, entry.Name())
		if local, ok := Read(candidate, dir); ok {
			scan.Schemas = append(scan.Schemas, local)
		}
	}

	sort.Slice(scan.Schemas, func(i, j int) bool {
		return strings.ToLower(scan.Schemas[i].Label()) < strings.ToLower(scan.Schemas[j].Label())
	})

	return scan
}

func Read(dir, directory string) (Local, bool) {
	if _, err := os.Stat(filepath.Join(dir, schema.SchemaFile)); err != nil {
		return Local{}, false
	}

	local := Local{Dir: dir, Directory: directory}

	s, err := schema.Load(dir)
	if err != nil {
		local.Unreadable = err.Error()
		return local, true
	}

	local.Name = s.Name
	local.Artifacts = len(s.Artifacts)

	if strings.TrimSpace(local.Name) == "" {
		local.Name = filepath.Base(dir)
	}

	return local, true
}

func ScanAll(dirs []string) []Scan {
	scans := make([]Scan, 0, len(dirs))
	for _, dir := range dirs {
		scans = append(scans, ScanDir(dir))
	}
	return scans
}

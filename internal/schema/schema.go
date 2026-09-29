package schema

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const SchemaFile = "schema.yaml"

type Artifact struct {
	ID          string   `yaml:"id"`
	Generates   string   `yaml:"generates"`
	Template    string   `yaml:"template"`
	Description string   `yaml:"description"`
	Instruction string   `yaml:"instruction"`
	Requires    []string `yaml:"requires"`
}

type Apply struct {
	Requires    []string `yaml:"requires"`
	Tracks      string   `yaml:"tracks"`
	Instruction string   `yaml:"instruction"`
}

type Schema struct {
	Name        string     `yaml:"name"`
	Version     int        `yaml:"version"`
	Description string     `yaml:"description"`
	Artifacts   []Artifact `yaml:"artifacts"`
	Apply       Apply      `yaml:"apply"`
}

func Parse(source string, raw []byte) (Schema, error) {
	dec := yaml.NewDecoder(strings.NewReader(string(raw)))
	dec.KnownFields(true)

	var s Schema
	if err := dec.Decode(&s); err != nil && err.Error() != "EOF" {
		return Schema{}, fmt.Errorf("reading %s: %w", source, err)
	}

	for i := range s.Artifacts {
		if len(s.Artifacts[i].Requires) == 0 {
			s.Artifacts[i].Requires = nil
		}
	}
	if len(s.Apply.Requires) == 0 {
		s.Apply.Requires = nil
	}

	return s, nil
}

func Load(dir string) (Schema, error) {
	path := filepath.Join(dir, SchemaFile)

	raw, err := os.ReadFile(path)
	if err != nil {
		return Schema{}, fmt.Errorf("reading %s: %w", path, err)
	}

	return Parse(path, raw)
}

func (s Schema) Artifact(id string) (Artifact, bool) {
	for _, a := range s.Artifacts {
		if a.ID == id {
			return a, true
		}
	}
	return Artifact{}, false
}

func (s Schema) IDs() []string {
	ids := make([]string, 0, len(s.Artifacts))
	for _, a := range s.Artifacts {
		ids = append(ids, a.ID)
	}
	return ids
}

func (s Schema) Gates() map[string]bool {
	gates := make(map[string]bool, len(s.Apply.Requires))
	for _, id := range s.Apply.Requires {
		gates[id] = true
	}
	return gates
}

func (a Artifact) TemplatePath(dir string) string {
	return filepath.Join(dir, "templates", filepath.FromSlash(a.Template))
}

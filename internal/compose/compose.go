package compose

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/speclib/openspec-schema-manager/internal/schema"
)

var (
	ErrCollision   = errors.New("that id is already on the canvas")
	ErrSelfLink    = errors.New("an artifact cannot require itself")
	ErrNotOnCanvas = errors.New("that artifact is not on the canvas")
)

type Artifact struct {
	ID          string   `json:"id"`
	Generates   string   `json:"generates"`
	Description string   `json:"description"`
	Template    string   `json:"template"`
	Instruction string   `json:"instruction,omitempty"`
	Requires    []string `json:"requires,omitempty"`

	SourceDir  string `json:"source_dir"`
	SourceName string `json:"source_name"`
	SourceRef  string `json:"source_ref,omitempty"`
	SourceID   string `json:"source_id"`
}

func (a Artifact) Renamed() bool { return a.ID != a.SourceID }

type Source struct {
	Dir       string            `json:"dir"`
	Name      string            `json:"name"`
	Ref       string            `json:"ref,omitempty"`
	Artifacts []schema.Artifact `json:"artifacts"`
}

type Composition struct {
	Artifacts []Artifact `json:"artifacts"`
	Gates     []string   `json:"gates,omitempty"`
	Tracks    string     `json:"tracks,omitempty"`
	Sources   []Source   `json:"sources,omitempty"`
}

func New() *Composition { return &Composition{} }

func FromSchema(s schema.Schema, dir, ref string) *Composition {
	c := New()
	c.AddSource(s, dir, ref)

	for _, a := range s.Artifacts {
		c.Artifacts = append(c.Artifacts, Artifact{
			ID:          a.ID,
			Generates:   a.Generates,
			Description: a.Description,
			Template:    a.Template,
			Instruction: a.Instruction,
			Requires:    append([]string{}, a.Requires...),
			SourceDir:   dir,
			SourceName:  s.Name,
			SourceRef:   ref,
			SourceID:    a.ID,
		})
	}

	c.Gates = append([]string{}, s.Apply.Requires...)
	c.Tracks = s.Apply.Tracks

	return c
}

func (c *Composition) AddSource(s schema.Schema, dir, ref string) {
	for _, existing := range c.Sources {
		if existing.Dir == dir {
			return
		}
	}

	c.Sources = append(c.Sources, Source{
		Dir:       dir,
		Name:      s.Name,
		Ref:       ref,
		Artifacts: append([]schema.Artifact{}, s.Artifacts...),
	})
}

func (c *Composition) Has(id string) bool {
	return slices.ContainsFunc(c.Artifacts, func(a Artifact) bool { return a.ID == id })
}

func (c *Composition) IDs() []string {
	ids := make([]string, 0, len(c.Artifacts))
	for _, a := range c.Artifacts {
		ids = append(ids, a.ID)
	}
	return ids
}

// Add copies an artifact onto the canvas, keeping only the requirements that
// already resolve. An artifact does not pull its dependencies in behind it: the
// user chose what to link, and a composer that keeps adding things nobody asked
// for is worse than one that adds too little.
func (c *Composition) Add(source Source, a schema.Artifact, as string) error {
	id := strings.TrimSpace(as)
	if id == "" {
		id = a.ID
	}

	if c.Has(id) {
		return fmt.Errorf("%w: %s", ErrCollision, id)
	}

	var requires []string
	for _, req := range a.Requires {
		if c.Has(req) {
			requires = append(requires, req)
		}
	}

	c.Artifacts = append(c.Artifacts, Artifact{
		ID:          id,
		Generates:   a.Generates,
		Description: a.Description,
		Template:    a.Template,
		Instruction: a.Instruction,
		Requires:    requires,
		SourceDir:   source.Dir,
		SourceName:  source.Name,
		SourceRef:   source.Ref,
		SourceID:    a.ID,
	})

	return nil
}

func (c *Composition) Remove(id string) error {
	index := slices.IndexFunc(c.Artifacts, func(a Artifact) bool { return a.ID == id })
	if index < 0 {
		return fmt.Errorf("%w: %s", ErrNotOnCanvas, id)
	}

	c.Artifacts = slices.Delete(c.Artifacts, index, index+1)

	for i := range c.Artifacts {
		c.Artifacts[i].Requires = without(c.Artifacts[i].Requires, id)
	}

	c.Gates = without(c.Gates, id)

	return nil
}

func without(list []string, id string) []string {
	var out []string
	for _, item := range list {
		if item != id {
			out = append(out, item)
		}
	}
	return out
}

func (c *Composition) Link(dependent, requirement string) error {
	if dependent == requirement {
		return ErrSelfLink
	}
	if !c.Has(dependent) {
		return fmt.Errorf("%w: %s", ErrNotOnCanvas, dependent)
	}
	if !c.Has(requirement) {
		return fmt.Errorf("%w: %s", ErrNotOnCanvas, requirement)
	}

	for i, a := range c.Artifacts {
		if a.ID != dependent {
			continue
		}
		if slices.Contains(a.Requires, requirement) {
			return nil
		}
		c.Artifacts[i].Requires = append(a.Requires, requirement)
		return nil
	}

	return nil
}

func (c *Composition) Unlink(dependent, requirement string) error {
	for i, a := range c.Artifacts {
		if a.ID == dependent {
			c.Artifacts[i].Requires = without(a.Requires, requirement)
			return nil
		}
	}

	return fmt.Errorf("%w: %s", ErrNotOnCanvas, dependent)
}

func (c *Composition) ToggleGate(id string) error {
	if !c.Has(id) {
		return fmt.Errorf("%w: %s", ErrNotOnCanvas, id)
	}

	if slices.Contains(c.Gates, id) {
		c.Gates = without(c.Gates, id)
		return nil
	}

	c.Gates = append(c.Gates, id)

	return nil
}

func (c *Composition) IsGate(id string) bool { return slices.Contains(c.Gates, id) }

func (c *Composition) SetTracks(path string) { c.Tracks = strings.TrimSpace(path) }

func (c *Composition) Rename(from, to string) error {
	to = strings.TrimSpace(to)

	if from == to {
		return nil
	}
	if to == "" {
		return errors.New("an artifact needs an id")
	}
	if !c.Has(from) {
		return fmt.Errorf("%w: %s", ErrNotOnCanvas, from)
	}
	if c.Has(to) {
		return fmt.Errorf("%w: %s", ErrCollision, to)
	}

	for i, a := range c.Artifacts {
		if a.ID == from {
			c.Artifacts[i].ID = to
		}
		for j, req := range c.Artifacts[i].Requires {
			if req == from {
				c.Artifacts[i].Requires[j] = to
			}
		}
	}

	for i, gate := range c.Gates {
		if gate == from {
			c.Gates[i] = to
		}
	}

	return nil
}

// Schema builds what the canvas describes, so that the one definition of a
// valid schema lives in internal/schema. A composer that agreed with itself and
// disagreed with the validator would let a user build something the detail view
// then calls broken.
func (c *Composition) Schema(name string) schema.Schema {
	artifacts := make([]schema.Artifact, 0, len(c.Artifacts))
	for _, a := range c.Artifacts {
		artifacts = append(artifacts, schema.Artifact{
			ID:          a.ID,
			Generates:   a.Generates,
			Template:    a.Template,
			Description: a.Description,
			Instruction: a.Instruction,
			Requires:    append([]string{}, a.Requires...),
		})
	}

	return schema.Schema{
		Name:        name,
		Version:     1,
		Description: "Composed with ossm",
		Artifacts:   artifacts,
		Apply: schema.Apply{
			Requires: append([]string{}, c.Gates...),
			Tracks:   c.Tracks,
		},
	}
}

func (c *Composition) Findings(name string) schema.Findings {
	if len(c.Artifacts) == 0 {
		return schema.Findings{{
			Severity: schema.Fatal,
			Message:  "the canvas is empty; add an artifact from the palette with a",
		}}
	}

	return schema.Validate(c.Schema(name))
}

func (c *Composition) Valid(name string) bool { return c.Findings(name).Valid() }

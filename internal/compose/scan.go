package compose

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"sync"
)

type Warning struct {
	Artifact  string
	File      string
	Reference string
	Message   string
}

func (w Warning) String() string {
	return fmt.Sprintf("%s: %s mentions %s, which is not on the canvas", w.Artifact, w.File, w.Reference)
}

// Scan looks through each canvas artifact's template and instruction for
// references to artifacts that are not on the canvas.
//
// Only ids and generated paths from the sources the user actually added are
// looked for. Scanning for arbitrary words would produce noise nobody reads,
// and the point of the warning is that it is worth acting on.
func (c *Composition) Scan() []Warning {
	wanted := c.absentReferences()
	if len(wanted) == 0 {
		return nil
	}

	var warnings []Warning
	reported := make(map[string]bool)

	for _, a := range c.Artifacts {
		text, file, err := a.templateText()
		if err != nil {
			key := "read:" + file
			if !reported[key] {
				reported[key] = true
				warnings = append(warnings, Warning{
					Artifact: a.ID,
					File:     file,
					Message:  "the template could not be read: " + err.Error(),
				})
			}
		}

		text += "\n" + a.Instruction

		for _, reference := range wanted {
			if !mentions(text, reference) {
				continue
			}

			key := a.ID + "\x00" + reference
			if reported[key] {
				continue
			}
			reported[key] = true

			warnings = append(warnings, Warning{
				Artifact:  a.ID,
				File:      file,
				Reference: reference,
			})
		}
	}

	return warnings
}

func (a Artifact) templateText() (string, string, error) {
	if strings.TrimSpace(a.Template) == "" {
		return "", "", nil
	}

	path := filepath.Join(a.SourceDir, "templates", filepath.FromSlash(a.Template))

	raw, err := os.ReadFile(path)
	if err != nil {
		return "", a.Template, err
	}

	return string(raw), a.Template, nil
}

func (c *Composition) absentReferences() []string {
	onCanvas := c.IDs()

	generated := make(map[string]bool, len(c.Artifacts))
	for _, a := range c.Artifacts {
		generated[a.Generates] = true
	}

	seen := make(map[string]bool)
	var references []string

	for _, source := range c.Sources {
		for _, a := range source.Artifacts {
			if !slices.Contains(onCanvas, a.ID) && !seen[a.ID] {
				seen[a.ID] = true
				references = append(references, a.ID)
			}
			if a.Generates != "" && !generated[a.Generates] && !seen[a.Generates] && !strings.Contains(a.Generates, "*") {
				seen[a.Generates] = true
				references = append(references, a.Generates)
			}
		}
	}

	sort.Strings(references)

	return references
}

var wordCache sync.Map

// mentions matches on word boundaries, so design does not match designer and
// tasks.md does not match subtasks.md.
//
// The compiled patterns are cached in a sync.Map rather than a plain one: a
// composition is scanned from a Bubble Tea update and from tests running in
// parallel, and a plain map written from two goroutines is a crash rather than
// a slow path.
func mentions(text, reference string) bool {
	cached, ok := wordCache.Load(reference)
	if !ok {
		cached = regexp.MustCompile(`(^|[^\w./-])` + regexp.QuoteMeta(reference) + `($|[^\w-])`)
		wordCache.Store(reference, cached)
	}

	return cached.(*regexp.Regexp).MatchString(text)
}

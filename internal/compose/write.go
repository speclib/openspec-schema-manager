package compose

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/speclib/openspec-schema-manager/internal/schema"
	"github.com/speclib/openspec-schema-manager/internal/source"
)

var ErrNotValid = errors.New("the composition is not valid")

func (c *Composition) Write(intoDir, name string) (string, error) {
	if err := source.ValidName(name); err != nil {
		return "", err
	}

	name = strings.TrimSpace(name)

	if findings := c.Findings(name); !findings.Valid() {
		return "", fmt.Errorf("%w: %s", ErrNotValid, findings.Fatal()[0].Message)
	}

	destination := filepath.Join(intoDir, name)
	if _, err := os.Stat(destination); err == nil {
		return "", fmt.Errorf("%w: %s", source.ErrDestinationBusy, destination)
	}

	if err := os.MkdirAll(intoDir, 0o755); err != nil {
		return "", fmt.Errorf("preparing %s: %w", intoDir, err)
	}

	staged, err := os.MkdirTemp(intoDir, ".compose-*")
	if err != nil {
		return "", fmt.Errorf("preparing %s: %w", intoDir, err)
	}
	defer os.RemoveAll(staged)

	content := filepath.Join(staged, "content")
	if err := os.MkdirAll(content, 0o755); err != nil {
		return "", err
	}

	templates, err := c.copyTemplates(content)
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(filepath.Join(content, schema.SchemaFile), []byte(c.render(name, templates)), 0o644); err != nil {
		return "", err
	}

	if err := os.Rename(content, destination); err != nil {
		return "", fmt.Errorf("writing %s: %w", destination, err)
	}

	return destination, nil
}

// copyTemplates copies each artifact's template and returns the path each
// artifact should declare.
//
// Two artifacts from different schemas can declare the same template path;
// tasks.md is not rare. On a collision the second is written under a path
// prefixed with its source schema's name, so nothing is silently lost.
func (c *Composition) copyTemplates(into string) (map[string]string, error) {
	paths := make(map[string]string, len(c.Artifacts))
	taken := make(map[string]string, len(c.Artifacts))

	for _, a := range c.Artifacts {
		if strings.TrimSpace(a.Template) == "" {
			continue
		}

		from := filepath.Join(a.SourceDir, "templates", filepath.FromSlash(a.Template))

		body, err := os.ReadFile(from)
		if err != nil {
			return nil, fmt.Errorf("copying the template for %s: %w", a.ID, err)
		}

		rel := a.Template
		if owner, clash := taken[rel]; clash && owner != a.SourceDir {
			rel = sanitise(a.SourceName) + "/" + a.Template
		}
		taken[rel] = a.SourceDir

		target := filepath.Join(into, "templates", filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(target, body, 0o644); err != nil {
			return nil, err
		}

		paths[a.ID] = rel
	}

	return paths, nil
}

func sanitise(name string) string {
	replacer := strings.NewReplacer("/", "-", `\`, "-", " ", "-")
	return replacer.Replace(strings.ToLower(strings.TrimSpace(name)))
}

func (c *Composition) render(name string, templates map[string]string) string {
	var b strings.Builder

	b.WriteString(c.provenance())

	b.WriteString("name: " + name + "\n")
	b.WriteString("version: 1\n")
	b.WriteString("description: Composed with ossm\n")
	b.WriteString("artifacts:\n")

	for _, a := range c.Artifacts {
		b.WriteString("  - id: " + quote(a.ID) + "\n")
		b.WriteString("    generates: " + quote(a.Generates) + "\n")
		b.WriteString("    description: " + quote(a.Description) + "\n")

		if template, ok := templates[a.ID]; ok {
			b.WriteString("    template: " + quote(template) + "\n")
		}

		if a.Instruction != "" {
			b.WriteString("    instruction: |\n")
			for _, line := range strings.Split(strings.TrimRight(a.Instruction, "\n"), "\n") {
				b.WriteString("      " + line + "\n")
			}
		}

		if len(a.Requires) > 0 {
			b.WriteString("    requires:\n")
			for _, req := range a.Requires {
				b.WriteString("      - " + quote(req) + "\n")
			}
		} else {
			b.WriteString("    requires: []\n")
		}
	}

	b.WriteString("apply:\n")
	b.WriteString("  requires:\n")
	for _, gate := range c.Gates {
		b.WriteString("    - " + quote(gate) + "\n")
	}
	b.WriteString("  tracks: " + quote(c.Tracks) + "\n")

	return b.String()
}

func (c *Composition) provenance() string {
	var b strings.Builder

	b.WriteString("# Composed with ossm from:\n")

	for _, source := range c.Sources {
		var mine []string
		for _, a := range c.Artifacts {
			if a.SourceDir != source.Dir {
				continue
			}
			if a.Renamed() {
				mine = append(mine, a.SourceID+" as "+a.ID)
			} else {
				mine = append(mine, a.ID)
			}
		}

		if len(mine) == 0 {
			continue
		}

		ref := source.Ref
		if ref == "" {
			ref = "default branch"
		}

		b.WriteString("#   " + source.Name + " @ " + ref + ": " + strings.Join(mine, ", ") + "\n")
		b.WriteString("#     " + source.Dir + "\n")
	}

	b.WriteString("#\n")

	return b.String()
}

func quote(s string) string {
	if s == "" {
		return `""`
	}

	if strings.ContainsAny(s, `:#{}[]&*!|>'"%@`+"`") || strings.HasPrefix(s, " ") || strings.HasSuffix(s, " ") {
		return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
	}

	return s
}

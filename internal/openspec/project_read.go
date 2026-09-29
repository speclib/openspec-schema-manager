package openspec

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

const UnknownSchema = "unknown"

type ProjectChange struct {
	Name           string
	Schema         string
	CompletedTasks int
	TotalTasks     int
}

type Project struct {
	Root     string
	Default  string
	Schemas  []ResolvedSchema
	Changes  []ProjectChange
	Problems []string
}

func (p Project) DefaultKnown() bool {
	return p.Default != "" && p.Default != UnknownSchema
}

func DefaultSchema(root string) string {
	raw, err := os.ReadFile(filepath.Join(root, "openspec", "config.yaml"))
	if err != nil {
		return UnknownSchema
	}

	var cfg struct {
		Schema string `yaml:"schema"`
	}
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return UnknownSchema
	}

	if strings.TrimSpace(cfg.Schema) == "" {
		return UnknownSchema
	}

	return strings.TrimSpace(cfg.Schema)
}

const schemaLookupConcurrency = 4

func ReadProject(ctx context.Context, cli CLI, root string) Project {
	p := Project{Root: root, Default: DefaultSchema(root)}

	if cli == nil {
		p.Problems = append(p.Problems, "no OpenSpec adapter is configured")
		return p
	}

	schemas, err := cli.Schemas(ctx, root)
	if err != nil {
		p.Problems = append(p.Problems, describeCLIError("the available schemas could not be listed", err))
	} else {
		p.Schemas = schemas
	}

	changes, err := cli.Changes(ctx, root)
	if err != nil {
		p.Problems = append(p.Problems, describeCLIError("the changes could not be listed", err))
		return p
	}

	p.Changes = make([]ProjectChange, len(changes))

	var wg sync.WaitGroup
	gate := make(chan struct{}, schemaLookupConcurrency)

	for i, c := range changes {
		p.Changes[i] = ProjectChange{
			Name:           c.Name,
			Schema:         UnknownSchema,
			CompletedTasks: c.CompletedTasks,
			TotalTasks:     c.TotalTasks,
		}

		wg.Add(1)
		go func(i int, name string) {
			defer wg.Done()

			gate <- struct{}{}
			defer func() { <-gate }()

			if schema, err := cli.ChangeSchema(ctx, root, name); err == nil && schema != "" {
				p.Changes[i].Schema = schema
			}
		}(i, c.Name)
	}

	wg.Wait()

	return p
}

func describeCLIError(what string, err error) string {
	if errors.Is(err, ErrNotInstalled) {
		return what + ": the openspec CLI is not on PATH"
	}
	return what + ": " + err.Error()
}

func (s ResolvedSchema) Shadowing() bool {
	return len(s.Shadows) > 0
}

func (s ResolvedSchema) OriginLabel() string {
	switch s.Source {
	case SourcePackage:
		return "built-in"
	case SourceProject:
		return "project"
	case SourceUser:
		return "user"
	default:
		return s.Source
	}
}

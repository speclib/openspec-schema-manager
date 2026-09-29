package source

import (
	"context"
	"fmt"

	"github.com/speclib/openspec-schema-manager/internal/graph"
	"github.com/speclib/openspec-schema-manager/internal/schema"
)

type Origin string

const (
	OriginRegistry Origin = "registry"
	OriginLocal    Origin = "local"
)

type Resolved struct {
	Dir      string
	Schema   schema.Schema
	Graph    graph.Graph
	Findings schema.Findings
	Metrics  graph.Metrics
	Origin   Origin
	Label    string
	Source   Source
}

func (r Resolved) Cyclic() bool {
	_, err := r.Graph.Order()
	return err != nil
}

func ResolveDir(dir string, label string) (Resolved, error) {
	s, err := schema.Load(dir)
	if err != nil {
		return Resolved{}, err
	}

	return assemble(dir, s, OriginLocal, label, Source{}), nil
}

func Resolve(ctx context.Context, f Fetcher, src Source, label string, refetch bool) (Resolved, error) {
	dir, err := f.Fetch(ctx, src, refetch)
	if err != nil {
		return Resolved{}, err
	}

	s, err := schema.Load(dir)
	if err != nil {
		return Resolved{}, fmt.Errorf("reading the fetched schema: %w", err)
	}

	return assemble(dir, s, OriginRegistry, label, src), nil
}

func assemble(dir string, s schema.Schema, origin Origin, label string, src Source) Resolved {
	g := graph.Build(s)

	metrics, err := g.Metrics()
	if err != nil {
		metrics = graph.Metrics{Artifacts: g.Len()}
	}

	return Resolved{
		Dir:      dir,
		Schema:   s,
		Graph:    g,
		Findings: schema.ValidateDir(s, dir),
		Metrics:  metrics,
		Origin:   origin,
		Label:    label,
		Source:   src,
	}
}

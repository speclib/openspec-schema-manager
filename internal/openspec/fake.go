package openspec

import "context"

type Fake struct {
	SchemasResult []ResolvedSchema
	SchemasErr    error
	SchemasCalls  []string
}

func (f *Fake) Schemas(_ context.Context, dir string) ([]ResolvedSchema, error) {
	f.SchemasCalls = append(f.SchemasCalls, dir)
	return f.SchemasResult, f.SchemasErr
}

package openspec

import "context"

type Fake struct {
	SchemasResult []ResolvedSchema
	SchemasErr    error
	SchemasCalls  []string

	ChangesResult []Change
	ChangesErr    error

	ChangeSchemas   map[string]string
	ChangeSchemaErr error

	Validations   map[string]Validation
	ValidateErr   error
	ValidateCalls []string
}

func (f *Fake) Schemas(_ context.Context, dir string) ([]ResolvedSchema, error) {
	f.SchemasCalls = append(f.SchemasCalls, dir)
	return f.SchemasResult, f.SchemasErr
}

func (f *Fake) Changes(context.Context, string) ([]Change, error) {
	return f.ChangesResult, f.ChangesErr
}

func (f *Fake) ChangeSchema(_ context.Context, _, change string) (string, error) {
	if f.ChangeSchemaErr != nil {
		return "", f.ChangeSchemaErr
	}
	return f.ChangeSchemas[change], nil
}

func (f *Fake) ValidateSchema(_ context.Context, _, name string) (Validation, error) {
	f.ValidateCalls = append(f.ValidateCalls, name)

	if f.ValidateErr != nil {
		return Validation{}, f.ValidateErr
	}

	if v, ok := f.Validations[name]; ok {
		return v, nil
	}

	return Validation{Name: name, Valid: true}, nil
}

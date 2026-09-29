// Package schema models an OpenSpec workflow schema: schema.yaml, its
// artifacts and its apply gate.
//
// It parses, validates and writes a schema, and reports what is wrong with an
// invalid one: an unknown artifact in requires, a cycle, a missing template
// file, an unknown artifact in apply.requires.
//
// This package is pure. It takes bytes and paths, not network connections, and
// is fully unit tested. Graph questions belong to internal/graph.
package schema

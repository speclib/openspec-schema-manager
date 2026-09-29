// Package registry reads the published registry of OpenSpec workflow schemas.
//
// It fetches openspec-schemas.json, caches it, honours a time to live, filters
// it, and merges in the schemas OpenSpec reports as built in. An entry is thin
// by design: it says where a schema lives, not what it contains.
//
// Entries are data written by third parties. This package parses and validates
// them; it never acts on their contents. Fetching what an entry points at is
// internal/source's job.
package registry

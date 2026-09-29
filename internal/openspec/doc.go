// Package openspec is the adapter over the openspec command line tool.
//
// OpenSpec owns project state. ossm keeps no install lockfile and no record of
// what is installed; it asks OpenSpec, preferring --json output, and reads the
// project's own files. Everything that shells out to openspec lives here,
// behind an interface with a fake for tests.
//
// OpenSpec 1.10.0 has no install-from-source command, so installing a schema is
// implemented here as fetch, write and validate. NOTES.md records why. When
// OpenSpec grows a command, only this package changes.
package openspec

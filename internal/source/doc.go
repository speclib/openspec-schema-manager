// Package source fetches a schema folder from the repository, path and ref an
// entry names, and caches it keyed by that triple.
//
// A shallow git clone is preferred; fetching raw files is the fallback for a
// source that is not a git repository. When an entry names no ref, the
// repository's default branch is resolved rather than assumed, because it is
// not always main.
//
// The fetched bytes are untrusted. This package writes them to the cache and
// nothing else. Parsing them is internal/schema's job.
package source

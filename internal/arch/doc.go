// Package arch holds the tests that enforce the project's structure rather
// than its behaviour.
//
// It contains no production code. Its tests ask the Go toolchain what the
// build graph actually looks like, so a rule stated in CLAUDE.md is checked by
// the same view of imports the compiler has.
package arch

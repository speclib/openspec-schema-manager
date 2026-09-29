// Package compose holds the composer's state: artifacts drawn from several
// source schemas, the edges between them, and the identifier remapping that
// resolves a collision when two sources use the same artifact id.
//
// It validates continuously (no cycles, every requires resolves, at least one
// apply gate, a tracked file set, no two artifacts generating the same path)
// and scans template text for references to artifacts that are not on the
// canvas.
//
// Composition is build time: the result is one standalone schema. Runtime
// composition does not exist in OpenSpec. This package is pure; writing the
// result to disk is its only side effect and is a separate call.
package compose

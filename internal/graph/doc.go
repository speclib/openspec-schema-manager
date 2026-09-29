// Package graph turns a schema's artifacts and their requires edges into a
// directed acyclic graph and answers questions about it.
//
// It builds the graph, detects cycles, derives the figures that make two
// schemas comparable (artifact count, longest dependency chain, gates before
// apply), and emits a Mermaid flowchart.
//
// The Mermaid output lives here rather than in the TUI so that a web front can
// reuse it. This package is pure and fully unit tested.
package graph

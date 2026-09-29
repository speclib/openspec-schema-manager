---
# openspec-schema-manager-zjwg
title: Mermaid flowchart output
status: completed
type: epic
priority: normal
created_at: 2026-09-29T18:01:15Z
updated_at: 2026-09-29T19:16:59Z
parent: openspec-schema-manager-xzx4
---

Emit a graph TD flowchart from the artifact DAG, marking apply gate nodes distinctly, in its own package so a web front can reuse it.

## Summary of Changes

`Graph.Mermaid` emits a `graph TD` flowchart, in the graph's stable order, with
apply gates as hexagons and everything else as rectangles. Two shapes rather
than a class definition, because the ASCII renderer in milestone 04 reads shapes
and ignores styling.

An artifact id is a string from a third party, and Mermaid node identifiers
cannot hold spaces, quotation marks or brackets. An id matching
`^[A-Za-z][A-Za-z0-9_-]*$` is used directly; anything else gets `n<index>` and
keeps its own text as an escaped, quoted label. The index is the artifact's
position, so two awkward ids cannot collide. A fixture carries ids holding a
space, a quotation mark, a bracket and a semicolon, and a test asserts a newline
in an id does not add a line to the diagram.

Output is byte-identical across runs, asserted twenty times over.

It lives in `internal/graph` rather than in the TUI so the planned web front can
reuse it, which is the only reason the graph and the renderer are separate
packages at all.

openspec-link: openspec/changes/archive/2026-09-29-model-the-schema-and-its-graph

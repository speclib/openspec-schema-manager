---
# openspec-schema-manager-rxbk
title: Artifact DAG and comparison metrics
status: completed
type: epic
priority: normal
created_at: 2026-09-29T18:01:15Z
updated_at: 2026-09-29T19:16:59Z
parent: openspec-schema-manager-xzx4
---

Build the graph from artifacts and requires edges, and derive the figures that make two schemas comparable: artifact count, longest dependency chain, gates before apply.

## Summary of Changes

`internal/graph` builds one node per artifact and one edge per `requires` entry,
directed from the requirement to the artifact that needs it, so an edge points
the way work flows.

The order is Kahn's algorithm with the ready set resolved in declaration order.
A map-based topological sort gives a different order per run, and the diagram
would then move between runs for no reason a user could see. Declaration order
is the tiebreaker because it is the order the schema's author chose. A test runs
the sort twenty times and asserts the result never moves.

The figures are artifact count, longest chain, gate count, and the number of
artifacts nothing requires. The longest chain counts nodes rather than edges, so
one artifact is a chain of 1; a chain of 0 would mean an empty schema, which
validation already rejects. Over a branchy schema it takes the longest path, not
the artifact count, which is the whole reason it is worth reporting.

An unknown requirement builds no edge and does not block ordering. Validation
already reports it as fatal, and a graph that refuses to be built cannot show
the user what is wrong with it.

100% covered, which the gate now holds as the floor.

openspec-link: openspec/changes/archive/2026-09-29-model-the-schema-and-its-graph

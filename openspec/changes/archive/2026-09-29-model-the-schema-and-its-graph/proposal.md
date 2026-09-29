# Model the schema and its graph

Beans epics: `openspec-schema-manager-4pq2` (Schema model, parse and validate),
`openspec-schema-manager-rxbk` (Artifact DAG and comparison metrics),
`openspec-schema-manager-zjwg` (Mermaid flowchart output).

## Why

Everything ossm does past listing a schema needs to understand one. The detail
view shows the artifacts and the apply gate. The diagram draws the dependency
edges. The composer edits them, and refuses to write a schema with a cycle in
it. All three want the same answer to the same question, and inventing it three
times is how they end up disagreeing about what a valid schema is.

Comparing schemas is the point of ossm, and a list of names does not let anyone
compare. Artifact count, the longest chain and how many gates stand before apply
are what turn "this one looks heavier" into a number.

## What Changes

- ossm parses a `schema.yaml` into a model carrying the schema's name, version,
  description, artifacts and apply gate.
- An artifact carries its id, what it generates, its template, its optional
  description and instruction, and the artifact ids it requires.
- ossm reports what is wrong with an invalid schema rather than the first
  problem it meets: an unknown id in `requires`, a cycle, a duplicate id, an
  unknown id in `apply.requires`, an empty artifact list, a missing template
  file, two artifacts generating the same path.
- Validation separates what makes a schema unusable from what makes it
  suspicious. A cycle is fatal. A template file that is missing on disk is fatal
  only when there is a disk to look at, because a schema read from a registry
  entry has no files yet.
- The artifacts and their `requires` edges become a directed acyclic graph, with
  a stable topological order so two runs never draw the same schema differently.
- Comparison figures come off that graph: artifact count, the longest dependency
  chain, how many artifacts gate apply, and how many artifacts nothing depends
  on.
- The graph emits a Mermaid `graph TD` flowchart, with apply-gate nodes marked
  distinctly, in its own package so the planned web front can reuse it.

## Capabilities

### New Capabilities

- `schema-model`: what a schema is, what an artifact carries, and what the apply
  block means.
- `schema-validation`: what makes a schema invalid, what makes it merely
  suspicious, and how both are reported.
- `artifact-graph`: the graph built from artifacts and `requires`, its ordering,
  and the figures derived from it.
- `mermaid-output`: the flowchart emitted from the graph.

### Modified Capabilities

None.

## Impact

- New: `internal/schema` and `internal/graph`, both pure and both fully unit
  tested, plus fixture schemas covering a straight chain, a branchy graph, a
  cycle and several malformed cases.
- Nothing renders a diagram yet. Milestone 04 turns the Mermaid text into
  boxes; keeping the two apart is what lets the web front reuse this half.
- The model is read from the schemas OpenSpec actually ships and from the
  fixtures in the briefing, not from the briefing's prose. Where they disagree
  the real schemas win and the difference goes in `NOTES.md`.
- A schema is a third party's file. It is parsed and rendered, never executed,
  and an `instruction` block is text like any other.

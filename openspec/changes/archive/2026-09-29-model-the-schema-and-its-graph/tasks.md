# Tasks

## 1. Fixtures

- [x] 1.1 Fixture schemas under `internal/schema/testdata`: a straight chain, a
      branchy graph with a join, a single artifact, a cycle, a self-reference,
      a duplicate id, an unknown requirement, an unknown apply gate, two
      artifacts generating the same path, and an orphan artifact
- [x] 1.2 Copy the real `spec-driven` schema as a fixture, so the model is
      tested against what OpenSpec actually ships
- [x] 1.3 A fixture folder with `templates/` present and one with a template
      missing

## 2. The model and parsing

- [x] 2.1 Define `Schema`, `Artifact` and `Apply`
- [x] 2.2 Parse with `KnownFields(true)`; an unknown key is an error naming it
- [x] 2.3 `requires` absent and `requires: []` both give no requirements
- [x] 2.4 `generates` keeps its glob as written
- [x] 2.5 `name` keeps upstream casing
- [x] 2.6 An `instruction` is carried as text and used for nothing else
- [x] 2.7 Test against every fixture including the real `spec-driven`
- [x] 2.8 `go test ./...` passes

## 3. Validation

- [x] 3.1 `Finding` with a severity, an artifact and a message
- [x] 3.2 Fatal: empty artifact list, duplicate id, empty id, unknown
      requirement, self-reference, cycle, unknown apply gate, empty
      `apply.requires`, absent `apply.tracks`
- [x] 3.3 Warning: two artifacts generating the same path, an artifact with no
      template, an artifact nothing requires that does not gate apply
- [x] 3.4 Every finding names the artifact or the schema and states the rule
- [x] 3.5 All problems are reported in one pass, not just the first
- [x] 3.6 A cycle names its members in the order they connect
- [x] 3.7 `ValidateDir` also checks template files exist; `Validate` does not
- [x] 3.8 Table test over every fixture and every rule
- [x] 3.9 `go test ./...` passes

## 4. The graph

- [x] 4.1 Build nodes and edges from artifacts and `requires`, directed from
      requirement to dependent
- [x] 4.2 Topological order by Kahn's algorithm, ties settled by declaration
      order
- [x] 4.3 Ordering a cyclic graph reports the cycle rather than a partial order
- [x] 4.4 Report which nodes gate apply
- [x] 4.5 Test the order is identical across runs and that the tiebreaker is
      declaration order
- [x] 4.6 `go test ./...` passes

## 5. Figures

- [x] 5.1 Artifact count, longest chain, gate count, count of artifacts nothing
      requires
- [x] 5.2 Longest chain counts nodes, so one artifact gives 1
- [x] 5.3 Longest chain over a branchy graph takes the longest path
- [x] 5.4 Table test over the chain, branch, single and wide fixtures
- [x] 5.5 `go test ./...` passes

## 6. Mermaid

- [x] 6.1 Emit `graph TD` with a node per artifact and an edge per requirement
- [x] 6.2 Gate nodes use the hexagon shape, plain nodes the rectangle
- [x] 6.3 An id that is a valid Mermaid identifier is used directly
- [x] 6.4 Any other id gets a generated identifier and keeps its text as an
      escaped, quoted label
- [x] 6.5 Two ids that would collide get distinct identifiers
- [x] 6.6 Output is byte-identical across runs and follows the stable order
- [x] 6.7 Table test over every fixture, including ids holding a space, a quote,
      a bracket and a semicolon
- [x] 6.8 `go test ./...` passes

## 7. Close out

- [x] 7.1 Record in `NOTES.md` any difference between the briefing's account of
      `schema.yaml` and what the real schemas declare
- [x] 7.2 Raise the coverage floors for `internal/schema` and `internal/graph`
      to their measured values, both at or above the briefing's 80% for core
- [x] 7.3 `bash scripts/coverage-gate.sh` and `nix flake check` pass
- [x] 7.4 Fill in the three epics' summaries and mark them completed; close
      milestone 03
- [x] 7.5 Archive the change, commit and push

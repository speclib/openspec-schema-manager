---
# openspec-schema-manager-xzx4
title: 03 Schema model and graph
status: completed
type: milestone
priority: normal
created_at: 2026-09-29T18:00:33Z
updated_at: 2026-09-29T19:16:57Z
---

The pure core: parse and validate schema.yaml, build the artifact DAG, compute comparison metrics, and emit Mermaid. No IO, fully unit tested.

## Summary of Changes

Milestone 03 is one OpenSpec change, `model-the-schema-and-its-graph`, adding
four capabilities: `schema-model`, `schema-validation`, `artifact-graph` and
`mermaid-output`.

Both packages are pure and carry the project's highest coverage floors: 100% for
`internal/graph` and 99% for `internal/schema`, against the briefing's 80% for
core packages.

The milestone also turned up a real bug in milestone 02's wiring, found by the
nix sandbox rather than by a local run. `Init` batched the cache read and the
fetch, so a fast fetch could be overwritten by the stale read that started
before it. The cached registry is now applied in `New`, and `Init` asks only for
the refresh. Two tests hold it: one that the cache is drawn before `Init` runs,
one that a refresh survives.

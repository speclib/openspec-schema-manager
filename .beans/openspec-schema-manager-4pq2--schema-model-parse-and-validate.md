---
# openspec-schema-manager-4pq2
title: Schema model, parse and validate
status: completed
type: epic
priority: normal
created_at: 2026-09-29T18:01:15Z
updated_at: 2026-09-29T19:16:59Z
parent: openspec-schema-manager-xzx4
---

The schema.yaml model and its validation: unknown artifact in requires, cycles, missing template file, unknown apply.requires. Pure, no IO.

## Summary of Changes

`internal/schema` parses `schema.yaml` into `Schema`, `Artifact` and `Apply`,
and reports every problem in one pass rather than the first.

Parsing is strict, the opposite of the registry parser. A schema is written by
the user or fetched from a repository they chose, so a key ossm silently ignores
there is a workflow step that quietly does not happen. That is worse than a
failed parse. `NOTES.md` records the pairing.

Validation separates fatal from warning, because the two have different
consequences. A cycle means the workflow cannot run. Two artifacts writing the
same path means one clobbers the other, which is usually a mistake and
occasionally deliberate; refusing to open the schema over it would make ossm
less useful than a text editor.

A cycle names its members in the order they connect, from a depth-first walk
with a colour per node. "tasks → design → tasks" is a fixable message; "this
schema has a cycle" is not. A self-reference is reported as its own kind, because
that is what the reader is looking at.

The template-file check needs a directory and runs only when given one. A
registry entry has no files until it is fetched, and a validator that reports
every template as missing teaches people to ignore it.

Tested against the `spec-driven` schema OpenSpec 1.10.0 actually ships, not only
against fixtures. That is how `apply.instruction` was found: the briefing's
account of `schema.yaml` omits it. Recorded in `NOTES.md`.

openspec-link: openspec/changes/archive/2026-09-29-model-the-schema-and-its-graph

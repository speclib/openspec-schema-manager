---
# openspec-schema-manager-w1v2
title: Composer validation and template warnings
status: completed
type: epic
priority: normal
created_at: 2026-09-29T18:01:15Z
updated_at: 2026-09-29T20:42:07Z
parent: openspec-schema-manager-plwt
---

Validate continuously against cycles, unresolved requires, a missing apply gate, an unset tracked file and duplicate generated paths, and warn when a template mentions an artifact or generated file that is not on the canvas.

## Summary of Changes

Validation runs after every edit, through `schema.Validate` on the schema the
canvas describes, plus the one thing only a composition can be wrong about: an
empty canvas. A cycle is reported the moment the link closing it is made.

Template scanning looks through each canvas artifact's template and instruction
for ids and generated paths of source artifacts that are not on the canvas.
Matching is on word boundaries, so `design` does not match `designer` and
`tasks.md` does not match `subtasks.md`. Only ids and paths from the sources the
user actually added are looked for; scanning for arbitrary words would produce
noise nobody reads, and a warning nobody reads is worse than none.

Globs are not scanned for, because `specs/**/*.md` appearing in a template says
nothing about whether anything produces it.

The warnings are advisory and a composition with warnings can still be written.
A template mentioning an artifact the user deliberately left out is their
decision, and a composer that refuses over it is one people work around by
editing the template to lie.

One concurrency fix came out of testing: the compiled pattern cache was a plain
map written from a Bubble Tea update and from parallel tests, which is a crash
rather than a slow path. It is a `sync.Map`, and the suite runs under `-race`.

openspec-link: openspec/changes/archive/2026-09-29-compose-a-schema

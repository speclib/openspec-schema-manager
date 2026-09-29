---
# openspec-schema-manager-jbe0
title: Composer model and canvas
status: completed
type: epic
priority: normal
created_at: 2026-09-29T18:01:15Z
updated_at: 2026-09-29T20:42:07Z
parent: openspec-schema-manager-plwt
---

The composition state: artifacts picked from several sources, their edges, and the id remapping that resolves collisions on add. Pure and unit tested.

## Summary of Changes

`internal/compose` holds the canvas: artifacts copied from source schemas, their
edges, the gates and the tracked file. Each artifact carries its provenance
rather than pointing at a source object, so a draft is a plain JSON round trip
and resuming needs nothing but the file.

Two decisions worth keeping.

An artifact arrives without the requirements that do not resolve. It does not
pull its dependencies in behind it, and adding one later does not restore an
edge that was dropped. The user chose what to link, which is the whole point of
a composer; a tool that keeps adding things nobody asked for is worse than one
that adds too little, and a missing edge is at least visible on the canvas.

A collision is reported rather than renamed for you. Two schemas both declaring
`tasks` is common, and which one keeps the name is a decision about the workflow
being built. Renaming to `tasks-2` automatically is the kind of help that has to
be undone.

`Composition.Schema` builds what the canvas describes and validation runs
`schema.Validate` on it rather than reimplementing anything. A test asserts the
two agree. A composer that disagreed with the validator would let a user build
something the detail view then calls broken.

openspec-link: openspec/changes/archive/2026-09-29-compose-a-schema

---
# openspec-schema-manager-hkub
title: Keyboard edge editing
status: completed
type: epic
priority: normal
created_at: 2026-09-29T18:01:15Z
updated_at: 2026-09-29T20:42:07Z
parent: openspec-schema-manager-plwt
---

The editing surface: add, remove, link, unlink, toggle apply gate, set the tracked file and rename an artifact id so every edge that references it follows.

## Summary of Changes

`a` add, `x` remove, `l` link, `u` unlink, `g` gate, `t` tracks, `R` rename,
`d` diagram, `w` write, `s` save, `o` open a draft. Left and right move between
the palette and the canvas, because `tab` belongs to the frame.

Removing takes every edge naming the artifact in either direction and its gate.
Renaming updates every requirement and the gate, and a rename to an id in use is
refused. Renaming to the id an artifact already has is not a collision, which is
the case a naive check gets wrong.

Every operation that needs text is its own mode with its own prompt, and
`Capturing()` reports true in all of them, so the frame keeps its hands off.
That is six modes, and the reason they are separate states rather than flags is
that each has different keys and the help overlay reports them.

`p` on a schema in the Registry, Project or Local tab sends it to the palette
without leaving the tab, resolved the same way opening its detail is: fetched if
it is a registry entry, read from disk if it is not.

openspec-link: openspec/changes/archive/2026-09-29-compose-a-schema

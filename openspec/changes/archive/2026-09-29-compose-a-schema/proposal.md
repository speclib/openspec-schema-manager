# Compose a schema

Beans epics: `openspec-schema-manager-jbe0` (Composer model and canvas),
`openspec-schema-manager-hkub` (Keyboard edge editing),
`openspec-schema-manager-w1v2` (Composer validation and template warnings),
`openspec-schema-manager-oyf0` (Write a composed schema and drafts).

## Why

Browsing and duplicating both assume one schema is nearly right. Often none is.
Someone wants the research step from one, the review gate from another and the
rest from a third, and duplicating gives them a starting point that is three
quarters wrong.

The artifact graph is the thing ossm already understands. Everything needed to
assemble one from parts is already in place: a parsed model, a graph with a
stable order, validation that reports everything at once, and a renderer. What
is missing is a surface for choosing artifacts and drawing edges between them.

## What Changes

- The Composer tab holds a palette and a canvas. The palette lists the artifacts
  of any schemas the user has added as sources, grouped by schema. The canvas
  holds the artifacts picked so far and the edges between them.
- Sources are added from the Registry, Project or Local tabs with `p`, or from
  the Composer by opening a schema.
- `a` adds the palette selection to the canvas. An id already on the canvas is
  a collision, and the user is asked for a new one before it is added.
- `x` removes the selected artifact and every edge touching it.
- `l` starts a link from the selected artifact and picks a target; `enter`
  confirms, adding a `requires` edge. `u` removes an edge.
- `g` toggles the selected artifact as an apply gate. `t` sets the tracked file.
- `R` renames an artifact, updating every edge and every gate that names it.
- Validation runs after every change, and shows what is wrong: a cycle, an
  unresolved requirement, no apply gate, no tracked file, two artifacts writing
  the same path.
- After each add, the artifact's template and instruction text is scanned for
  references to artifacts and generated files that are not on the canvas, and
  each is listed as a warning with the file to edit.
- `w` writes the canvas as a new schema folder into a local schemas directory:
  a generated `schema.yaml` and each template copied from its source, with a
  comment block recording which schema and ref each artifact came from.
- `s` saves the composition as a draft in the state directory, and a draft can
  be resumed.

## Capabilities

### New Capabilities

- `composer-model`: what a composition is, how an artifact is added, and how a
  collision is resolved.
- `composer-editing`: the keyboard operations on artifacts and edges.
- `composer-validation`: what is checked continuously, and the template
  reference warnings.
- `composer-write`: what is written, where, and what the written schema records
  about where it came from.
- `composer-drafts`: saving and resuming a composition.

### Modified Capabilities

- `app-frame`: the Composer tab stops being a placeholder, and a schema can be
  added to the composer from other tabs.

## Impact

- New: `internal/compose`, pure and fully unit tested, and the composer screen.
- The composer writes a schema folder into a local schemas directory, which the
  Local tab then lists. It never writes into a project: composing produces a
  schema, and installing it is the existing, separate act.
- Template text is scanned, never executed, and the warnings are advisory. A
  composition with warnings can still be written, because a template mentioning
  an artifact the user deliberately left out is their decision to make.
- The written `schema.yaml` is generated rather than stitched, so the comment
  block recording provenance is the only comment in it. Templates are copied
  byte for byte.

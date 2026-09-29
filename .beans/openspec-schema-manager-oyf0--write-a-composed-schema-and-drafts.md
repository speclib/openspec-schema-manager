---
# openspec-schema-manager-oyf0
title: Write a composed schema and drafts
status: completed
type: epic
priority: normal
created_at: 2026-09-29T18:01:15Z
updated_at: 2026-09-29T20:42:07Z
parent: openspec-schema-manager-plwt
---

Write the canvas as a new schema folder with a generated schema.yaml recording which source and ref each artifact came from, and save or resume a composition as a draft.

## Summary of Changes

`w` writes the canvas as a schema folder into a local schemas directory, and an
end to end test copies the result into a real OpenSpec project, watches
`openspec schema validate` accept it, and creates a change with it.

`schema.yaml` is generated rather than stitched. A composed schema has no single
source document to preserve comments from, and generating it is the only way the
artifact order and the requirements are exactly what the canvas holds. The
provenance block is therefore the only comment in the file, which makes it easy
to find: it names each source schema, its ref and which artifacts came from it,
recording `tasks as review` when an artifact was renamed.

Templates are copied byte for byte. Two artifacts from different schemas can
declare the same template path, and `tasks.md` is not rare, so on a collision
the second is written under a path prefixed with its source schema's name and
that artifact's `template` names it. Otherwise a schema would silently lose one
of its templates.

Drafts are one JSON file each under the state directory. A corrupt draft is
reported and skipped rather than hiding the rest. Resuming reports artifacts
whose source directory has gone and leaves the composition editable: the user
can remove those and write the rest, which beats refusing to open it.

openspec-link: openspec/changes/archive/2026-09-29-compose-a-schema

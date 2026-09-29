---
# openspec-schema-manager-iy57
title: Acceptance end to end suite
status: completed
type: epic
priority: normal
created_at: 2026-09-29T18:01:15Z
updated_at: 2026-09-29T21:00:12Z
parent: openspec-schema-manager-gypt
---

Assert every acceptance criterion in the briefing against the built binary: browse outside a project, project view inside one, offline detail, install seen by OpenSpec, duplicate and edit, and a composed schema OpenSpec accepts.

## Summary of Changes

`test/e2e/acceptance_test.go` holds one case per criterion in the briefing's
section 8, named after the criterion and quoting it, all driving the built
binary.

It is a separate file from `frame_test.go` because its audience is different:
someone checking whether the proof of concept did what it promised should be
able to read one file.

The offline criterion is the one worth describing. It opens a registry schema,
quits, removes both the source repository and the registry file, starts ossm
again, and asserts the schema still opens with its artifacts and draws its
diagram. Nothing is stubbed; the files are gone.

The composing criterion builds a schema from two fixtures, watches a template
warning appear for an artifact deliberately left off the canvas, makes a cycle
and watches it reported, undoes it, writes the result, and then copies it into a
real OpenSpec project where `openspec schema validate` accepts it and
`openspec new change --schema` uses it.

Where a criterion needs `openspec` or `git`, the case skips with the reason. A
criterion that silently passes because its precondition was missing is worse
than one that visibly did not run.

openspec-link: openspec/changes/archive/2026-09-29-finish-the-proof-of-concept

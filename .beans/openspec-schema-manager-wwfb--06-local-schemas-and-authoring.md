---
# openspec-schema-manager-wwfb
title: 06 Local schemas and authoring
status: completed
type: milestone
priority: normal
created_at: 2026-09-29T18:00:33Z
updated_at: 2026-09-29T20:21:14Z
---

Schemas that are not in the registry: configured directories, a path prompt, recents, duplicate, a file tree, and an EDITOR round trip that revalidates on return.

## Summary of Changes

Milestone 06 is one OpenSpec change, `work-with-local-schemas`, adding
`local-schemas`, `path-prompt`, `duplicate` and `file-tree`, and extending
`app-frame` with the path prompt.

The Local tab replaced the last placeholder but one. A user can now find a
schema that nearly fits, copy it, and edit it without leaving ossm, which is
what makes the browsing worth doing.

The milestone's finding is that `openspec schema fork` cannot serve here: it
takes a resolvable name rather than a path, and it fails with EACCES on any
read-only source, which is every schema OpenSpec ships on a Nix install.
`NOTES.md` carries both, with the commands that show them.

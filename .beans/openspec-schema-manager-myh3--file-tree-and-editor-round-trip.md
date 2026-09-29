---
# openspec-schema-manager-myh3
title: File tree and editor round trip
status: completed
type: epic
priority: normal
created_at: 2026-09-29T18:01:15Z
updated_at: 2026-09-29T20:21:15Z
parent: openspec-schema-manager-wwfb
---

Browse a local schema's files, open the selected one in VISUAL, EDITOR or vi by suspending the TUI, and revalidate on return with errors shown inline.

## Summary of Changes

`t` shows a schema's files as a tree, `e` opens the selected one in `$VISUAL`,
`$EDITOR` or `vi` through `tea.ExecProcess`, and the schema is re-read and
re-validated when the editor exits.

The editor value is split on spaces so `EDITOR="code --wait"` works. No shell is
involved: passing it through `sh -c` would let a crafted value do more than
edit, and buys only quoting no real editor setting needs.

The tree lists `schema.yaml` first and then `templates/` in path order, the same
way every time, so the cursor does not land on a different file between runs.

The end to end case is the one that proves it: a stub editor appends a line to a
template, and the test waits on the file rather than on the screen, because
nothing on screen changes the moment an external process exits. Before the edit
the tree reports a missing template as fatal; after it, the schema validates.

openspec-link: openspec/changes/archive/2026-09-29-work-with-local-schemas

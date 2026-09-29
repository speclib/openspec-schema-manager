---
# openspec-schema-manager-t4t5
title: Duplicate a schema
status: completed
type: epic
priority: normal
created_at: 2026-09-29T18:01:15Z
updated_at: 2026-09-29T20:21:15Z
parent: openspec-schema-manager-wwfb
---

Copy any schema into a chosen local directory under a new name, rewriting name in schema.yaml, preferring OpenSpec's own fork command where it fits.

## Summary of Changes

`c` copies the selected schema into the first configured directory under a new
name. ossm does the copying itself; `openspec schema fork` was tested and ruled
out for two reasons, both recorded in `NOTES.md`.

Fork takes a schema name OpenSpec can already resolve, not a path, so a registry
schema that has not been installed and a folder opened by path are both
invisible to it. Those are the two cases duplication exists for.

And fork copies the source preserving its mode, then reopens the copy to rewrite
the name. A schema in the Nix store is mode 444, so forking any built-in schema
fails with EACCES on every Nix install. ossm writes every copied file mode 644
instead, and a test duplicates a deliberately read-only source and asserts the
copy is writable.

Only the `name:` line is rewritten, as text. A schema carries comments,
instruction blocks and a deliberate artifact order; a YAML round trip loses the
first and can reorder the rest. A test duplicates a schema with all three and
asserts they survive.

A name is refused when it is empty, holds a separator, or is `.` or `..`,
because the name becomes a directory. An occupied destination is refused without
writing anything.

openspec-link: openspec/changes/archive/2026-09-29-work-with-local-schemas

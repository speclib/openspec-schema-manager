---
# openspec-schema-manager-wl35
title: Project view
status: completed
type: epic
priority: normal
created_at: 2026-09-29T18:01:15Z
updated_at: 2026-09-29T20:02:24Z
parent: openspec-schema-manager-vv9r
---

The project standing in the working directory: its default schema, every schema available to it with where it resolves from, and each change with the schema it uses, all sourced from OpenSpec rather than from a parallel record.

## Summary of Changes

The Project tab shows the root, the default schema, every schema with where it
resolves from and what it shadows, and every change with its schema and task
progress. `r` re-reads it, `enter` opens a schema's detail from its resolved
path, and `s` sets the project default.

Outside a project the tab says so, says browsing works anyway, and says `--path`
points at one. That replaced the milestone 01 placeholder.

Setting the default rewrites only the `schema:` line, so the commented guidance
`openspec init` writes into `config.yaml` survives. A test asserts a comment is
still there afterwards.

The screen holds the same detail screen the Registry screen holds, so `enter`
opens a schema from either tab with no second implementation.

One layout fix came from the nix sandbox rather than a local run: the status
message sat in the header chain after the project root, and a long root
truncated it away. It has its own line now. The sandbox's shorter build path was
what made the truncation visible.

openspec-link: openspec/changes/archive/2026-09-29-stand-in-a-project-and-install

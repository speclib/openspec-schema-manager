---
# openspec-schema-manager-8ff6
title: 04 Detail view and diagram
status: completed
type: milestone
priority: normal
created_at: 2026-09-29T18:00:33Z
updated_at: 2026-09-29T19:34:43Z
---

The schema detail screen: header, artifact table, apply gate, metrics, lazily fetched source, and an ASCII flow diagram in a scrollable viewport.

## Summary of Changes

Milestone 04 is one OpenSpec change, `show-a-schema-in-detail`, adding
`source-fetch`, `schema-detail` and `diagram-view`, and extending
`schema-listing` with opening a row.

A schema can now be opened from the list, fetched from its repository at the ref
its entry names, read offline on every later open, and seen as a diagram.

The milestone closed open question 6: mermaid-ascii exports an importable
renderer, so ossm draws in process with no binary shipped.

It also settled what is not built. There is no raw-file fallback for a source
that is not a git repository, and `NOTES.md` says why rather than leaving the
gap to be discovered.

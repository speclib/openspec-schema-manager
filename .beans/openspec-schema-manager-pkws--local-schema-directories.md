---
# openspec-schema-manager-pkws
title: Local schema directories
status: completed
type: epic
priority: normal
created_at: 2026-09-29T18:01:15Z
updated_at: 2026-09-29T20:21:15Z
parent: openspec-schema-manager-wwfb
---

Scan the configured schemas_dirs for folders holding a schema.yaml and present them in their own section.

## Summary of Changes

`source.ScanDir` looks one level inside each configured directory for a folder
holding a `schema.yaml`. Not recursive: `schemas_dirs` points at a directory of
schemas, and walking a home directory looking for `schema.yaml` is how a tool
earns a reputation for being slow.

A folder whose `schema.yaml` does not parse is listed and marked unreadable, not
hidden. It is exactly the folder the user wants to open and fix, and hiding it
makes ossm useless at the moment it would be most useful.

A configured directory that does not exist is reported and the others are still
scanned. A directory holding no schemas is reported as holding none rather than
disappearing, so a wrong path in the configuration is visible.

A schema declaring no name falls back to its folder name, so a half-written
schema still has something to select.

openspec-link: openspec/changes/archive/2026-09-29-work-with-local-schemas

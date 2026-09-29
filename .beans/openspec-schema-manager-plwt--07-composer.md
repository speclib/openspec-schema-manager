---
# openspec-schema-manager-plwt
title: 07 Composer
status: completed
type: milestone
priority: normal
created_at: 2026-09-29T18:00:33Z
updated_at: 2026-09-29T20:42:06Z
---

Build a standalone schema from artifacts drawn from several sources: palette, canvas, keyboard edge editing, continuous validation, template reference warnings, write, and drafts.

## Summary of Changes

Milestone 07 is one OpenSpec change, `compose-a-schema`, adding
`composer-model`, `composer-editing`, `composer-validation`, `composer-write`
and `composer-drafts`, and extending `app-frame` with `p`.

The last placeholder is gone; `internal/tui/placeholder.go` was deleted rather
than left as dead code.

The end to end case is the milestone's proof: two local schemas are sent to the
palette, artifacts from both are added with one renamed through a collision
prompt, linked, gated and tracked, the result is written, and OpenSpec validates
it and creates a change with it.

`internal/compose` is pure and sits at a 95% floor, behind only `graph` and
`schema`, which is where the briefing wanted the core packages.

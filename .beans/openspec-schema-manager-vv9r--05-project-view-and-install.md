---
# openspec-schema-manager-vv9r
title: 05 Project view and install
status: completed
type: milestone
priority: normal
created_at: 2026-09-29T18:00:33Z
updated_at: 2026-09-29T20:02:22Z
---

Everything that touches a real OpenSpec project: project detection, the openspec CLI adapter, the project view, and the guarded install flow.

## Summary of Changes

Milestone 05 is one OpenSpec change, `stand-in-a-project-and-install`, adding
`openspec-adapter`, `project-view` and `install`, and modifying
`schema-listing` and `schema-validation`.

ossm is now useful rather than only informative: a user inside a project sees
what it has and can install a schema from the registry into it, with nothing
written before they agree and OpenSpec's own verdict shown afterwards.

Open question 4 is closed. The install route was run by hand end to end before
anything was built on it, and an end to end test now asserts the same three
facts against a schema ossm installed itself.

The milestone's most useful finding was not planned: OpenSpec requires an
artifact `description` that the briefing calls optional. The install flow
surfaced it correctly the first time it ran, which is the system working, but
`NOTES.md` now records it and ossm reports it before the install rather than
after.

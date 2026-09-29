---
# openspec-schema-manager-b3iw
title: Built-in schema merge
status: completed
type: epic
priority: normal
created_at: 2026-09-29T18:01:15Z
updated_at: 2026-09-29T19:05:49Z
parent: openspec-schema-manager-3hlr
---

List the schemas OpenSpec reports as built in and merge them into the registry view, marked as built in, because a built-in schema is available without being installed and is deliberately absent from the registry.

## Summary of Changes

`openspec schema which --all --json` is the source, read through an interface in
`internal/openspec` with a fake for tests and a shell-script stub for the real
runner.

Its output shape is recorded in `NOTES.md`, including the one detail that would
otherwise cost an afternoon: the JSON goes to standard output while
`Note: Schema commands are experimental and may change.` goes to standard error.
The adapter reads standard output alone.

A missing or failing `openspec` binary never hides the registry. The list shows
what it has and says the built-in schemas could not be listed, with the reason.
Refusing to show anything because one of two sources is unavailable is what
makes a tool useless on a machine that is only half set up.

Rows sort by name case-insensitively, then by origin, then by `id`. The
tiebreakers exist because `name` is explicitly not unique across the registry,
and two rows sharing one must not swap places between runs. A test asserts the
same inputs give the same order twice.

A built-in row carries no ref, because there is no source to pin, and is not
installable.

One unrelated fix came out of this: `exec.CommandContext` with a buffered
stdout blocks in `Wait` until every process holding the pipe exits, which a
child of the CLI can outlive the CLI itself. `cmd.WaitDelay` bounds it, and the
timeout test went from five seconds to one.

openspec-link: openspec/changes/archive/2026-09-29-read-the-registry

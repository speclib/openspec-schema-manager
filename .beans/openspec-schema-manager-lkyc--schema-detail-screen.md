---
# openspec-schema-manager-lkyc
title: Schema detail screen
status: completed
type: epic
priority: normal
created_at: 2026-09-29T18:01:15Z
updated_at: 2026-09-29T19:34:44Z
parent: openspec-schema-manager-8ff6
---

Header, artifact table, apply gate and comparison figures, for a registry entry, a built-in schema or a local folder alike.

## Summary of Changes

The detail screen shows the schema's identity, the artifact table in declaration
order, the apply gate and tracked file, the comparison figures, and its
validation findings with fatal separated from warnings.

A schema that does not validate still opens. Knowing a schema has a cycle is
worth more before installing it than after, and a viewer that refuses to show
broken things is useless exactly when it is needed.

The screen is pushed over the Registry tab rather than becoming a fifth tab. It
belongs to whatever listed it, and milestone 06 opens the same screen from the
Local tab. The frame gained no knowledge of it: the Registry screen holds it and
delegates `Update`, `View`, `Keys` and `Capturing`, which is what kept the
frame's routing rules unchanged.

A built-in or local row is read from its path and never fetched, asserted by a
test that counts fetcher calls.

`o` shows the repository URL rather than opening a browser, because ossm runs
over SSH as often as not.

One behaviour looked like a bug in an end to end test and is not: the built-in
schema is listed before the registry arrives, so it holds the selection, and the
selection is deliberately kept across a refresh. The test presses `g` to reach
the top of the settled list.

openspec-link: openspec/changes/archive/2026-09-29-show-a-schema-in-detail

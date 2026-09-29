---
# openspec-schema-manager-sx6d
title: Update detection
status: completed
type: epic
priority: normal
created_at: 2026-09-29T18:01:15Z
updated_at: 2026-09-29T21:00:12Z
parent: openspec-schema-manager-gypt
---

Decide what update detection can honestly claim given that OpenSpec records no provenance for an installed schema, then ship it on that basis or leave it out and say so.

## Summary of Changes

Shipped as a comparison, not an update check, and it says which it is.

OpenSpec records nothing about where an installed schema came from, so ossm
cannot answer "is this behind?". It can answer "do these files differ from what
the registry offers today?", and those are different questions. `u` on the
Project tab fetches the entry declaring the same name and compares the files.

Six answers, each distinct: identical, differs, nothing in the registry declares
that name, several entries do, the source could not be reached, or the schema is
built into OpenSpec.

A difference names the limitation rather than hiding it, and a test asserts the
word "update" never appears in the message. Someone who installed a schema and
has not touched it learns that upstream moved; someone who edited it learns
their copy has drifted. Neither is told something false.

Matching is by `name`, which the registry explicitly does not make unique, so
more than one match is its own answer naming the candidates rather than a guess.

Comparison is asked for and never runs when the project view is drawn, because
it fetches. A tool whose whole offline story is that it needs no network should
not turn opening a tab into one round trip per schema.

`source.json`, ossm's own note inside a cached schema, is skipped when
comparing. An installed schema never has one, and counting it would make every
comparison differ.

openspec-link: openspec/changes/archive/2026-09-29-finish-the-proof-of-concept

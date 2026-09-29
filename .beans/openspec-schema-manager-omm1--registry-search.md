---
# openspec-schema-manager-omm1
title: Registry search
status: completed
type: epic
priority: normal
created_at: 2026-09-29T18:01:15Z
updated_at: 2026-09-29T19:05:49Z
parent: openspec-schema-manager-3hlr
---

Fuzzy filter over id, name and description, matching how specgetty filters.

## Summary of Changes

`/` filters over `id`, `name` and `description` as a case-insensitive substring
match. `esc` clears it, `enter` closes the input and keeps it applied, and the
filter text is drawn whenever it applies, so a narrowed list is never
unexplained.

Fuzzy matching was considered and not used. With a few dozen entries a substring
match is predictable, and predictability is worth more than cleverness on a list
someone is scanning to compare things. It is one function if the registry grows.

The selection stays on the selected row while it still matches and moves to the
first match when it does not, so the list never holds rows with nothing
selected.

Filtering forced a change to the frame. The frame takes `tab` and the digits
before the screen sees them, which is what keeps the tab bar working once a
screen grows a list. A filter breaks that: typing `4` must not jump to the
Composer. So `Screen` gained `Capturing() bool`, and while a screen captures,
the frame takes `ctrl+c` and nothing else. An end to end case types `4` into the
filter and asserts the Composer did not appear.

openspec-link: openspec/changes/archive/2026-09-29-read-the-registry

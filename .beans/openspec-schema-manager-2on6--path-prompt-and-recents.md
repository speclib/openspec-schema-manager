---
# openspec-schema-manager-2on6
title: Path prompt and recents
status: completed
type: epic
priority: normal
created_at: 2026-09-29T18:01:15Z
updated_at: 2026-09-29T20:21:15Z
parent: openspec-schema-manager-wwfb
---

Open any schema folder by path with completion, from a key binding or a --path flag, and keep a capped recents list in the state directory.

## Summary of Changes

`:` and `ctrl+o` open a path prompt from any tab, so it lives in the frame
rather than in the Local screen. It is the first thing in ossm that captures
text at frame level, and it follows the rule the screens already use: while it
is open the frame takes `ctrl+c` and passes everything else down. An end to end
test types a digit into it and asserts no tab changed.

Completion is `filepath.Glob` filtered to directories, filling in the longest
common prefix and listing the matches. Not a readline: a prompt used a few times
a session does not need history or word motion, and every one of those is a key
that would have to stop meaning what it means everywhere else.

The recents file lives in the state directory, is written atomically, holds no
duplicates and is capped by `recents_cap`. A cap of zero remembers nothing, which
is how a user turns it off.

A remembered path that no longer holds a schema is shown as missing rather than
dropped. Dropping it silently means the user cannot tell whether they
misremembered the path or the folder moved.

`--path` now opens a folder rather than changing which project ossm thinks it is
in, which is what the briefing describes and what the flag's help says.

openspec-link: openspec/changes/archive/2026-09-29-work-with-local-schemas

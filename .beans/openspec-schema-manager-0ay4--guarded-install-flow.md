---
# openspec-schema-manager-0ay4
title: Guarded install flow
status: completed
type: epic
priority: normal
created_at: 2026-09-29T18:01:15Z
updated_at: 2026-09-29T20:02:24Z
parent: openspec-schema-manager-vv9r
---

Install a registry entry into openspec/schemas/<name>/, showing what will be written and whether it collides before writing anything, then validating through OpenSpec. Refuse outside a project. Setting the project default is a separate, explicit step.

## Summary of Changes

`i` installs a registry schema into `openspec/schemas/<declared name>/`, through
a state machine: preparing, confirming, confirming an overwrite, writing, done
or failed. Nothing is written before the user confirms, and a test asserts the
destination does not exist after declining.

An occupied destination is its own confirmation with its own key, not a flag on
the first one. Its wording says why: OpenSpec records no provenance, so ossm
cannot tell whether what is there came from this entry. Pressing `y` at that
prompt does nothing; overwriting needs `o`.

Writing copies to a staging directory beside the destination, moves any existing
directory aside, renames the new one into place, and removes the old only after
that succeeds. A failure leaves either the old schema or none.

`openspec/config.yaml` is never touched by an install. A test compares it byte
for byte before and after.

After writing, OpenSpec's own validation runs and its verdict is shown. That is
what caught the finding that mattered most in this milestone: OpenSpec 1.10.0
rejects a schema whose artifact declares no `description`, which the briefing
calls optional and its own `research-first` fixture omits. Found by an end to
end test watching the rejection appear on screen. ossm now reports a missing
description as fatal in its own validation too, naming OpenSpec as the source of
the rule, and every fixture declares one.

The end to end suite asserts the whole thing against real OpenSpec: install,
then `openspec schema which minimalist --json` answers `"source": "project"`,
then `openspec new change --schema minimalist` succeeds.

openspec-link: openspec/changes/archive/2026-09-29-stand-in-a-project-and-install

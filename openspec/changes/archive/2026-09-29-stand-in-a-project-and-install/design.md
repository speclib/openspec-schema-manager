# Design

## What the CLI can answer, checked before building on it

Confirmed against OpenSpec 1.10.0:

| Question | Answer |
| --- | --- |
| Which schemas resolve, and from where | `openspec schema which --all --json` |
| Is this schema valid | `openspec schema validate <name> --json` |
| What changes are there | `openspec list --json` |
| Which schema does a change use | `openspec status --change <name> --json` |
| What is the project default | nothing |

The last row is why the default is read from `openspec/config.yaml`.
`openspec status --change <name> --json` does carry
`planningHome.defaultSchema`, but it needs a change to exist, and a project with
no changes is exactly where someone is deciding which schema to adopt. Reading
the project's own file is what CLAUDE.md allows and what works everywhere.

The install route was run end to end before anything was built on it: copy into
`openspec/schemas/minimalist/`, then `openspec schema validate minimalist
--json` reports `"valid": true`, `openspec schema which minimalist --json`
reports `"source": "project"`, and `openspec new change x --schema minimalist`
succeeds. That closes open question 4.

## One call per change is acceptable

`openspec list --json` gives the changes; the schema each uses needs
`openspec status --change <name> --json` per change. That is one process per
change, which is fine for the tens of changes a project has and would not be for
thousands. The calls run concurrently with a small bound, and a change whose
status cannot be read is listed with an unknown schema rather than dropped.

Reading each change's `.openspec.yaml` directly would be one file read instead.
It is not done, because the mapping from a change to its schema is OpenSpec's to
define and the file is its private layout. The adapter is the place to change if
that ever costs too much.

## The install flow is a state machine, not a prompt

```
idle → previewing → confirming → [confirmingOverwrite] → writing → done | failed
```

`previewing` has fetched the source and knows the destination, what is in it,
and what would be written. Nothing is written until `writing`. The overwrite
confirmation is its own state rather than a flag on the first one, because the
briefing asks for the collision to be a separate explicit step and a checkbox on
a confirmation is not that.

Every state is a value in the model, so the whole flow is testable through
`Update` with no terminal and no filesystem beyond a temporary directory.

## Writing

The schema is copied to a temporary directory beside the destination, then
renamed into place. The old directory, if any, is moved aside first and removed
only after the rename succeeds. So a failure partway leaves either the old
schema or no schema, never half of a new one.

`openspec/config.yaml` is not touched by an install at all. Setting the default
rewrites only the `schema:` line, preserving every comment and every other key,
because that file is full of commented guidance a user may have edited.

## What is refused and why

- Outside a project: no destination exists.
- A built-in schema: it is already resolvable, and copying it into the project
  would shadow the package copy with a stale fork of it.
- A schema the project already resolves: this is an upgrade, and an upgrade
  ossm cannot verify. OpenSpec records no provenance, so ossm cannot tell
  whether the existing directory came from this entry, from another, or from a
  hand edit. It is offered as an explicit overwrite and named as one.

## Pinning is not done

The registry is a catalogue and an entry tracking a branch installs whatever is
there today. Recording what was installed would be ossm keeping install state,
which D7 rules out. The detail view already shows whether an entry pins a ref,
which is the honest thing to show, and milestone 08 decides what update
detection can claim.

## The project screen

The Project screen holds the same detail screen the Registry screen holds, so
`enter` opens a schema from either tab with no second implementation. The
resolver already reads a local path without fetching, so a project schema opens
from where OpenSpec says it resolves.

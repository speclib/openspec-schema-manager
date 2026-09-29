---
# openspec-schema-manager-qjtn
title: OpenSpec CLI adapter
status: completed
type: epic
priority: normal
created_at: 2026-09-29T18:01:15Z
updated_at: 2026-09-29T20:02:24Z
parent: openspec-schema-manager-vv9r
---

An interface over the openspec CLI with a fake for tests, preferring --json output, covering project detection, schema which --all, schema validate and schema fork.

## Summary of Changes

`internal/openspec` now covers everything ossm asks the CLI: `schema which
--all`, `schema validate`, `list`, and `status --change`. All behind one
interface with a fake, and a shell-script stub exercises the real runner.

Two details the adapter had to learn, both found by running the CLI rather than
reading about it.

`openspec schema validate` exits non-zero when a schema is invalid while still
writing its report to standard output. The report is the answer; the exit status
alone is not. The runner now returns stdout alongside the error and
`ValidateSchema` reads the report first, treating the exit status as an error
only when nothing parseable came back.

No command reports a project's default schema. `status --change` carries
`planningHome.defaultSchema` but needs a change to exist, and a project with no
changes is exactly where someone is deciding which schema to adopt. So the
default is read from `openspec/config.yaml`, which is what CLAUDE.md allows and
what works everywhere.

Per-change schema lookups are one process each, run concurrently with a bound of
four. A change whose status cannot be read is listed with an unknown schema
rather than dropped.

ossm still keeps no record of anything. Every read goes to OpenSpec and the
project's own files.

openspec-link: openspec/changes/archive/2026-09-29-stand-in-a-project-and-install

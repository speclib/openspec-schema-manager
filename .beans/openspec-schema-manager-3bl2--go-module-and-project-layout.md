---
# openspec-schema-manager-3bl2
title: Go module and project layout
status: completed
type: epic
priority: normal
created_at: 2026-09-29T18:01:15Z
updated_at: 2026-09-29T18:10:46Z
parent: openspec-schema-manager-6ycu
---

Establish the module path, the cmd/ossm entrypoint and the internal package boundaries the briefing lays out, with the rule that internal/tui depends on the others and never the reverse.

## Summary of Changes

The module is `github.com/speclib/openspec-schema-manager`, Go 1.26, with the
entry point at `cmd/ossm` so the build produces a binary named `ossm` without a
rename step. specgetty keeps its entry point in `src/` and renames in
`postInstall`; this is the one place ossm departs from its sibling.

All eight packages the briefing names exist, each holding a doc comment that
states what belongs in it and what does not. An empty package with a stated
purpose makes the next change land in the right place.

The dependency rule is enforced rather than reviewed. `internal/arch` shells out
to `go list -json ./...` and fails when any package other than `internal/tui`
and `cmd/ossm` reaches the TUI, transitively or through a test import. The test
was proved to bite by adding the forbidden import and watching it fail.

`run` takes writers rather than files, which put everything except `main()`
within reach of an in-process test and moved `cmd/ossm` from 57.7% to 96.2%.
`--version` and `--help` answer without touching the terminal, so the binary is
drivable from a script before any TUI exists.

openspec-link: openspec/changes/archive/2026-09-29-scaffold-go-project-and-gates

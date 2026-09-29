---
# openspec-schema-manager-6jhy
title: CI and the coverage gate
status: completed
type: epic
priority: normal
created_at: 2026-09-29T18:01:15Z
updated_at: 2026-09-29T18:10:46Z
parent: openspec-schema-manager-6ycu
---

go vet, the full offline test suite and a coverage floor, wired so that nix flake check and a local run enforce the same thing.

## Summary of Changes

`scripts/coverage-gate.sh` runs `go vet ./...`, the full suite with a coverage
profile, and the floor check, and it is the only place any of those three is
written down. `nix flake check` calls it and `.github/workflows/gate.yml` calls
it. Neither restates a command or a floor, so a green check in one means the
same as in the other.

The floors are set from what the scaffold measures rather than from a number
picked in advance: `cmd/ossm` and TOTAL at 96.0 against a measured 96.2. A floor
chosen before there is code to measure is either trivially met or immediately in
the way.

The floors are allowed to fall, but only as a deliberate edit with the reason
written beside it. They will fall when the TUI lands, because wiring carries
lines a test can execute but cannot meaningfully assert. Pretending otherwise
would make the ratchet a thing people work around rather than with.

CI is two jobs: the gate on a plain Go toolchain, and `nix flake check`. The
workflow header records that a networked check belongs in a separate workflow,
because the gate's suite has to stay runnable in a sandbox with no network, the
way the registry repository already splits its source checking.

openspec-link: openspec/changes/archive/2026-09-29-scaffold-go-project-and-gates

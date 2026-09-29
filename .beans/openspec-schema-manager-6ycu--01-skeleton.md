---
# openspec-schema-manager-6ycu
title: 01 Skeleton
status: completed
type: milestone
priority: normal
created_at: 2026-09-29T18:00:33Z
updated_at: 2026-09-29T18:45:24Z
---

Repo scaffolding that everything else lands on: Go module, plain-nix flake with supported architectures, CI running go test / go vet / nix flake check, XDG config and paths, and an empty Bubble Tea app with a tab bar and help overlay.

## Summary of Changes

Milestone 01 is done in three OpenSpec changes.

`scaffold-go-project-and-gates` established the module, the package boundaries
and the gate. `scripts/coverage-gate.sh` is the single source of truth for what
passing means, and `nix flake check` and CI both call it. A test in
`internal/arch` enforces that nothing but the TUI and the command reach
`internal/tui`, and it was proved to bite before being kept.

`resolve-xdg-paths-and-configuration` gave `internal/config` the job of owning
every path ossm writes to. It creates nothing on load, decodes strictly so a
typo is an error rather than silence, and expands a leading tilde in
`schemas_dirs` once so no consumer has to.

`build-the-app-frame-and-e2e-harness` built the frame every later screen mounts
into, and the harness that drives the real binary on a real pseudo terminal
inside the nix sandbox.

`ossm` now runs: four tabs, help on `?`, `q` to quit, a status line, and a
placeholder per tab naming what will live there and which milestone builds it.
It refuses to start without a terminal and says so in a sentence.

Coverage is at or above the floor in every package with code, and the floors are
recorded with the reason for each. `cmd/ossm` was lowered on purpose when the
TUI landed, because the handover to Bubble Tea is unreachable without a
terminal and is covered end to end instead.

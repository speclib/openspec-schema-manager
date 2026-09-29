---
# openspec-schema-manager-ub7r
title: Nix flake with explicit supported systems
status: completed
type: epic
priority: normal
created_at: 2026-09-29T18:01:15Z
updated_at: 2026-09-29T18:10:46Z
parent: openspec-schema-manager-6ycu
---

A flake that builds ossm and runs the checks, enumerating supported systems with plain nixpkgs.lib.genAttrs rather than flake-utils, matching the registry repo and specgetty.

## Summary of Changes

`flake.nix` takes nixpkgs as its only input and maps an explicit four-entry
`supportedSystems` list with `nixpkgs.lib.genAttrs`. flake-utils is absent, which
is where this differs from the registry repository.

Outputs are `packages.{ossm,default}`, `checks.{build,gate}` and
`devShells.default` carrying go, gopls, git, jj and openspec. The dev shell
carries openspec and jj because the end to end tests drive both for real.

`checks.gate` runs `scripts/coverage-gate.sh` in a one-line `checkPhase` and
restates none of its commands or floors.

The Go toolchain is `buildGo126Module` from the pinned channel, matching the
1.26 that go.mod declares. specgetty pins 1.25 for the same reason: Bubble Tea
v2 raises the module's directive above the nixpkgs default, so the toolchain has
to move without the rest of nixpkgs moving with it.

`vendorHash = null` holds while the module has no dependencies. It gets a real
hash in milestone 02, when Bubble Tea arrives.

One friction worth recording: the repository is jj with a git backend, and Nix
reads the git index. A file jj has snapshotted is invisible to `nix flake check`
until `git add` stages it.

openspec-link: openspec/changes/archive/2026-09-29-scaffold-go-project-and-gates

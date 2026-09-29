---
# openspec-schema-manager-ub7r
title: Nix flake with explicit supported systems
status: todo
type: epic
created_at: 2026-09-29T18:01:15Z
updated_at: 2026-09-29T18:01:15Z
parent: openspec-schema-manager-6ycu
---

A flake that builds ossm and runs the checks, enumerating supported systems with plain nixpkgs.lib.genAttrs rather than flake-utils, matching the registry repo and specgetty.

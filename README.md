# openspec-schema-manager

`ossm` is a terminal application for discovering, inspecting, installing,
authoring and composing [OpenSpec](https://github.com/Fission-AI/OpenSpec)
workflow schemas.

A schema is a way of working with OpenSpec and an AI agent: it declares the
artifacts a change produces, their templates, and the order they depend on each
other. A light schema suits a solo side project; a heavier one with review gates
suits a team. ossm exists so that finding the one that fits a project takes
seconds rather than an afternoon of reading repositories.

It reads the [OpenSpec schema
registry](https://github.com/speclib/openspec-schema-registry) and merges in the
schemas OpenSpec reports as built in. OpenSpec owns project state: ossm keeps no
lockfile of its own and asks the `openspec` CLI what is installed.

Status: proof of concept, built milestone by milestone. See
`briefing/BRIEFING.md` for the full brief, `NOTES.md` for where real OpenSpec
behaviour differs from it, and the `.beans/` tracker for what is done and what
is next.

## Building

```sh
nix develop          # go, gopls, git, jj and openspec
nix build            # produces ./result/bin/ossm
nix flake check      # build, vet, tests and the coverage floors
```

Without Nix:

```sh
go build ./cmd/ossm
bash scripts/coverage-gate.sh
```

`scripts/coverage-gate.sh` is the single source of truth for what passing means.
`nix flake check` and CI both call it.

## Licence

MIT.

# Design

## Module path and repository name

`github.com/speclib/openspec-schema-manager`, matching the suggested repository
in the briefing package. The binary is `ossm`, built from `cmd/ossm`.

specgetty puts its entry point in `src/` and renames the binary in
`postInstall`. That works, but it costs a rename step and reads oddly to a Go
reader. `cmd/ossm` gives the right binary name from the build, so ossm departs
from its sibling here and nowhere else.

## Go version and the toolchain in nixpkgs

Bubble Tea v2 and its siblings declare `go 1.25.0`, which is above the default
Go in nixpkgs 25.05. specgetty solves this with `buildGo125Module` from the same
pinned channel, which moves the toolchain without moving the rest of nixpkgs.
ossm does the same, so both projects fail and recover together when the channel
moves.

## Package layout

```
cmd/ossm/            main, flags, wiring
internal/config      XDG paths, config.yml, recents
internal/registry    fetch, cache, parse, search, merge built-ins
internal/source      fetch a schema folder from a source and ref, cache
internal/schema      schema.yaml model, parse, validate, write
internal/graph       artifact DAG: build, cycle check, metrics, Mermaid output
internal/openspec    adapter over the openspec CLI
internal/compose     canvas, edges, id remap, template scan, write
internal/tui         Bubble Tea models per screen
```

Every package is created empty in this change, each with a doc comment stating
what belongs in it and what does not. An empty package with a stated purpose is
a cheap way to make the next change land in the right place; an absent package
invites the next change to invent its own.

## Enforcing the boundary

The rule is that `internal/tui` may import the others and none of them may
import it. A test in its own package walks the module with `go list` output
parsed from `go list -deps -json ./...`, and fails when a package outside
`internal/tui` and `cmd/ossm` lists `internal/tui` among its imports.

Using `go list` rather than parsing imports by hand means the test sees what the
compiler sees, including imports reached through build tags.

## The gate script

`scripts/coverage-gate.sh` is the single source of truth. It runs:

1. `go vet ./...`
2. `go test -coverprofile=... ./...`
3. a per-package and overall floor check against values recorded in the script

`nix flake check` calls it in a derivation whose `checkPhase` is one line, and
CI calls it directly. Neither restates the commands or the floors.

The floors start at whatever this scaffold measures and are raised in the change
that adds the tested code. Picking a number before there is code to measure
gives a floor that is either trivially met or immediately in the way.

## Coverage of a TUI

Bubble Tea models are testable through `Update` and `View` without a terminal,
so the TUI package is not exempt from the floor. It gets a lower floor than the
pure packages, because wiring and layout carry lines that a test can execute but
cannot meaningfully assert. `internal/graph` and `internal/schema` are pure and
carry the highest floors, as the briefing asks.

## Flags before the TUI exists

`ossm --version` and `ossm --help` are handled before any terminal setup, so the
end to end harness can drive the binary in CI where no terminal is attached.
`--path <dir>` is parsed and stored from this change onward even though nothing
consumes it until milestone 06, so the flag surface does not churn.

## The flake

Inputs are nixpkgs alone. `supportedSystems` is an explicit list of
`x86_64-linux`, `x86_64-darwin`, `aarch64-linux` and `aarch64-darwin`, mapped
with `nixpkgs.lib.genAttrs`, matching the registry repository and specgetty.
flake-utils is not taken as an input.

`packages.default` builds ossm from `package.nix`. `checks.build` is that
package, and `checks.tests` runs the gate script. The dev shell carries Go, the
`openspec` CLI, `jj`, `beans` and `git`, because the end to end tests drive real
`openspec` and real repositories.

## Version string

A `VERSION` file read at build time and stamped through `-ldflags`, so the
flake, CI and a local `go build` all report the same value and nothing has to be
edited in two places at release time.

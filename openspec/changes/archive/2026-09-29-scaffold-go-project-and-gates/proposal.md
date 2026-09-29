# Scaffold the Go project and its gates

Beans epics: `openspec-schema-manager-3bl2` (Go module and project layout),
`openspec-schema-manager-ub7r` (Nix flake with explicit supported systems),
`openspec-schema-manager-6jhy` (CI and the coverage gate).

## Why

Nothing can be built yet. There is no module, no build, no test command and no
gate, so every later milestone would define its own idea of what "passing"
means. The briefing asks for each milestone to build, pass tests and be
committed before the next one starts, and that promise is only checkable if one
command decides it.

Doing this first also fixes the package boundaries while they are still free.
`internal/tui` depending on `internal/graph` and never the reverse is cheap to
establish now and expensive to recover later.

## What Changes

- A Go module at `github.com/speclib/openspec-schema-manager` with the entry
  point at `cmd/ossm`, producing a binary named `ossm`.
- The empty `internal/` packages the briefing names, each with its doc comment
  and its test file, so the boundaries exist before anything fills them.
- `ossm --version` and `ossm --help` answer without a terminal, so the binary is
  testable from a script before any TUI exists.
- A Nix flake exposing `packages.default`, `devShells.default` and `checks`,
  enumerating supported systems with `nixpkgs.lib.genAttrs` over an explicit
  list rather than with flake-utils.
- One gate script that runs `go vet`, the full test suite and a coverage floor,
  called by both `nix flake check` and CI, so a green check in one place means
  the same thing as in the other.
- A GitHub Actions workflow running that gate on push and on pull request.
- A dependency boundary test that fails when `internal/tui` is imported by a
  package that is not the TUI or the command.

## Capabilities

### New Capabilities

- `project-gates`: what the build, the test suite and the coverage floor
  guarantee, and where the single source of truth for the floors lives.

### Modified Capabilities

None. This is the first change.

## Impact

- New: `go.mod`, `cmd/ossm/`, `internal/*/`, `flake.nix`, `package.nix`,
  `scripts/coverage-gate.sh`, `.github/workflows/`.
- The coverage floor starts where the scaffold lands rather than at a number
  picked in advance, and ratchets upward as milestones add tested code. A floor
  chosen before there is code to measure is a number nobody can defend.
- The module path commits to the repository name. Renaming the repository later
  costs an import rewrite across every file.
- No network is used by the test suite, because `nix flake check` builds in a
  sandbox without one. Anything needing the network is a separate workflow, the
  way the registry repository already handles source checking.

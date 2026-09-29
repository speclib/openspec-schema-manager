# Tasks

## 1. Go module and entry point

- [x] 1.1 `go mod init github.com/speclib/openspec-schema-manager`, Go 1.25
- [x] 1.2 Add `VERSION` at the repo root and `cmd/ossm/main.go` reading a
      version stamped through `-ldflags`, defaulting to `dev`
- [x] 1.3 Parse `--version`, `--help` and `--path <dir>` before any terminal
      setup; `--version` and `--help` print and exit zero
- [x] 1.4 `go build ./...` and `go vet ./...` pass

## 2. Package skeleton and the boundary test

- [x] 2.1 Create `internal/config`, `internal/registry`, `internal/source`,
      `internal/schema`, `internal/graph`, `internal/openspec`,
      `internal/compose` and `internal/tui`, each with a doc comment stating
      what belongs in it and what does not
- [x] 2.2 Write `internal/arch/boundary_test.go` that shells out to
      `go list -deps -json ./...` and fails when a package other than
      `internal/tui` or `cmd/ossm` imports `internal/tui`
- [x] 2.3 Prove the test bites: add the forbidden import temporarily, watch it
      fail, remove it
- [x] 2.4 `go test ./...` passes

## 3. Flags and their tests

- [x] 3.1 Table test over the flag parser covering `--version`, `--help`,
      `--path`, an unknown flag and no arguments
- [x] 3.2 Assert `--version` and `--help` write to standard output and exit
      zero with no terminal attached
- [x] 3.3 `go test ./...` passes

## 4. The gate script

- [x] 4.1 Write `scripts/coverage-gate.sh` running `go vet ./...`, then
      `go test -coverprofile` over all packages, then the floor check
- [x] 4.2 Record the floors in the script as the single source of truth: an
      overall floor and per-package floors, with `internal/graph` and
      `internal/schema` carrying the highest
- [x] 4.3 Set the initial floors from what the scaffold measures, not from a
      number chosen in advance
- [x] 4.4 Failure output names the package and the measured percentage
- [x] 4.5 The script exits non-zero if vet, tests or the floor check fails

## 5. The flake

- [x] 5.1 `flake.nix` with nixpkgs as the only input, pinned to the same channel
      the registry repo and specgetty use
- [x] 5.2 `supportedSystems` as an explicit list mapped with
      `nixpkgs.lib.genAttrs`; no flake-utils input
- [x] 5.3 `package.nix` building `cmd/ossm` into a binary named `ossm`, using
      `buildGo125Module`
- [x] 5.4 `checks.build` is the package; `checks.tests` runs
      `scripts/coverage-gate.sh` and restates none of its commands or floors
- [x] 5.5 `devShells.default` carries go, openspec, jj, beans and git
- [x] 5.6 `nix flake check` passes
- [x] 5.7 `nix build` produces a binary answering `--version`

## 6. CI

- [x] 6.1 `.github/workflows/gate.yml` running `scripts/coverage-gate.sh` on
      push and on pull request
- [x] 6.2 The workflow invokes the script and restates none of its steps
- [x] 6.3 Note in the workflow that networked checks belong in a separate
      workflow, as the registry repo does

## 7. Close out

- [x] 7.1 `go vet ./...`, `go test ./...` and `nix flake check` all pass
- [x] 7.2 Fill in the three beans epics' summaries and mark them completed
- [x] 7.3 Archive the change and commit

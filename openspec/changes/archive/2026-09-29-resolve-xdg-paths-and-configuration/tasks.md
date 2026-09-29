# Tasks

## 1. Paths

- [x] 1.1 Resolve the config, cache and state roots from `XDG_CONFIG_HOME`,
      `XDG_CACHE_HOME` and `XDG_STATE_HOME`, falling back to `~/.config`,
      `~/.cache` and `~/.local/state`
- [x] 1.2 Derive `ConfigFile`, `RegistryFile`, `SchemaCache`, `RecentsFile` and
      `DraftsDir` from those roots and nowhere else
- [x] 1.3 Read the environment through injectable functions so tests need no
      `t.Setenv` and can run in parallel
- [x] 1.4 Table test: each variable set, each unset, both unset, and an
      assertion that cache paths and state paths move independently
- [x] 1.5 `go test ./...` passes

## 2. Config struct and defaults

- [x] 2.1 Define `Config` with the registry URL, the time to live, the schema
      directories and the recents cap, and nothing else
- [x] 2.2 Defaults: the published registry URL, 24h, an empty list, 20
- [x] 2.3 `Load` returns the defaults when no file exists, without error
- [x] 2.4 Test that a partial file leaves the omitted settings at their defaults
- [x] 2.5 `go test ./...` passes

## 3. Strict, typed decoding

- [x] 3.1 Decode with `KnownFields(true)` so an unknown key is an error
- [x] 3.2 Decode `registry_ttl` as a string and parse it as a duration, so a
      bare number is rejected and the error names the field
- [x] 3.3 Reject a negative recents cap, naming the field
- [x] 3.4 Every error names the file
- [x] 3.5 Table test over malformed YAML, a non-mapping top level, a wrongly
      typed value, an unknown key and a negative cap
- [x] 3.6 `go test ./...` passes

## 4. Tilde expansion

- [x] 4.1 Expand a leading `~/` in each `schemas_dirs` entry at load time
- [x] 4.2 Leave a tilde that is not leading alone
- [x] 4.3 Test both, plus a bare `~` and an already absolute path
- [x] 4.4 `go test ./...` passes

## 5. Creating directories

- [x] 5.1 `EnsureDir` creates a directory and its parents, and is a no-op when
      it exists
- [x] 5.2 Assert `Load` creates nothing: load against an empty cache and state
      root, then assert both are still empty
- [x] 5.3 `go test ./...` passes

## 6. Example configuration

- [x] 6.1 Write `docs/config.example.yml` with every key, its default and one
      line saying what it does
- [x] 6.2 Test that the example decodes with `KnownFields(true)`, so it cannot
      drift from the struct
- [x] 6.3 `go test ./...` passes

## 7. Close out

- [x] 7.1 Raise the coverage floor for `internal/config` in the gate script to
      the measured value
- [x] 7.2 `bash scripts/coverage-gate.sh` and `nix flake check` pass
- [x] 7.3 Fill in the epic's summary and mark it completed
- [x] 7.4 Archive the change and commit

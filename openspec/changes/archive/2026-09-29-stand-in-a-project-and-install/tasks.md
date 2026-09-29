# Tasks

## 1. The adapter

- [x] 1.1 `Changes(ctx, dir)` from `openspec list --json`, carrying each
      change's name, done and total tasks
- [x] 1.2 `ChangeSchema(ctx, dir, name)` from `openspec status --change <name>
      --json`
- [x] 1.3 `Validate(ctx, dir, name)` from `openspec schema validate <name>
      --json`, carrying validity and the issues
- [x] 1.4 `DefaultSchema(dir)` reading `openspec/config.yaml`; absent or
      unparseable reads as unknown, not as an error
- [x] 1.5 Extend the fake with all of them
- [x] 1.6 Tests against a shell stub for each: success, no changes, a failing
      command, output that does not parse, a missing binary
- [x] 1.7 `go test ./...` passes

## 2. Reading a project

- [x] 2.1 One type carrying the root, the default schema, the resolved schemas
      and the changes
- [x] 2.2 Per-change schema lookups run concurrently with a small bound; a
      change whose status cannot be read is listed with an unknown schema
- [x] 2.3 A shadowed schema is reported as shadowing
- [x] 2.4 Test with the fake: a full project, a project with no changes, a
      project where the CLI fails
- [x] 2.5 `go test ./...` passes

## 3. The Project screen

- [x] 3.1 Inside a project: root, default schema, schemas with origins, changes
      with schemas and progress
- [x] 3.2 Outside a project: say so, say browsing works, say installing needs
      one
- [x] 3.3 A project with no changes says so rather than showing an empty area
- [x] 3.4 `r` re-reads the project
- [x] 3.5 `enter` opens a schema's detail from its resolved path
- [x] 3.6 Movement and scrolling for a long list
- [x] 3.7 `Keys()` lists what it handles
- [x] 3.8 Test through `Update` and `View` for each
- [x] 3.9 `go test ./...` passes

## 4. The install flow

- [x] 4.1 The state machine: idle, previewing, confirming, confirming an
      overwrite, writing, done, failed
- [x] 4.2 Preview: destination, the files that would be written, and what is
      already there
- [x] 4.3 Nothing is written before confirmation; a test asserts the project is
      untouched after declining
- [x] 4.4 An occupied destination refuses and offers overwriting as its own
      confirmation
- [x] 4.5 Write to a temporary directory beside the destination and rename; move
      the old one aside first and remove it only after the rename succeeds
- [x] 4.6 Run OpenSpec validation afterwards and report what it says
- [x] 4.7 A schema OpenSpec rejects is reported as installed and invalid
- [x] 4.8 Validation that cannot run is reported as unverified
- [x] 4.9 The destination is `openspec/schemas/<declared name>/`, not the source
      path's last segment
- [x] 4.10 Test each state and each refusal
- [x] 4.11 `go test ./...` passes

## 5. Refusals

- [x] 5.1 Refuse outside a project, saying installing needs one
- [x] 5.2 Refuse for a built-in schema, saying it ships with OpenSpec
- [x] 5.3 Refuse for a schema the project already resolves, offering overwriting
      as its own choice
- [x] 5.4 Test each
- [x] 5.5 `go test ./...` passes

## 6. Setting the default

- [x] 6.1 A separate action with its own confirmation
- [x] 6.2 Rewrite only the `schema:` line, preserving comments and every other
      key
- [x] 6.3 A test asserts `openspec/config.yaml` is byte for byte unchanged after
      an install
- [x] 6.4 A test asserts the comments survive setting the default
- [x] 6.5 `go test ./...` passes

## 7. Wiring

- [x] 7.1 `i` on a registry row starts the install
- [x] 7.2 The project view and the schema list refresh after an install
- [x] 7.3 The Project screen is built from the adapter in `main`
- [x] 7.4 `go test ./...` passes

## 8. End to end

- [x] 8.1 A real OpenSpec project created by the test, with a real git repository
      as the registry's source
- [x] 8.2 The Project tab shows the root, the default schema and `spec-driven`
- [x] 8.3 Installing writes the schema and `openspec schema which` finds it
- [x] 8.4 A change can be created with the installed schema
- [x] 8.5 Declining the confirmation leaves the project untouched
- [x] 8.6 Installing outside a project is refused

## 9. Close out

- [x] 9.1 Record in `NOTES.md` that open question 4 is closed and how, and that
      no command reports the project default
- [x] 9.2 Raise the coverage floors to the measured values
- [x] 9.3 `bash scripts/coverage-gate.sh` and `nix flake check` pass
- [x] 9.4 Fill in the three epics' summaries and mark them completed; close
      milestone 05
- [x] 9.5 Archive the change, commit and push

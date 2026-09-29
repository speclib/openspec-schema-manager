# Tasks

## 1. Comparison

- [x] 1.1 Compare an installed schema folder against a fetched source folder,
      file by file, reporting changed, added and removed
- [x] 1.2 Report identical, differs, nothing to compare, several candidates,
      unreachable, and built-in as distinct answers
- [x] 1.3 Match an installed schema to a registry entry by name; several matches
      is its own answer naming them
- [x] 1.4 Say which ref was compared against, or that it was the default branch
- [x] 1.5 Table test over each answer
- [x] 1.6 `go test ./...` passes

## 2. The Project action

- [x] 2.1 `u` compares the selected schema and reports in place
- [x] 2.2 Nothing is fetched when the project view is drawn or refreshed
- [x] 2.3 A difference never reads as an update, and names the limitation
- [x] 2.4 A comparison changes no file in the project
- [x] 2.5 Test each through `Update` and `View`
- [x] 2.6 `go test ./...` passes

## 3. The acceptance suite

- [x] 3.1 `test/e2e/acceptance_test.go` with one case per criterion, named after
      it
- [x] 3.2 Outside a project: registry and local schemas shown, install refused
- [x] 3.3 Inside a project: default schema, schemas with origins, changes with
      schemas
- [x] 3.4 A registry schema opens with artifacts and a diagram, and still opens
      after the source is made unreachable
- [x] 3.5 Installing is seen by OpenSpec and a change can be created with it
- [x] 3.6 Duplicate, edit a template, return, and see the updated content and
      the validation result
- [x] 3.7 Compose from two fixtures: catch a cycle, warn about a dangling
      reference, write a schema OpenSpec accepts and can create a change with
- [x] 3.8 A criterion that cannot be checked here skips with the reason
- [x] 3.9 The suite runs with no network

## 4. The key reference

- [x] 4.1 Generate `docs/keys.md` from every screen's `Keys()` and the global
      list
- [x] 4.2 A test asserts the file matches what the code would produce now
- [x] 4.3 `go test ./...` passes

## 5. Demos

- [x] 5.1 A script that builds a fixture registry, a demo project and a local
      schemas directory, so a recording needs no network
- [x] 5.2 VHS tapes in `demo/` for browsing the registry, opening a schema and
      its diagram, installing, and composing
- [x] 5.3 Record them and check the output in
- [x] 5.4 Add vhs to the dev shell

## 6. The README

- [x] 6.1 What ossm is and what it does
- [x] 6.2 How to run it, with and without Nix
- [x] 6.3 Configuration, with the example file
- [x] 6.4 What it cannot do: no update detection beyond comparison, git needed
      to fetch, no raw-file fallback, schema commands experimental
- [x] 6.5 The demos and the key reference linked
- [x] 6.6 How the repository is organised and where the specs and beans live

## 7. Close out

- [x] 7.1 Update `NOTES.md` on what update detection ships as and why
- [x] 7.2 Raise the coverage floors to the measured values
- [x] 7.3 `bash scripts/coverage-gate.sh` and `nix flake check` pass
- [x] 7.4 Fill in the three epics' summaries and mark them completed; close
      milestone 08 and the project
- [x] 7.5 Archive the change, commit and push

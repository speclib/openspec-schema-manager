# Tasks

## 1. The source fetcher

- [x] 1.1 `Source{Repo, Path, Ref}` and a `Fetcher` interface returning a
      directory
- [x] 1.2 Cache path from `sha256(repo\x00ref\x00path)` under the schema cache
      root, with a `source.json` recording the three in full
- [x] 1.3 A cached source returns its directory without running git
- [x] 1.4 Git implementation: `init`, `remote add`, `fetch --depth 1`,
      `checkout --detach FETCH_HEAD`
- [x] 1.5 With no ref, fetch `HEAD` so the repository resolves its own default
      branch
- [x] 1.6 Copy the named path out of the checkout into a temporary directory in
      the cache root, verify `schema.yaml` parses, then rename into place
- [x] 1.7 A missing path, a ref that does not resolve, and a directory holding
      no `schema.yaml` each fail naming the reason
- [x] 1.8 A failed fetch leaves an existing cached copy readable and leaves
      nothing partial behind
- [x] 1.9 A missing `git` fails saying so, distinguishably from any other error
- [x] 1.10 Refetch replaces the cached copy
- [x] 1.11 Tests build real git repositories in temporary directories, including
      one whose default branch is `master`; no test reaches the network
- [x] 1.12 `go test ./...` passes

## 2. Resolving what to open

- [x] 2.1 One type carrying a resolved schema: its directory, its parsed
      `Schema`, its graph, its findings, its origin, and its repository and ref
      where it has them
- [x] 2.2 A registry row resolves by fetching; a built-in or local row resolves
      by reading its path
- [x] 2.3 Findings come from `ValidateDir`, since there is now a directory
- [x] 2.4 Test each origin, including a schema that does not validate
- [x] 2.5 `go test ./...` passes

## 3. The detail screen

- [x] 3.1 Header: name, version, description, repository and ref, or path for a
      local schema
- [x] 3.2 A schema tracking the default branch says so rather than naming a
      branch
- [x] 3.3 Artifact table: id, generates, template, requires, in declaration
      order, with "nothing" rather than a blank for no requirements
- [x] 3.4 Apply gate, tracked file, and the figures from `graph.Metrics`
- [x] 3.5 Findings shown with fatal separated from warnings
- [x] 3.6 A schema that does not validate still opens
- [x] 3.7 The table scrolls when there are more artifacts than rows
- [x] 3.8 `Keys()` lists `d`, `R`, `o`, `esc` and the movement keys
- [x] 3.9 Test through `Update` and `View` for each of those
- [x] 3.10 `go test ./...` passes

## 4. Opening and leaving

- [x] 4.1 `enter` on a listed row opens the detail; `enter` on an empty list
      does nothing
- [x] 4.2 `esc` and `q` return to the list with the same row selected
- [x] 4.3 `q` on the detail does not quit ossm
- [x] 4.4 An uncached schema reports that it is being fetched and does not block
      the frame
- [x] 4.5 A failed fetch shows the reason and offers `R`
- [x] 4.6 `R` refetches and redraws
- [x] 4.7 `o` shows the repository URL, or the path when there is no repository
- [x] 4.8 Test each, with a fake fetcher
- [x] 4.9 `go test ./...` passes

## 5. The diagram

- [x] 5.1 Render through `mermaid-ascii` in process from `graph.Mermaid`
- [x] 5.2 `UseAscii` from whether the environment's locale says UTF-8
- [x] 5.3 Render once into a string when the diagram opens or the schema changes
- [x] 5.4 A cyclic schema says it cannot be drawn and names the artifacts
- [x] 5.5 A schema with no artifacts says there is nothing to draw
- [x] 5.6 Test against the chain, branchy, wide and cycle fixtures
- [x] 5.7 `go test ./...` passes

## 6. The viewport

- [x] 6.1 Split the rendered diagram into lines once
- [x] 6.2 Slice by rune, not by byte, so box-drawing characters survive
- [x] 6.3 Scroll up, down, left and right, clamped at all four edges
- [x] 6.4 A diagram that fits never moves
- [x] 6.5 `d` toggles the diagram
- [x] 6.6 Table test over a diagram taller, wider, both and neither
- [x] 6.7 `go test ./...` passes

## 7. End to end

- [x] 7.1 A fixture registry pointing at a git repository built by the test
- [x] 7.2 `enter` on a row opens the detail and shows its artifacts
- [x] 7.3 `d` opens the diagram and box-drawing characters appear
- [x] 7.4 `esc` returns to the list with the same row selected
- [x] 7.5 A second open of the same schema makes no network request
- [x] 7.6 `enter` on a built-in row opens it without fetching

## 8. Close out

- [x] 8.1 Record in `NOTES.md` that mermaid-ascii exports an importable
      renderer, closing open question 6
- [x] 8.2 Add `git` to the dev shell and to the flake check's inputs
- [x] 8.3 Raise the coverage floors to the measured values
- [x] 8.4 `bash scripts/coverage-gate.sh` and `nix flake check` pass
- [x] 8.5 Fill in the three epics' summaries and mark them completed; close
      milestone 04
- [x] 8.6 Archive the change, commit and push

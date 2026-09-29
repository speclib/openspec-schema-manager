# Tasks

## 1. The entry model

- [x] 1.1 Mirror the registry's JSON Schema field for field in
      `internal/registry`: `id`, `name`, `description`, `artifacts`, `source`,
      `requires`, `status`, `superseded_by`, `license`, `language`
- [x] 1.2 Accessors apply the documented default for an absent optional field:
      `Status`, `Language`, `License`, `Pinned`
- [x] 1.3 `Pinned` reports whether `source.ref` was set; nothing substitutes a
      branch name for an absent ref
- [x] 1.4 Table test over each absent optional field and its documented meaning
- [x] 1.5 `go test ./...` passes

## 2. Loading and validating

- [x] 2.1 Parse a registry document, ignoring unknown fields
- [x] 2.2 Reject an entry missing any of the six required fields, naming it by
      `id`, or by position when `id` itself is missing
- [x] 2.3 Reject an `id` not shaped `<owner>/<name>` and an empty `artifacts`
- [x] 2.4 Reject a document whose top level is not an object or whose `schemas`
      is missing or not an array, naming the source
- [x] 2.5 An empty `schemas` array is valid and yields no entries
- [x] 2.6 Table test over every rejection above plus a valid document carrying
      an unknown field
- [x] 2.7 Fixtures shaped like the real file live in `internal/registry/testdata`
- [x] 2.8 `go test ./...` passes

## 3. Fetching

- [x] 3.1 Define the one-method `Fetcher` interface
- [x] 3.2 HTTP implementation: timeout, size limit, non-success status becomes
      an error naming the status
- [x] 3.3 `file://` URLs are supported through the same interface
- [x] 3.4 Test with a fake fetcher: success, a non-success status, a body that
      is not JSON, a body that is JSON but not a registry, a body over the size
      limit, and a timeout
- [x] 3.5 No test makes a network request
- [x] 3.6 `go test ./...` passes

## 4. The cache

- [x] 4.1 Cache file records `fetched_at`, the URL it came from, and the
      registry document
- [x] 4.2 Reading reports the entries and the age; a corrupt or missing cache
      reads as no cache, not as an error
- [x] 4.3 A cache recorded against a different URL is treated as no cache
- [x] 4.4 Writing validates first, then writes to a temporary file in the same
      directory and renames
- [x] 4.5 A failed fetch leaves the cache file byte for byte unchanged
- [x] 4.6 Test each of those, including an interrupted write leaving no partial
      file
- [x] 4.7 `go test ./...` passes

## 5. Built-in schemas

- [x] 5.1 Read the output shape of `openspec schema which --all --json` from the
      installed CLI and record it in `NOTES.md`
- [x] 5.2 Add the adapter method behind the existing interface, with a fake
      for tests
- [x] 5.3 A missing or failing `openspec` binary yields no built-in schemas and
      a reportable reason, never an error that hides the registry
- [x] 5.4 Test with a fake adapter: schemas reported, none reported, the binary
      missing, output that does not parse
- [x] 5.5 `go test ./...` passes

## 6. The merged listing

- [x] 6.1 Merge registry entries and built-in schemas into one row type
      carrying name, artifact count, workflow shape, origin and ref
- [x] 6.2 A built-in row carries no ref and is marked as already available
- [x] 6.3 A deprecated row says so and names its replacement
- [x] 6.4 Sort by name case-insensitively, then origin, then id
- [x] 6.5 Test that the same inputs give the same order, and that two rows
      sharing a name keep a stable relative order
- [x] 6.6 `go test ./...` passes

## 7. Search

- [x] 7.1 Case-insensitive substring match over `id`, `name` and `description`
- [x] 7.2 The selection stays on the selected row while it matches and moves to
      the first match when it does not
- [x] 7.3 An empty filter matches everything
- [x] 7.4 Table test over each field matching, case, no match, and the selection
      rules
- [x] 7.5 `go test ./...` passes

## 8. Capturing keys in the frame

- [x] 8.1 Add `Capturing() bool` to `Screen`
- [x] 8.2 While a screen captures, the frame takes `ctrl+c` and nothing else
- [x] 8.3 Placeholders never capture
- [x] 8.4 Test that a digit typed into a filter does not switch tabs and that
      `ctrl+c` still quits
- [x] 8.5 `go test ./...` passes

## 9. The Registry screen

- [x] 9.1 Draw the merged list with a selection, scrolling for a list longer
      than the pane
- [x] 9.2 `/` opens the filter, `esc` clears it, the filter text is drawn while
      it applies
- [x] 9.3 `r` refreshes; report while it runs and when it finishes or fails
- [x] 9.4 The status line carries the entry count and the cache age
- [x] 9.5 With no cache and no network, say the registry has never been fetched
      and that `r` will try again
- [x] 9.6 Report when the built-in schemas could not be listed
- [x] 9.7 The tab's `Keys()` lists `/`, `esc`, `r` and the movement keys
- [x] 9.8 Test through `Update` and `View`, with no network and no CLI
- [x] 9.9 `go test ./...` passes

## 10. Wiring

- [x] 10.1 Read the cache on start, synchronously
- [x] 10.2 Fetch behind it as a `tea.Cmd` when the cache is older than the time
      to live, replacing the list when it arrives
- [x] 10.3 A failed fetch writes to the status line and leaves the list alone
- [x] 10.4 Test the whole sequence through `Update` with a fake fetcher

## 11. End to end

- [x] 11.1 A fixture registry served from a `file://` URL through a written
      `config.yml`
- [x] 11.2 Starting with that config lists the fixture's entries
- [x] 11.3 `/` filters the list and `esc` restores it
- [x] 11.4 A digit typed into the filter does not switch tabs
- [x] 11.5 With an unreachable URL and no cache, the tab says so
- [x] 11.6 With a cache and an unreachable URL, the list still works and the age
      is shown

## 12. Close out

- [x] 12.1 Record in `NOTES.md` the `openspec schema which --all --json` output
      shape and any disagreement with the registry's JSON Schema
- [x] 12.2 Raise the coverage floors to the measured values
- [x] 12.3 `bash scripts/coverage-gate.sh` and `nix flake check` pass
- [x] 12.4 Fill in the three epics' summaries and mark them completed; close
      milestone 02
- [x] 12.5 Archive the change, commit and push

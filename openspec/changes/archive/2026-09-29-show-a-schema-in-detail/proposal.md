# Show a schema in detail

Beans epics: `openspec-schema-manager-ld29` (Schema source fetch and cache),
`openspec-schema-manager-lkyc` (Schema detail screen),
`openspec-schema-manager-f007` (Diagram rendering in a viewport).

## Why

The registry list says a schema exists and roughly how big it is. It cannot
answer the question someone actually has, which is whether this way of working
suits the project in front of them. That needs the artifacts, what each one
produces, what gates implementation, and the shape of the dependencies, and the
shape is the part a table is worst at showing.

A registry entry is thin on purpose. Everything past the list has to come from
the schema's own repository, which means fetching it, which means caching it,
because nobody should wait for a network round trip to look at a schema twice.

## What Changes

- Opening a schema fetches its folder from the repository, path and ref its
  entry names, and caches it under the cache root keyed by that triple.
- A schema whose entry names no ref is fetched at the repository's default
  branch, resolved by asking the repository rather than assuming `main`.
- A cached schema opens without touching the network. `R` on the detail view
  refetches it.
- Fetching needs `git`. Without it, the detail view says so plainly and the rest
  of ossm keeps working.
- A schema already on disk, whether built into OpenSpec or in a local directory,
  is read from where it is and never fetched.
- The detail view shows the name, version, description, source and ref; a table
  of artifacts with their id, what they generate, their template and their
  requirements; the apply gate and the tracked file; and the comparison figures.
- Any validation finding is shown with the schema, fatal ones apart from
  warnings, because a schema that does not validate is worth knowing about
  before installing it.
- `d` toggles a flow diagram drawn from the artifact graph, scrollable in both
  directions when it is larger than the pane.
- `o` copies the source repository URL to the status line so it can be opened
  by hand. `esc` and `q` return to the list.

## Capabilities

### New Capabilities

- `source-fetch`: how a schema folder is fetched from a source, what is cached,
  and what happens when the network, the ref or `git` is missing.
- `schema-detail`: what the detail view shows and how it is reached and left.
- `diagram-view`: how the diagram is drawn and scrolled.

### Modified Capabilities

- `schema-listing`: a row can now be opened, so the listing gains an action and
  has to say which rows can be opened and which cannot.

## Impact

- New: `internal/source`, the detail and diagram screens in `internal/tui`, and
  a dependency on `github.com/pgavlin/mermaid-ascii` used in process.
- `git` becomes a runtime dependency for fetching, and joins the dev shell and
  the flake's check inputs. It is already there for the tests.
- Source fetching is tested against git repositories created inside the test,
  so the suite still makes no network request.
- No mechanism is built for a source that is not a git repository. Every entry
  in the registry today is one, and a raw-file fallback cannot list a directory
  without a host-specific API. The adapter is one interface, so adding one later
  changes nothing above it.

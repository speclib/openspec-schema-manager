# Read the registry

Beans epics: `openspec-schema-manager-4kcf` (Registry fetch and cache),
`openspec-schema-manager-omm1` (Registry search),
`openspec-schema-manager-b3iw` (Built-in schema merge).

## Why

The Registry tab is the reason ossm exists. A user who never finds out that
OpenSpec supports custom schemas cannot compare them, and comparing them is what
the registry is for.

Two things have to be true before that tab is worth opening. It has to work on a
train, because a tool that needs the network to show you a list you already had
is a tool people stop trusting. And it has to show every schema available, not
just the installable ones: `spec-driven` is deliberately absent from the
registry because it ships with OpenSpec, and a list that omits it is lying about
what a user can choose.

## What Changes

- ossm fetches `openspec-schemas.json` from the configured URL, validates it,
  and writes it to the cache.
- On start it reads the cache and shows it immediately. When the cache is older
  than the configured time to live, a fresh copy is fetched in the background
  and the list updates when it arrives.
- `r` forces a refresh whatever the cache age says.
- With no cache and no network, the tab says so and offers `r`, rather than
  showing an empty list.
- With a cache and no network, the list works and the status line says how old
  it is.
- A fetch that returns something ossm cannot parse leaves the existing cache
  alone. A bad copy on the server must not destroy a good copy on disk.
- `/` filters the list over id, name and description. `esc` clears the filter.
- The schemas OpenSpec reports as built in are merged into the same list,
  marked as built in, and cannot be installed because they are already there.
- Each entry shows its artifact count, the shape of its workflow, and whether
  its source pins a ref or tracks a branch.
- A deprecated entry is shown as deprecated and names its replacement.

## Capabilities

### New Capabilities

- `registry-model`: what an entry is, which fields are required, and what an
  absent optional field means.
- `registry-cache`: fetching, caching, the time to live, refreshing, and what
  happens when the network or the cache is missing.
- `registry-search`: how the filter matches and what it matches against.
- `schema-listing`: the merged list of registry and built-in schemas, and what
  each row carries.

### Modified Capabilities

- `app-frame`: the Registry tab stops being a placeholder, so its keys and its
  status line change.

## Impact

- New: `internal/registry` with the model, the loader, the cache and the search;
  `internal/openspec` gains the adapter call that lists built-in schemas;
  `internal/tui` gains the Registry screen.
- The entry model is copied from the registry repository's JSON Schema, not
  invented here. Where the two disagree, the registry wins and the difference
  goes in `NOTES.md`.
- ossm validates what it fetches but does not reimplement the registry's own
  linting. A description over 100 characters is the registry's problem to
  reject; ossm shows what it was given.
- Unknown fields in an entry are ignored rather than rejected. The registry may
  add a field before ossm knows about it, and refusing to read the whole file
  over one unknown key would strand every user until they upgrade.

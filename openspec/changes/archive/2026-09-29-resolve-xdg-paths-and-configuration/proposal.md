# Resolve XDG paths and configuration

Beans epic: `openspec-schema-manager-lpys` (Config, XDG paths and state).

## Why

Six of the seven remaining milestones write a file somewhere: a cached registry,
a fetched schema folder, a recents list, a composition draft, a local schema.
If each works out its own path, ossm ends up with several ideas of where its
data lives, and a user who moves `XDG_CACHE_HOME` finds half of it followed and
half of it did not.

The defaults also need one home. The registry URL, the cache time to live and
the list of local schema directories are settings a user will want to change
before any of the code that reads them exists.

## What Changes

- ossm reads `config.yml` from `$XDG_CONFIG_HOME/ossm/`, falling back to
  `~/.config/ossm/` when that variable is unset, and runs on defaults when the
  file is absent.
- A missing configuration file is not an error. A malformed one is, and the
  message names the file and the field.
- The settings are the registry URL, the cache time to live, the local schema
  directories, and the cap on the recents list. Every one has a default that
  works with no configuration at all.
- Paths for the cache root, the state root, the registry cache file, the schema
  cache directory, the recents file and the drafts directory are all derived in
  one place, so no other package composes a path.
- `~` at the start of a configured directory expands to the user's home, because
  a hand-written `schemas_dirs` entry is the one place a user will reach for it.
- A directory is created when something is written into it, not when the
  configuration is read. Browsing the registry must not leave a drafts directory
  behind.
- Unknown keys in `config.yml` are an error rather than silence, so a typo
  is reported instead of quietly taking no effect.

## Capabilities

### New Capabilities

- `configuration`: what ossm reads, where it reads it from, what it does when it
  is missing or malformed, and what each setting defaults to.

### Modified Capabilities

None.

## Impact

- New: `internal/config`, its tests, and a documented example `config.yml`.
- Nothing yet consumes the settings. Milestone 02 is the first reader. Fixing
  the shape now means the registry fetcher is written against an interface that
  already exists rather than inventing one and having it rewritten.
- The default registry URL commits to the address the registry repository
  published in `openspec-schema-registry-9hi8`. It is a setting so that a
  fork, a mirror or a local file can be pointed at without a rebuild.

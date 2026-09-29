---
# openspec-schema-manager-lpys
title: Config, XDG paths and state
status: completed
type: epic
priority: normal
created_at: 2026-09-29T18:01:15Z
updated_at: 2026-09-29T18:15:35Z
parent: openspec-schema-manager-6ycu
---

config.yml under XDG_CONFIG_HOME, the cache root under XDG_CACHE_HOME and the state root under XDG_STATE_HOME, with defaults for registry URL, TTL and schemas_dirs.

## Summary of Changes

`internal/config` resolves the three XDG roots and derives every path ossm
writes to: the registry cache file, the fetched schema cache, the recents file
and the drafts directory. `Paths` and `Config` are separate types because one
comes from the environment and the other from a file, and letting a setting
control a location is how a cache ends up somewhere a user did not ask for.

`Load` creates nothing. `EnsureDir` and `EnsureParent` are called by whatever is
about to write. A test loads against an empty root and asserts it is still
empty, so browsing the registry cannot leave a drafts directory behind.

Decoding is strict: `KnownFields(true)` turns a typo into an error naming the
key. A setting that is silently ignored looks exactly like a setting that did
not work, and that is the harder bug.

`registry_ttl` is decoded as a string and parsed as a duration, so a bare `30`
is rejected rather than guessed at. `0s` is accepted and means fetch on every
launch, which is how a user asks for that without a flag.

A leading `~/` expands in `schemas_dirs` alone, at load time, so no consumer has
to remember to do it. A tilde anywhere else in a path is left alone, because it
is part of a real name there.

The environment is read through injectable functions rather than `t.Setenv`,
which keeps the package's tests parallel. Coverage is 100%, and the floor is set
at 98 because the unreadable-file test skips when the suite runs as root, as
some CI containers do.

`docs/config.example.yml` carries every key at its default, and a test decodes
it strictly so it cannot drift from the struct.

openspec-link: openspec/changes/archive/2026-09-29-resolve-xdg-paths-and-configuration

---
# openspec-schema-manager-4kcf
title: Registry fetch and cache
status: completed
type: epic
priority: normal
created_at: 2026-09-29T18:01:15Z
updated_at: 2026-09-29T19:05:49Z
parent: openspec-schema-manager-3hlr
---

Fetch the published registry over HTTP, cache it, honour a configurable TTL, refresh on demand, and keep working from cache when the network is gone. Report cache age.

## Summary of Changes

`internal/registry` mirrors the registry repository's JSON Schema field for
field, fetches through a one-method `Fetcher` interface so no test touches the
network, and caches to a file recording the fetch time and the URL it came from.

Three decisions are worth keeping.

The fetch time is recorded rather than read from the file's modification time,
because a checkout, a restore or a backup tool moves that time without anything
having been fetched. The URL is recorded too, so changing `registry_url`
invalidates the cache instead of silently serving a mirror's copy as the
canonical one.

A response is validated before it is written, and written to a temporary file in
the same directory and renamed. A bad copy on the server must not destroy a good
copy on disk, and a test asserts the cache is byte for byte unchanged after a
failed refresh for four kinds of failure.

`source.ref` absent is never turned into `"main"`. `Pinned()` reports whether it
was set and `RefLabel()` says "default branch", because the seeded registry
already holds a repository whose default branch is `master`. The same reasoning
gives `EffectiveStatus`, `EffectiveLanguage` and `LicenseKnown` rather than
handing out zero values.

A `file://` URL goes through the same `Fetcher`, which is how the end to end
tests point ossm at a fixture registry with no server.

Reading is deliberately more permissive than the registry's own validation: an
unknown field is ignored, and the 100-character description cap is not enforced.
`NOTES.md` records why, and why `internal/config` does the opposite.

openspec-link: openspec/changes/archive/2026-09-29-read-the-registry

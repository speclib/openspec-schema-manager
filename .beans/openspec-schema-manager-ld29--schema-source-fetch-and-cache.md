---
# openspec-schema-manager-ld29
title: Schema source fetch and cache
status: completed
type: epic
priority: normal
created_at: 2026-09-29T18:01:15Z
updated_at: 2026-09-29T19:34:44Z
parent: openspec-schema-manager-8ff6
---

Fetch a schema folder from its source repo, path and ref into the cache keyed by source and ref, preferring a shallow git clone and falling back to raw file fetch.

## Summary of Changes

`internal/source` fetches a schema folder with git and caches it under a hash of
`repo`, `ref` and `path` together.

The git commands are `init`, `remote add`, `fetch --depth 1`, `checkout --detach
FETCH_HEAD`, not `git clone --branch`. Clone with a branch fails on a commit
SHA, and an entry may pin one. With no ref the fetch is `origin HEAD`, which
makes the repository resolve its own default branch. That is the case the
registry already warns about, and the tests cover it with repositories whose
default branch is `master` and `trunk`.

The cache key is a hash rather than a readable path. A repository URL, a ref and
a path can each hold characters a directory name cannot, and sanitising three of
them into one name is how two sources end up sharing a directory. A
`source.json` inside each cached folder records all three in full, so the cache
is still inspectable by hand.

The ref recorded is what the entry asked for, not what it resolved to. Keying on
the resolved commit would refetch on every upstream push, which is the opposite
of caching.

A fetch is assembled in a temporary directory in the cache root, verified to
hold a `schema.yaml` that parses, and only then renamed into place. Tests assert
a failed refetch leaves the cached copy readable and no `.fetch-` directory
behind.

No fallback was built for a source that is not a git repository. Every registry
entry today is one, and a raw-file fetch cannot list a directory without a
host-specific API, so the fallback would be a GitHub adapter under a general
name. Recorded in `NOTES.md`.

Every test builds its own git repository in a temporary directory, so the suite
exercises the real git commands and still makes no network request.

openspec-link: openspec/changes/archive/2026-09-29-show-a-schema-in-detail

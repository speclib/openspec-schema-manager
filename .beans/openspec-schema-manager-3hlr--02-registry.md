---
# openspec-schema-manager-3hlr
title: 02 Registry
status: completed
type: milestone
priority: normal
created_at: 2026-09-29T18:00:33Z
updated_at: 2026-09-29T19:05:48Z
---

Read the published registry and make it usable offline: fetch, cache, TTL, search, and the merge with the schemas OpenSpec reports as built-in.

## Summary of Changes

Milestone 02 is one OpenSpec change, `read-the-registry`.

The Registry tab is real. It reads the cache synchronously on start, shows it,
and fetches behind it when the cache is older than the time to live. It works
offline, says how old the cache is, and says so plainly when the registry has
never been fetched. `r` forces a refresh and a failed one leaves the list alone.

Five capabilities were specified: `registry-model`, `registry-cache`,
`registry-search`, `schema-listing`, and a modification to `app-frame` for
screens that capture text.

The end to end suite grew five cases driving a fixture registry through a
`file://` URL: listing it, filtering and restoring, a digit typed into the
filter not switching tabs, an unreachable registry with no cache, and a cached
registry surviving its source being deleted.

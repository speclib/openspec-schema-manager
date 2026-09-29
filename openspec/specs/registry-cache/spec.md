# registry-cache Specification

## Purpose
TBD - created by archiving change read-the-registry. Update Purpose after archive.

## Requirements

### Requirement: The cache is shown first and refreshed behind it

On start ossm SHALL read the cached registry and present it without waiting for
the network. When the cache is older than the configured time to live, a fresh
copy SHALL be fetched in the background and the list SHALL update when it
arrives.

#### Scenario: A fresh cache

- **WHEN** the cache is younger than the time to live
- **THEN** it is presented and no fetch is made

#### Scenario: A stale cache

- **WHEN** the cache is older than the time to live
- **THEN** the cached list is presented immediately and a fetch runs behind it,
  replacing the list when it succeeds

#### Scenario: A time to live of zero

- **WHEN** the time to live is zero
- **THEN** every start fetches, and the cached list is still shown first

### Requirement: A refresh can always be asked for

`r` SHALL fetch a fresh copy whatever the cache age says, and SHALL report while
it is running and when it finishes.

#### Scenario: Refresh is pressed on a fresh cache

- **WHEN** `r` is pressed and the cache is younger than the time to live
- **THEN** a fetch runs anyway

#### Scenario: Refresh fails

- **WHEN** a forced refresh cannot reach the registry
- **THEN** the failure is reported in the status line and the list keeps showing
  the cached copy

### Requirement: A bad response never destroys a good cache

A fetch SHALL be written to the cache only after it parses and validates. A
response that is unreadable, truncated or not a registry SHALL leave the
existing cache untouched.

#### Scenario: The server returns something unparseable

- **WHEN** a fetch returns a body that is not valid JSON
- **THEN** the cache file is unchanged and the failure is reported

#### Scenario: The server returns an error status

- **WHEN** a fetch returns a non-success HTTP status
- **THEN** the cache file is unchanged and the status is reported

#### Scenario: The response is a valid registry

- **WHEN** a fetch returns a registry that parses and validates
- **THEN** it replaces the cache, and the replacement is atomic so an
  interrupted write cannot leave a half-written cache

### Requirement: Offline is a working state, not an error

ossm SHALL work from cache with no network at all, and SHALL report the cache's
age so a user can judge it.

#### Scenario: Cached and offline

- **WHEN** there is a cache and the network is unreachable
- **THEN** the list works, and the status line says how old the cache is

#### Scenario: No cache and offline

- **WHEN** there is no cache and the network is unreachable
- **THEN** the tab says the registry has never been fetched and that `r` will
  try again, rather than showing an empty list

### Requirement: The cache records when it was fetched

The cache SHALL record when it was written, so that age can be reported and the
time to live can be judged without depending on the file's modification time.

#### Scenario: Age is reported

- **WHEN** a cached registry is presented
- **THEN** its age is derived from the recorded fetch time

#### Scenario: The cache is corrupt

- **WHEN** the cache file cannot be read or parsed
- **THEN** ossm behaves as though there were no cache, and says so

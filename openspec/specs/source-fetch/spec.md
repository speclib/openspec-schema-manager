# source-fetch Specification

## Purpose
TBD - created by archiving change show-a-schema-in-detail. Update Purpose after archive.

## Requirements

### Requirement: A schema folder is fetched from the repository, path and ref its entry names

Fetching SHALL take a repository, a directory inside it and an optional ref, and
SHALL yield the contents of that directory. The directory named by the entry
SHALL be used; it SHALL NOT be derived from the schema's name.

#### Scenario: An entry naming a path

- **WHEN** an entry names a repository and a path inside it
- **THEN** that directory's contents are fetched, including `schema.yaml` and
  everything under `templates/`

#### Scenario: A path that does not match the name

- **WHEN** an entry declares a name that differs from the last segment of its
  path
- **THEN** the path is used, because a schema's directory is not always its name

#### Scenario: The path does not exist in the repository

- **WHEN** the named directory is absent from the fetched repository
- **THEN** fetching fails saying which path was not found in which repository

### Requirement: An absent ref means the repository's default branch

When an entry names no ref, fetching SHALL resolve the repository's own default
branch rather than assuming a name for it.

#### Scenario: A repository whose default branch is not main

- **WHEN** an entry with no ref points at a repository whose default branch is
  `master`
- **THEN** it is fetched successfully

#### Scenario: A ref is named

- **WHEN** an entry names a tag, a branch or a commit
- **THEN** that ref is fetched, and a ref that does not resolve fails saying so

### Requirement: A fetched schema is cached by source, ref and path

A fetched folder SHALL be cached under the cache root, keyed by the repository,
the ref and the path together, so that two schemas from one repository and one
schema at two refs do not overwrite each other.

#### Scenario: The same schema is opened twice

- **WHEN** a schema is opened a second time
- **THEN** it is read from the cache and no network request is made

#### Scenario: Two schemas from one repository

- **WHEN** two entries name the same repository and different paths
- **THEN** each is cached separately and neither replaces the other

#### Scenario: One schema at two refs

- **WHEN** the same repository and path are fetched at two different refs
- **THEN** both are cached and both remain readable

#### Scenario: A refetch is asked for

- **WHEN** a refetch is asked for
- **THEN** the source is fetched again and the cached copy is replaced

### Requirement: A failed fetch does not destroy a cached copy

A fetch SHALL be assembled somewhere other than its final cache location and
moved into place only once it holds a readable schema. A failure SHALL leave any
previously cached copy untouched.

#### Scenario: The fetch fails partway

- **WHEN** a fetch fails after starting
- **THEN** the previously cached copy is still readable and nothing partial is
  left in the cache

#### Scenario: The fetched directory holds no schema

- **WHEN** the fetched directory holds no `schema.yaml`
- **THEN** the fetch is reported as failed and the cache is unchanged

### Requirement: Fetching needs git and says so when it is missing

Fetching SHALL use `git`. When `git` cannot be run, fetching SHALL fail with a
message naming that as the reason, and the rest of ossm SHALL keep working.

#### Scenario: git is not on PATH

- **WHEN** a fetch is attempted and `git` cannot be found
- **THEN** it fails saying `git` is needed to fetch a schema, and the registry
  list and every schema already on disk still work

#### Scenario: The network is unreachable

- **WHEN** a fetch is attempted with no network
- **THEN** it fails with what git reported, and any cached copy is still
  readable

### Requirement: A schema already on disk is not fetched

A schema resolved from a local directory or reported by OpenSpec as built in
SHALL be read from where it is. Fetching SHALL NOT be attempted for it.

#### Scenario: A built-in schema is opened

- **WHEN** a schema OpenSpec reports as built in is opened
- **THEN** it is read from the path OpenSpec gave, and no fetch is attempted

#### Scenario: A local schema is opened

- **WHEN** a schema in a local directory is opened
- **THEN** it is read from that directory

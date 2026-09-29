# registry-model Specification

## Purpose
TBD - created by archiving change read-the-registry. Update Purpose after archive.

## Requirements

### Requirement: An entry carries six required fields

An entry SHALL carry `id`, `name`, `description`, `artifacts`, `source.repo` and
`source.path`. An entry missing any of them SHALL be rejected, naming the entry
by `id` where it has one and by its position otherwise.

#### Scenario: A complete minimal entry

- **WHEN** an entry carries exactly the six required fields
- **THEN** it is read successfully, and every optional field holds its documented
  meaning when absent

#### Scenario: A required field is missing

- **WHEN** an entry has no `source.path`
- **THEN** reading fails naming that entry's `id` and the missing field

#### Scenario: An entry has no id to name

- **WHEN** an entry is missing `id` itself
- **THEN** reading fails naming the entry's position in the file

### Requirement: An absent optional field has a defined meaning

`source.ref` absent SHALL mean the repository's default branch. `status` absent
SHALL mean `active`. `language` absent SHALL mean `en`. `license` absent SHALL
mean the licence is unknown, which is not the same as unlicensed.
`requires.openspec` absent SHALL mean the entry states no minimum version.

#### Scenario: No ref is given

- **WHEN** an entry has no `source.ref`
- **THEN** it is reported as tracking the repository's default branch rather than
  as pinned, and no branch name is assumed

#### Scenario: No status is given

- **WHEN** an entry has no `status`
- **THEN** it is active

#### Scenario: No licence is given

- **WHEN** an entry has no `license`
- **THEN** the licence is reported as unknown, and not as absent or as public
  domain

### Requirement: An unknown field is ignored rather than rejected

An entry carrying a field ossm does not know SHALL be read, and the unknown
field SHALL be ignored. The registry may add a field before ossm knows about it,
and a reader that refuses the whole file over one unknown key strands every user
until they upgrade.

#### Scenario: A newer registry is read by an older ossm

- **WHEN** an entry carries a field ossm has no name for
- **THEN** the entry is read, its known fields take effect, and nothing is
  reported as an error

### Requirement: A deprecated entry stays visible and names its replacement

An entry with `status: deprecated` SHALL be read and presented as deprecated. An
entry naming a `superseded_by` SHALL report which entry replaces it.

#### Scenario: A deprecated entry

- **WHEN** an entry is deprecated
- **THEN** it appears in the list marked as deprecated, rather than being hidden

#### Scenario: A deprecated entry names its replacement

- **WHEN** a deprecated entry carries `superseded_by`
- **THEN** the replacement's `id` is shown with it

### Requirement: A pin is distinguished from a tracked branch

An entry SHALL report whether its source pins a ref or tracks the repository's
default branch, because the two mean different things to anyone installing it.

#### Scenario: The entry pins a tag

- **WHEN** an entry has `source.ref` set to a tag or a commit
- **THEN** it is reported as pinned, with the ref shown

#### Scenario: The entry tracks the default branch

- **WHEN** an entry has no `source.ref`
- **THEN** it is reported as tracking the default branch, and not as pinned to
  `main`

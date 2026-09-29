# schema-listing Specification

## Purpose
TBD - created by archiving change read-the-registry. Update Purpose after archive.

## Requirements

### Requirement: Built-in schemas are merged into the list

The schemas OpenSpec reports as built in SHALL appear in the same list as
registry entries, marked as built in. The registry deliberately omits them
because they are present without being installed, so a list showing only the
registry is not a list of what a user can choose.

#### Scenario: OpenSpec reports a built-in schema

- **WHEN** OpenSpec reports `spec-driven` as a built-in schema
- **THEN** it appears in the list marked as built in

#### Scenario: The openspec CLI is not installed

- **WHEN** the `openspec` CLI cannot be run
- **THEN** the registry entries are still listed, and the absence of the
  built-in schemas is reported rather than passed over

#### Scenario: A built-in schema shares a name with a registry entry

- **WHEN** a built-in schema and a registry entry declare the same name
- **THEN** both are listed and each says where it comes from, because they are
  different schemas that happen to share a name

### Requirement: A built-in schema cannot be installed

A built-in schema SHALL be presented as already available, and the install
action SHALL be unavailable for it with a reason given.

#### Scenario: Install is pressed on a built-in schema

- **WHEN** install is pressed on a schema marked as built in
- **THEN** it is refused with a message saying the schema ships with OpenSpec
  and needs no installing

### Requirement: Each row says enough to compare without opening it

A row SHALL carry the schema's name, how many artifacts it declares, the shape
of its workflow, and where it comes from. A registry row SHALL also say whether
it pins a ref or tracks a branch.

#### Scenario: A registry row

- **WHEN** a registry entry is listed
- **THEN** the row carries its name, its artifact count, the shape of its
  workflow, and its ref or that it tracks the default branch

#### Scenario: A built-in row

- **WHEN** a built-in schema is listed
- **THEN** the row says it is built in, and carries no ref, because there is no
  source to pin

#### Scenario: A deprecated row

- **WHEN** a deprecated entry is listed
- **THEN** the row says so, and names the replacement when the entry has one

### Requirement: The list has a stable order

The list SHALL be ordered so that the same registry and the same built-in
schemas always produce the same order, independent of the order they were read
in.

#### Scenario: The list is drawn twice

- **WHEN** the same registry is listed twice
- **THEN** the rows appear in the same order both times

#### Scenario: Built-in and registry schemas together

- **WHEN** built-in schemas are merged in
- **THEN** their position in the order is determined by the same rule as every
  other row, not by when they happened to be fetched

### Requirement: A listed schema can be opened

`enter` on a listed schema SHALL open its detail view. Every row SHALL be
openable, whether it comes from the registry, from OpenSpec's built-in schemas
or from a local directory, because reading a schema is not the same as
installing it.

#### Scenario: A registry row is opened

- **WHEN** `enter` is pressed on a registry row
- **THEN** its detail view opens, fetching the source if it is not cached

#### Scenario: A built-in row is opened

- **WHEN** `enter` is pressed on a built-in row
- **THEN** its detail view opens, read from the path OpenSpec reported, with no
  fetch attempted

#### Scenario: Nothing is selected

- **WHEN** `enter` is pressed with an empty list
- **THEN** nothing happens and nothing is reported as an error

# acceptance

## ADDED Requirements

### Requirement: Every acceptance criterion is a checkable statement

Each criterion the briefing lists SHALL be an end to end case driving the built
binary. A criterion SHALL NOT be recorded as met by inspection.

#### Scenario: The acceptance suite runs

- **WHEN** the acceptance suite runs
- **THEN** each of the briefing's criteria is exercised against the binary and
  passes or fails on its own

#### Scenario: A criterion cannot be checked

- **WHEN** a criterion cannot be checked in the environment the suite runs in
- **THEN** that case skips with the reason, rather than passing silently

### Requirement: The criteria

The suite SHALL check that: ossm outside a project shows the registry and local
schemas and refuses to install; ossm inside a project shows its default schema,
its available schemas with their origin and each change with its schema; opening
a registry schema shows its artifacts and a readable diagram and works offline
once cached; installing a registry schema into a project results in OpenSpec
seeing it and a change being creatable with it; duplicating a schema, editing a
template and returning shows the updated content with validation results; and
the composer can build a schema from two fixtures, catch a cycle, warn about a
dangling template reference and write a schema OpenSpec accepts.

#### Scenario: Outside a project

- **WHEN** ossm runs outside any OpenSpec project
- **THEN** the registry and local schemas are shown, and pressing install
  explains that installing needs a project

#### Scenario: Inside a project

- **WHEN** ossm runs inside an OpenSpec project
- **THEN** the default schema, the available schemas with their origins, and
  each change with its schema are shown

#### Scenario: Reading a schema offline

- **WHEN** a registry schema is opened, ossm is closed, the source is made
  unreachable and ossm is opened again
- **THEN** the schema still opens with its artifacts and its diagram

#### Scenario: Installing

- **WHEN** a registry schema is installed into a project
- **THEN** `openspec schema which` resolves it from the project and a change can
  be created with it

#### Scenario: Editing

- **WHEN** a schema is duplicated, a template is edited and the editor exits
- **THEN** the updated content and the validation result are shown

#### Scenario: Composing

- **WHEN** a schema is composed from artifacts of two fixtures
- **THEN** a cycle is caught, a dangling template reference is warned about, and
  the written schema is one OpenSpec accepts and can create a change with

### Requirement: The unit suite makes no network request

Every test outside the acceptance suite SHALL run with no network. The
acceptance suite SHALL use fixtures served from the filesystem rather than
reaching the real registry.

#### Scenario: The suite runs with no network

- **WHEN** the whole suite runs with no network available
- **THEN** it passes

# openspec-adapter

## ADDED Requirements

### Requirement: Everything that runs the openspec CLI lives behind one interface

Every call to the `openspec` command SHALL go through one interface with a fake
for tests. No other package SHALL run it.

#### Scenario: A screen needs something from OpenSpec

- **WHEN** a screen needs the project's schemas or changes
- **THEN** it asks the adapter, and the test for that screen uses the fake

#### Scenario: OpenSpec grows a command

- **WHEN** OpenSpec adds a command that replaces something ossm does by hand
- **THEN** only the adapter changes

### Requirement: ossm keeps no record of what is installed

ossm SHALL NOT write a lockfile, a manifest or any other record of installed
schemas. What a project has SHALL be answered by OpenSpec and by the project's
own files, every time it is asked.

#### Scenario: The project is read

- **WHEN** the project view is drawn
- **THEN** the schemas and changes come from OpenSpec and the project's files,
  and no file of ossm's own is read or written

#### Scenario: Something changes outside ossm

- **WHEN** a schema is added or removed while ossm is running, and the view is
  refreshed
- **THEN** the change is reflected, because nothing was cached

### Requirement: The default schema is read from the project's configuration

The project's default schema SHALL be read from `openspec/config.yaml`. No
OpenSpec command reports it for a project with no changes, and a project with no
changes is exactly where a user is deciding which schema to adopt.

#### Scenario: A project with no changes

- **WHEN** the project has no changes at all
- **THEN** its default schema is still reported

#### Scenario: The configuration names a schema

- **WHEN** `openspec/config.yaml` sets `schema: minimalist`
- **THEN** `minimalist` is reported as the default

#### Scenario: The configuration is missing or unreadable

- **WHEN** `openspec/config.yaml` is absent or cannot be parsed
- **THEN** the default is reported as unknown rather than guessed at

### Requirement: The adapter reports what each change uses

For each change the adapter SHALL report its name, the schema it uses, and how
many of its tasks are done.

#### Scenario: Changes are listed

- **WHEN** a project holds changes
- **THEN** each is reported with its name, its schema and its task progress

#### Scenario: A change whose schema cannot be determined

- **WHEN** a change's schema cannot be read
- **THEN** it is listed with its schema reported as unknown, rather than being
  left out

#### Scenario: A project with no changes

- **WHEN** a project holds no changes
- **THEN** an empty list is reported, not an error

### Requirement: A missing or failing CLI is reported, never fatal

When the `openspec` CLI is absent or a command fails, the adapter SHALL report
it as such and ossm SHALL keep working with everything that does not depend on
it.

#### Scenario: The CLI is not installed

- **WHEN** `openspec` is not on PATH
- **THEN** the project view says so and the registry and local tabs still work

#### Scenario: A command fails

- **WHEN** an `openspec` command exits non-zero
- **THEN** what it wrote to standard error is carried into the message

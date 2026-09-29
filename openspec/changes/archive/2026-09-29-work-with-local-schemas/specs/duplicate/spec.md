# duplicate

## ADDED Requirements

### Requirement: Any schema can be duplicated under a new name

`c` SHALL copy the selected schema into a chosen local schemas directory under a
new name, whatever the schema's origin. A registry schema, a built-in schema and
a local schema SHALL all be duplicable.

#### Scenario: A registry schema is duplicated

- **WHEN** `c` is pressed on a registry schema and a name is given
- **THEN** its fetched folder is copied into the chosen directory under that
  name

#### Scenario: A built-in schema is duplicated

- **WHEN** `c` is pressed on a schema OpenSpec reports as built in
- **THEN** it is copied from the path OpenSpec gave, and the copy is writable
  even when the original was not

#### Scenario: A local schema is duplicated

- **WHEN** `c` is pressed on a local schema
- **THEN** it is copied into the chosen directory under the new name

### Requirement: The copy declares its new name

The copied `schema.yaml` SHALL declare the new name, and nothing else in it
SHALL change.

#### Scenario: The name is rewritten

- **WHEN** a schema named `minimalist` is duplicated as `team-review`
- **THEN** the copy declares `name: team-review`

#### Scenario: Everything else survives

- **WHEN** a schema carrying comments, instructions and artifact order is
  duplicated
- **THEN** all of them are present in the copy unchanged

### Requirement: Duplicating shows where it will write and refuses to overwrite

Duplicating SHALL show the destination before writing, and SHALL refuse when
that directory already exists.

#### Scenario: The destination is shown

- **WHEN** a name is entered
- **THEN** the full destination path is shown before anything is written

#### Scenario: The destination is occupied

- **WHEN** the destination already exists
- **THEN** duplicating is refused, saying what is there, and nothing is written

#### Scenario: No local directory is configured

- **WHEN** no local schemas directory is configured
- **THEN** duplicating says so and says how to configure one

### Requirement: A name that cannot be a directory is refused

A new name SHALL be refused when it is empty, when it contains a path separator,
or when it is `.` or `..`, because the name becomes a directory name.

#### Scenario: An empty name

- **WHEN** an empty name is entered
- **THEN** it is refused and nothing is written

#### Scenario: A name with a separator

- **WHEN** a name containing `/` is entered
- **THEN** it is refused, because the name becomes one directory rather than a
  path

#### Scenario: A name that would escape the directory

- **WHEN** `..` is entered
- **THEN** it is refused

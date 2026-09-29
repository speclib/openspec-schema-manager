# update-detection

## ADDED Requirements

### Requirement: An installed schema can be compared against a registry entry

An installed schema SHALL be comparable against a registry entry declaring the
same name, by fetching that entry's source and comparing the files. Comparison
SHALL be asked for rather than run whenever the project is read, because it
fetches.

#### Scenario: The files are the same

- **WHEN** an installed schema's files match the entry's source
- **THEN** it is reported as identical to what the registry currently offers

#### Scenario: The files differ

- **WHEN** they differ
- **THEN** it is reported as differing, naming how many files differ and which

#### Scenario: A file is only in one of them

- **WHEN** the source holds a file the installed copy does not, or the other way
  round
- **THEN** that file is named as added or removed rather than as changed

#### Scenario: Comparison is not asked for

- **WHEN** the project view is drawn or refreshed
- **THEN** no comparison is made and nothing is fetched

### Requirement: A difference is never reported as an update

A difference SHALL be reported as a difference. ossm SHALL NOT say an update is
available, and SHALL say why it cannot.

#### Scenario: A difference is shown

- **WHEN** a difference is reported
- **THEN** the message says ossm cannot tell an upstream change from a local
  edit, because OpenSpec records nothing about where an installed schema came
  from

#### Scenario: The entry pins a ref

- **WHEN** the registry entry pins a ref
- **THEN** the comparison is against that ref, and the message says which

#### Scenario: The entry tracks a branch

- **WHEN** the entry names no ref
- **THEN** the comparison is against the repository's default branch, and the
  message says so rather than naming a branch

### Requirement: What cannot be compared says so

Comparison SHALL report plainly when there is nothing in the registry to compare
against, when the source cannot be reached, and when the schema is built into
OpenSpec.

#### Scenario: Nothing in the registry declares that name

- **WHEN** no registry entry declares the installed schema's name
- **THEN** it says there is nothing to compare against, rather than reporting no
  difference

#### Scenario: Several entries declare that name

- **WHEN** more than one registry entry declares the name
- **THEN** it says so and names them, because `name` is not unique across the
  registry and ossm cannot tell which one was installed

#### Scenario: The source cannot be reached

- **WHEN** the source cannot be fetched
- **THEN** it says the comparison could not be made and why

#### Scenario: A built-in schema

- **WHEN** a schema OpenSpec ships is compared
- **THEN** it says a built-in schema has no registry source to compare against

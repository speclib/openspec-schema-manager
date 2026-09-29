# schema-validation Specification

## Purpose
TBD - created by archiving change model-the-schema-and-its-graph. Update Purpose after archive.

## Requirements

### Requirement: Every problem is reported, not just the first

Validation SHALL report every problem it finds in one pass. A user fixing a
schema SHALL NOT have to run validation once per mistake.

#### Scenario: A schema with several problems

- **WHEN** a schema has a duplicate artifact id and an unknown id in `requires`
- **THEN** both are reported, each naming the artifact it applies to

#### Scenario: A valid schema

- **WHEN** a schema has no problems
- **THEN** validation reports none

### Requirement: What makes a schema unusable

Validation SHALL report as fatal: an empty artifact list, a duplicate artifact
id, an artifact id that is empty, an unknown artifact id in `requires`, an
artifact requiring itself, a cycle among `requires` edges, an unknown artifact
id in `apply.requires`, an empty `apply.requires`, and an absent `apply.tracks`.

#### Scenario: An unknown requirement

- **WHEN** an artifact requires an id no artifact declares
- **THEN** it is reported as fatal, naming both the artifact and the id it could
  not resolve

#### Scenario: A cycle

- **WHEN** two artifacts require each other, directly or through others
- **THEN** it is reported as fatal, naming the artifacts on the cycle in the
  order they connect

#### Scenario: An artifact requiring itself

- **WHEN** an artifact names its own id in `requires`
- **THEN** it is reported as fatal and named as a self-reference rather than as
  a cycle, because that is what a reader is looking at

#### Scenario: Nothing gates apply

- **WHEN** `apply.requires` is empty or absent
- **THEN** it is reported as fatal, because a schema whose implementation is
  gated by nothing has no workflow

#### Scenario: No tracked file

- **WHEN** `apply.tracks` is absent
- **THEN** it is reported as fatal

### Requirement: What makes a schema suspicious

Validation SHALL report as a warning: two artifacts declaring the same
`generates` pattern, an artifact declaring no template, and an artifact nothing
requires that does not gate apply. A warning SHALL NOT make a schema invalid.

#### Scenario: Two artifacts generating the same path

- **WHEN** two artifacts declare the same `generates`
- **THEN** it is reported as a warning naming both, because one will overwrite
  the other's output

#### Scenario: An artifact nothing reaches

- **WHEN** an artifact is required by nothing and does not gate apply
- **THEN** it is reported as a warning, because it will never be produced by the
  workflow it sits in

#### Scenario: A schema with only warnings

- **WHEN** a schema has warnings and no fatal problems
- **THEN** it is valid, and the warnings are reported alongside

### Requirement: A missing template is checked only where there are files

A missing template file SHALL be reported when validation is given a directory
to look in, and SHALL NOT be reported when it is not. A schema described by a
registry entry has no files yet, and reporting every template as missing would
make the check worthless.

#### Scenario: Validating a folder on disk

- **WHEN** a schema folder is validated and an artifact's template file does not
  exist under `templates/`
- **THEN** it is reported as fatal, naming the artifact and the path it looked
  for

#### Scenario: Validating a parsed schema with no folder

- **WHEN** a schema is validated with no directory
- **THEN** no template file is reported as missing, and every other check still
  runs

#### Scenario: The template exists

- **WHEN** an artifact's template file exists under `templates/`
- **THEN** nothing is reported for it

### Requirement: A problem names where it is

Every reported problem SHALL name the artifact it applies to, or the schema when
it applies to the schema as a whole, and SHALL state the rule rather than only
that something is wrong.

#### Scenario: A problem is read by someone fixing it

- **WHEN** any problem is reported
- **THEN** it names the artifact or the schema, states what is wrong, and gives
  the offending value where there is one

# composer-write Specification

## Purpose
TBD - created by archiving change compose-a-schema. Update Purpose after archive.

## Requirements

### Requirement: Writing produces a schema folder OpenSpec accepts

`w` SHALL write the composition as a schema folder into a local schemas
directory: a generated `schema.yaml` and each artifact's template copied from
the schema it came from. The result SHALL be a schema OpenSpec validates.

#### Scenario: A composition is written

- **WHEN** a valid composition is written under a name
- **THEN** the destination holds `schema.yaml` and a `templates/` tree, and
  OpenSpec's own validation accepts it

#### Scenario: The generated schema declares what the canvas holds

- **WHEN** a composition is written
- **THEN** `schema.yaml` declares the new name, every artifact with its id,
  generated path, description, template and requirements, and the apply gate and
  tracked file

#### Scenario: Templates are copied unchanged

- **WHEN** an artifact's template is copied
- **THEN** it is byte for byte what the source held

#### Scenario: Two artifacts whose templates share a path

- **WHEN** two artifacts from different schemas declare the same template path
- **THEN** each is written to a path that does not overwrite the other, and each
  artifact's `template` names its own

### Requirement: The written schema records where each artifact came from

The generated `schema.yaml` SHALL open with a comment block naming, for each
artifact, the schema and ref it came from and the id it had there.

#### Scenario: The provenance block

- **WHEN** a composition drawn from two schemas is written
- **THEN** the file opens with a comment naming both schemas, their refs, and
  which artifacts came from each

#### Scenario: A renamed artifact

- **WHEN** an artifact was renamed on the canvas
- **THEN** the comment records the id it had in its source schema as well as the
  one it has now

### Requirement: Writing refuses rather than overwriting

Writing SHALL refuse when the destination already exists, when the composition
has a fatal problem, when the name cannot be a directory, or when no local
schemas directory is configured.

#### Scenario: The destination exists

- **WHEN** the destination already exists
- **THEN** writing is refused, saying so, and nothing is written

#### Scenario: The composition is not valid

- **WHEN** the composition has a fatal problem
- **THEN** writing is refused, naming the problem, and nothing is written

#### Scenario: No local directory is configured

- **WHEN** no local schemas directory is configured
- **THEN** writing says so and says how to configure one

#### Scenario: A write that fails partway

- **WHEN** writing fails partway
- **THEN** nothing partial is left at the destination

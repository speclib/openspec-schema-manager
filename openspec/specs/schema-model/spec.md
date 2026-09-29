# schema-model Specification

## Purpose
TBD - created by archiving change model-the-schema-and-its-graph. Update Purpose after archive.

## Requirements

### Requirement: A schema carries a name, a version, a description and artifacts

A schema SHALL carry `name`, `version`, `description`, an ordered list of
`artifacts` and an `apply` block. `name` SHALL be preserved with the casing it
was written in, because it decides the directory the schema installs into and
the value passed to `--schema`.

#### Scenario: A schema is parsed

- **WHEN** a `schema.yaml` declaring a name, a version, a description, artifacts
  and an apply block is parsed
- **THEN** all five are available, and the artifacts are in the order they were
  written in

#### Scenario: Upstream casing is preserved

- **WHEN** a schema declares `name: SuperSpec`
- **THEN** the name reads `SuperSpec`, not a lowercased form

#### Scenario: A description is absent

- **WHEN** a schema declares no description
- **THEN** it parses, and the description is empty rather than invented

### Requirement: An artifact carries what it generates and what it needs first

An artifact SHALL carry `id`, `generates`, `template`, an optional
`description`, an optional `instruction`, and `requires` naming the artifact ids
that must exist before it. `requires` absent SHALL mean the same as an empty
list.

#### Scenario: A full artifact

- **WHEN** an artifact declares an id, what it generates, a template, a
  description, an instruction and its requirements
- **THEN** all six are available

#### Scenario: An artifact with no requirements

- **WHEN** an artifact declares `requires: []` or omits `requires`
- **THEN** it has no requirements in both cases, and the two forms are
  indistinguishable afterwards

#### Scenario: A generated path is a glob

- **WHEN** an artifact declares `generates: specs/**/*.md`
- **THEN** the pattern is kept as written, because it describes a set of files
  rather than one

### Requirement: The apply block names the gate and the tracked file

The `apply` block SHALL carry `requires`, the artifact ids that gate
implementation, `tracks`, the file that records progress, and an optional
`instruction`. An artifact named in `apply.requires` SHALL be reportable as a
gate.

#### Scenario: The gate is read

- **WHEN** a schema declares `apply.requires: [tasks]` and `apply.tracks:
  tasks.md`
- **THEN** `tasks` is reported as an apply gate and `tasks.md` as the tracked
  file

#### Scenario: Several artifacts gate apply

- **WHEN** `apply.requires` names more than one artifact
- **THEN** each of them is reported as a gate

### Requirement: An instruction is text, never an instruction to ossm

An `instruction` block SHALL be carried as text and SHALL NOT influence what
ossm does. A schema comes from a third party.

#### Scenario: An instruction contains something that reads like a command

- **WHEN** an artifact's instruction contains text that reads as a directive
- **THEN** it is stored and displayed as text, and changes nothing about how
  ossm parses, validates or renders the schema

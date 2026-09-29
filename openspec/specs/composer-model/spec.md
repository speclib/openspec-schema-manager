# composer-model Specification

## Purpose
TBD - created by archiving change compose-a-schema. Update Purpose after archive.

## Requirements

### Requirement: A composition is a set of artifacts drawn from several schemas

A composition SHALL hold artifacts copied from one or more source schemas, the
`requires` edges between them, which of them gate apply, and the tracked file.
Each artifact SHALL remember the schema it came from and the ref that schema was
read at.

#### Scenario: An artifact is added

- **WHEN** an artifact is added from a source schema
- **THEN** the composition holds it with everything it declared, and remembers
  which schema and ref it came from

#### Scenario: Artifacts from two schemas

- **WHEN** artifacts are added from two different schemas
- **THEN** each remembers its own source, so the written schema can record both

#### Scenario: A source is added twice

- **WHEN** the same source schema is added as a source twice
- **THEN** it appears once in the palette

### Requirement: An artifact arrives with nothing it cannot resolve

An artifact SHALL be added without the requirements that name artifacts not on
the canvas. Adding an artifact SHALL NOT make the composition invalid on its
own.

#### Scenario: An artifact requiring something not yet added

- **WHEN** an artifact requiring `proposal` is added and `proposal` is not on
  the canvas
- **THEN** it is added with no requirement on `proposal`, and the composition
  does not report an unresolved requirement

#### Scenario: An artifact requiring something already there

- **WHEN** an artifact requiring `proposal` is added and `proposal` is on the
  canvas
- **THEN** the requirement is kept and the edge appears

#### Scenario: A requirement added later

- **WHEN** `proposal` is added after an artifact that required it
- **THEN** no edge appears on its own, because the user chose what to link

### Requirement: An id collision is resolved before the artifact is added

An artifact whose id is already on the canvas SHALL NOT be added until a new id
is given. The new id SHALL be checked the same way.

#### Scenario: Two sources declare the same id

- **WHEN** an artifact whose id is already on the canvas is added
- **THEN** the composition reports the collision and adds nothing until a new id
  is given

#### Scenario: A new id is given

- **WHEN** a free id is given
- **THEN** the artifact is added under it, and it remembers the id it had in its
  source schema

#### Scenario: The new id also collides

- **WHEN** the given id is also taken
- **THEN** it is refused again rather than overwriting

### Requirement: A composition can start from an existing schema

A composition SHALL be able to start empty or from an existing schema, taking
its artifacts, edges, gates and tracked file.

#### Scenario: Starting empty

- **WHEN** a composition is started with no schema
- **THEN** it holds no artifacts and reports what is missing

#### Scenario: Starting from a schema

- **WHEN** a composition is started from a schema
- **THEN** it holds that schema's artifacts, edges, gates and tracked file, and
  each artifact remembers that schema as its source

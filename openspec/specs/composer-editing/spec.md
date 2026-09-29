# composer-editing Specification

## Purpose
TBD - created by archiving change compose-a-schema. Update Purpose after archive.

## Requirements

### Requirement: An artifact can be removed with its edges

`x` SHALL remove the selected artifact and every edge that names it, in either
direction, and SHALL remove it from the apply gate.

#### Scenario: An artifact in the middle of a chain is removed

- **WHEN** an artifact required by one artifact and requiring another is removed
- **THEN** both edges go with it, and no requirement is left naming it

#### Scenario: A gate is removed

- **WHEN** an artifact that gates apply is removed
- **THEN** it is no longer a gate, and the composition reports if nothing gates
  apply any more

### Requirement: Edges are added and removed from the keyboard

`l` SHALL start a link from the selected artifact, offer the artifacts it could
link to, and add a `requires` edge on confirmation. `u` SHALL remove an edge.

#### Scenario: A link is made

- **WHEN** a link is started from `tasks` and `specs` is chosen
- **THEN** `tasks` requires `specs`

#### Scenario: A link that already exists

- **WHEN** a link is made that already exists
- **THEN** nothing changes and no duplicate edge appears

#### Scenario: A link to itself

- **WHEN** an artifact is linked to itself
- **THEN** it is refused, because an artifact cannot require itself

#### Scenario: A link is abandoned

- **WHEN** a link is started and then cancelled
- **THEN** no edge is added

#### Scenario: An edge is removed

- **WHEN** an edge is removed
- **THEN** the requirement goes and the artifacts stay

### Requirement: The apply gate and the tracked file are set from the keyboard

`g` SHALL toggle the selected artifact as an apply gate. `t` SHALL set the
tracked file.

#### Scenario: A gate is added

- **WHEN** `g` is pressed on an artifact that is not a gate
- **THEN** it becomes one

#### Scenario: A gate is removed

- **WHEN** `g` is pressed on an artifact that is a gate
- **THEN** it stops being one

#### Scenario: The tracked file is set

- **WHEN** a tracked file is given
- **THEN** the composition records it and stops reporting that none is set

### Requirement: Renaming an artifact updates everything that names it

`R` SHALL rename the selected artifact, updating every `requires` entry and the
apply gate. A rename to an id already in use SHALL be refused.

#### Scenario: An artifact with dependents is renamed

- **WHEN** an artifact two others require is renamed
- **THEN** both requirements name the new id and none names the old one

#### Scenario: A gate is renamed

- **WHEN** an artifact that gates apply is renamed
- **THEN** the gate names the new id

#### Scenario: A rename that collides

- **WHEN** an artifact is renamed to an id already in use
- **THEN** it is refused and nothing changes

#### Scenario: A rename to the same id

- **WHEN** an artifact is renamed to the id it already has
- **THEN** nothing changes and it is not reported as a collision

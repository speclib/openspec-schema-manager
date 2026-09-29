# composer-validation

## ADDED Requirements

### Requirement: The composition is validated after every change

Validation SHALL run after every edit and report everything wrong at once: a
cycle, a requirement naming an artifact not on the canvas, no apply gate, no
tracked file, and two artifacts generating the same path.

#### Scenario: An edit creates a cycle

- **WHEN** a link is added that closes a cycle
- **THEN** the cycle is reported straight away, naming the artifacts on it

#### Scenario: An edit resolves a problem

- **WHEN** an edit resolves a reported problem
- **THEN** it stops being reported without the user asking

#### Scenario: Several problems at once

- **WHEN** a composition has no gate and two artifacts writing the same path
- **THEN** both are reported

#### Scenario: An empty composition

- **WHEN** the canvas is empty
- **THEN** it reports that it needs artifacts, a gate and a tracked file, rather
  than reporting nothing

### Requirement: A template that mentions something not on the canvas is a warning

After an artifact is added, its template text and its instruction SHALL be
scanned for references to artifact ids and generated file names that are not on
the canvas, and each SHALL be reported as a warning naming the file to edit.

#### Scenario: A template mentions a missing artifact

- **WHEN** an artifact's template mentions `design`, which is not on the canvas
- **THEN** a warning names the template file and the reference

#### Scenario: A template mentions a generated file that is not produced

- **WHEN** a template mentions `design.md` and nothing on the canvas generates it
- **THEN** a warning names it

#### Scenario: A reference that resolves

- **WHEN** a template mentions an artifact that is on the canvas
- **THEN** nothing is reported for it

#### Scenario: The missing artifact is added

- **WHEN** the artifact a warning named is added to the canvas
- **THEN** the warning goes

#### Scenario: A template that cannot be read

- **WHEN** a template file cannot be read
- **THEN** that is reported once and the other templates are still scanned

### Requirement: A warning does not stop a composition being written

Template reference warnings SHALL be advisory. A composition with warnings and
no fatal problems SHALL be writable.

#### Scenario: Writing with warnings

- **WHEN** a composition has template warnings and nothing fatal
- **THEN** it can be written, because a template mentioning an artifact the user
  deliberately left out is their decision

#### Scenario: Writing with a fatal problem

- **WHEN** a composition has a cycle
- **THEN** writing is refused, naming the problem

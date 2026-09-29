# schema-listing

## MODIFIED Requirements

### Requirement: A built-in schema cannot be installed

A schema OpenSpec already resolves SHALL be presented as already available, and
the install action SHALL be unavailable for it with a reason given. This covers
a schema built into OpenSpec and one already installed in the project: both are
already resolvable, and installing over them is a different act from installing.

#### Scenario: Install is pressed on a built-in schema

- **WHEN** install is pressed on a schema marked as built in
- **THEN** it is refused with a message saying the schema ships with OpenSpec
  and needs no installing

#### Scenario: Install is pressed on a schema already in the project

- **WHEN** install is pressed on a schema the project already resolves
- **THEN** it is refused with a message saying so, and overwriting is offered as
  its own choice rather than happening by default

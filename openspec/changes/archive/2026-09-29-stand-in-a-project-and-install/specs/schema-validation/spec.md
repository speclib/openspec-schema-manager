# schema-validation

## MODIFIED Requirements

### Requirement: What makes a schema unusable

Validation SHALL report as fatal: an empty artifact list, a duplicate artifact
id, an artifact id that is empty, an artifact with no description, an unknown
artifact id in `requires`, an artifact requiring itself, a cycle among
`requires` edges, an unknown artifact id in `apply.requires`, an empty
`apply.requires`, and an absent `apply.tracks`.

An artifact with no description is fatal because OpenSpec 1.10.0 rejects such a
schema outright. The briefing calls the field optional; the CLI decides what it
will accept, and a schema it rejects cannot be installed.

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

#### Scenario: An artifact with no description

- **WHEN** an artifact declares no description
- **THEN** it is reported as fatal, naming the artifact and saying that OpenSpec
  rejects a schema whose artifact has none

#### Scenario: Nothing gates apply

- **WHEN** `apply.requires` is empty or absent
- **THEN** it is reported as fatal, because a schema whose implementation is
  gated by nothing has no workflow

#### Scenario: No tracked file

- **WHEN** `apply.tracks` is absent
- **THEN** it is reported as fatal

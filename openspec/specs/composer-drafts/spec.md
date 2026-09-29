# composer-drafts Specification

## Purpose
TBD - created by archiving change compose-a-schema. Update Purpose after archive.

## Requirements

### Requirement: A composition can be saved and resumed

`s` SHALL save the composition as a draft in the state directory under a name. A
saved draft SHALL be listable and resumable, restoring the artifacts, edges,
gates and tracked file.

#### Scenario: A draft is saved and resumed

- **WHEN** a composition is saved and then resumed
- **THEN** the canvas holds the same artifacts, edges, gates and tracked file

#### Scenario: A draft is overwritten

- **WHEN** a draft is saved under a name already used
- **THEN** it replaces that draft, because saving again is what a user means by
  saving

#### Scenario: Drafts are listed

- **WHEN** drafts exist
- **THEN** they are listed by name, most recently saved first

#### Scenario: No drafts

- **WHEN** no drafts exist
- **THEN** the list says so rather than showing an empty pane

### Requirement: A draft that cannot be read does not stop the others

A draft file that cannot be parsed SHALL be reported and skipped, and the other
drafts SHALL still be listed.

#### Scenario: One draft is corrupt

- **WHEN** one draft file cannot be parsed
- **THEN** it is reported and the others are still listed and resumable

### Requirement: A draft remembers where its artifacts came from

A saved draft SHALL record each artifact's source schema directory, ref and
original id, so that resuming it can still copy the templates.

#### Scenario: A draft is resumed and written

- **WHEN** a draft is resumed and written
- **THEN** the templates are copied from the source schemas the draft recorded

#### Scenario: A source has gone

- **WHEN** a source directory a draft records no longer exists
- **THEN** resuming reports which artifacts cannot be written and the rest of
  the composition is still editable

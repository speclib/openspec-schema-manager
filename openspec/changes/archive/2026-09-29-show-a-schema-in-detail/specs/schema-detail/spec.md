# schema-detail

## ADDED Requirements

### Requirement: The detail view identifies the schema

The detail view SHALL show the schema's name, its version, its description, the
repository it comes from and the ref it was read at, and where it is on disk
when it is local.

#### Scenario: A registry schema is opened

- **WHEN** a registry schema is opened
- **THEN** the header carries its name, version, description, repository and the
  ref it was read at

#### Scenario: A schema tracking the default branch

- **WHEN** a schema whose entry names no ref is opened
- **THEN** the header says it tracks the default branch rather than naming a
  branch it does not have

#### Scenario: A local schema is opened

- **WHEN** a local or built-in schema is opened
- **THEN** the header carries its path instead of a repository and ref

### Requirement: The artifacts are shown as a table

The detail view SHALL list every artifact with its id, what it generates, its
template and what it requires, in the schema's declaration order.

#### Scenario: A schema with artifacts

- **WHEN** a schema is opened
- **THEN** every artifact appears with its id, generated path, template and
  requirements, in the order the schema declares them

#### Scenario: An artifact requiring nothing

- **WHEN** an artifact requires nothing
- **THEN** its requirements column says so rather than being blank

### Requirement: The apply gate and the figures are shown

The detail view SHALL show which artifacts gate apply and the file that tracks
progress, together with the artifact count, the longest dependency chain and the
number of gates.

#### Scenario: The gate is shown

- **WHEN** a schema is opened
- **THEN** the artifacts gating apply and the tracked file are named

#### Scenario: The figures are shown

- **WHEN** a schema is opened
- **THEN** the artifact count, the longest chain and the gate count are shown,
  so two schemas can be compared on them

### Requirement: Validation findings travel with the schema

The detail view SHALL show the schema's validation findings, separating fatal
problems from warnings. A schema that does not validate SHALL still be viewable.

#### Scenario: A schema with a fatal problem

- **WHEN** a schema with a cycle is opened
- **THEN** the view opens, shows what it can, and reports the problem as fatal

#### Scenario: A schema with warnings

- **WHEN** a schema with warnings and no fatal problems is opened
- **THEN** the warnings are shown and the schema is not presented as broken

#### Scenario: A schema with no findings

- **WHEN** a valid schema is opened
- **THEN** nothing is reported

### Requirement: The view is reached and left from the keyboard

`enter` on a listed schema SHALL open it. `esc` and `q` SHALL return to the list
with the same schema still selected. `R` SHALL refetch the source.

#### Scenario: A schema is opened and closed

- **WHEN** `enter` opens a schema and `esc` closes it
- **THEN** the list returns with the same row selected as before

#### Scenario: q closes the detail rather than quitting

- **WHEN** `q` is pressed on the detail view
- **THEN** the list returns and ossm keeps running

#### Scenario: A refetch is asked for

- **WHEN** `R` is pressed
- **THEN** the source is fetched again, the view reports while it runs, and the
  schema is redrawn from what arrives

### Requirement: Fetching is reported while it happens and when it fails

Opening a schema that is not cached SHALL report that it is being fetched, and a
failure SHALL be shown with its reason instead of an empty view.

#### Scenario: A schema is being fetched

- **WHEN** an uncached schema is opened
- **THEN** the view says it is being fetched and does not block the rest of the
  application

#### Scenario: The fetch fails

- **WHEN** the fetch fails
- **THEN** the view shows the reason and offers `R` to try again

### Requirement: The source repository can be read off the screen

`o` SHALL put the schema's source repository URL where it can be read and
copied, because ossm cannot open a browser on every machine it runs on.

#### Scenario: o is pressed on a registry schema

- **WHEN** `o` is pressed
- **THEN** the repository URL is shown in full

#### Scenario: o is pressed on a local schema

- **WHEN** `o` is pressed on a schema with no repository
- **THEN** its path is shown instead

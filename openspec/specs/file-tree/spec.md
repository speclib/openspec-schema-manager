# file-tree Specification

## Purpose
TBD - created by archiving change work-with-local-schemas. Update Purpose after archive.

## Requirements

### Requirement: A local schema's files are browsable

A local schema SHALL show its files as a tree: `schema.yaml` and everything
under `templates/`, in a stable order.

#### Scenario: A schema with templates

- **WHEN** a local schema is browsed
- **THEN** `schema.yaml` and each file under `templates/` are listed, nested by
  directory

#### Scenario: A schema with no templates directory

- **WHEN** a local schema has no `templates/` directory
- **THEN** `schema.yaml` is listed alone and the absence is not an error

#### Scenario: The order is stable

- **WHEN** the tree is drawn twice
- **THEN** the files appear in the same order

### Requirement: A file opens in the user's editor

`e` SHALL open the selected file in `$VISUAL`, else `$EDITOR`, else `vi`,
suspending the interface while the editor runs and resuming when it exits.

#### Scenario: VISUAL is set

- **WHEN** `$VISUAL` is set
- **THEN** it is used

#### Scenario: Only EDITOR is set

- **WHEN** `$VISUAL` is unset and `$EDITOR` is set
- **THEN** `$EDITOR` is used

#### Scenario: Neither is set

- **WHEN** neither is set
- **THEN** `vi` is used

#### Scenario: An editor command with arguments

- **WHEN** `$EDITOR` holds a command with arguments, such as `code --wait`
- **THEN** the command and its arguments are used as given

#### Scenario: The editor cannot be run

- **WHEN** the editor exits non-zero or cannot be started
- **THEN** it is reported and the interface resumes rather than ending

### Requirement: The schema is re-read after editing

On returning from the editor, the schema SHALL be parsed and validated again,
and the findings SHALL be shown against the tree.

#### Scenario: An edit that fixes a problem

- **WHEN** an edit resolves a validation finding and the editor exits
- **THEN** the finding is gone without the user asking for a refresh

#### Scenario: An edit that breaks the schema

- **WHEN** an edit makes `schema.yaml` unparseable
- **THEN** the error is shown against the tree and the tree is still usable

#### Scenario: An edit to a template

- **WHEN** a template file is edited
- **THEN** the schema is re-validated, so a template that was reported missing
  stops being reported

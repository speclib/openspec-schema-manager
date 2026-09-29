# path-prompt

## ADDED Requirements

### Requirement: Any folder can be opened by path

`:` and `ctrl+o` SHALL open a prompt taking a directory path. A directory
holding a `schema.yaml` SHALL be opened as a local schema.

#### Scenario: A path holding a schema

- **WHEN** a path holding a `schema.yaml` is entered
- **THEN** that schema is opened

#### Scenario: A path holding no schema

- **WHEN** a path with no `schema.yaml` is entered
- **THEN** the prompt says so and stays open

#### Scenario: A path that does not exist

- **WHEN** a path that does not exist is entered
- **THEN** the prompt says so and stays open

#### Scenario: The prompt is dismissed

- **WHEN** `esc` is pressed while the prompt is open
- **THEN** it closes and nothing is opened

### Requirement: The prompt completes a path

`tab` SHALL complete the path being typed against the directories that exist. A
leading `~/` SHALL expand to the user's home directory.

#### Scenario: One directory matches

- **WHEN** `tab` is pressed and exactly one directory matches what is typed
- **THEN** the path is completed to it

#### Scenario: Several directories match

- **WHEN** several directories match
- **THEN** the longest common prefix is filled in and the matches are shown

#### Scenario: Nothing matches

- **WHEN** nothing matches
- **THEN** the path is left as typed and nothing is reported as an error

#### Scenario: A leading tilde

- **WHEN** a path beginning `~/` is entered
- **THEN** it resolves inside the user's home directory

### Requirement: Opened paths are remembered

A path opened through the prompt or through `--path` SHALL be added to a recents
list kept in the state directory, most recent first, capped at `recents_cap`
and holding no duplicates.

#### Scenario: A path is opened

- **WHEN** a path is opened
- **THEN** it appears at the top of the recents list

#### Scenario: The same path is opened again

- **WHEN** a path already in the list is opened again
- **THEN** it moves to the top and appears once

#### Scenario: The cap is reached

- **WHEN** more paths are opened than the cap allows
- **THEN** the oldest fall off the end

#### Scenario: The recents file is missing or corrupt

- **WHEN** the recents file does not exist or cannot be parsed
- **THEN** the list is empty and opening a path still works and writes a new one

#### Scenario: A remembered path has gone

- **WHEN** a remembered path no longer holds a schema
- **THEN** it is shown as missing rather than being silently dropped

# project-view

## ADDED Requirements

### Requirement: An installed schema can be compared from the Project tab

`u` on a schema in the Project view SHALL compare it against the registry and
report the result in place, without changing anything.

#### Scenario: A comparison is asked for

- **WHEN** `u` is pressed on an installed schema
- **THEN** the comparison runs, reports while it runs, and shows its result

#### Scenario: A comparison changes nothing

- **WHEN** a comparison finishes, whatever it found
- **THEN** no file in the project has changed

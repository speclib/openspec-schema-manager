# schema-listing

## ADDED Requirements

### Requirement: A listed schema can be opened

`enter` on a listed schema SHALL open its detail view. Every row SHALL be
openable, whether it comes from the registry, from OpenSpec's built-in schemas
or from a local directory, because reading a schema is not the same as
installing it.

#### Scenario: A registry row is opened

- **WHEN** `enter` is pressed on a registry row
- **THEN** its detail view opens, fetching the source if it is not cached

#### Scenario: A built-in row is opened

- **WHEN** `enter` is pressed on a built-in row
- **THEN** its detail view opens, read from the path OpenSpec reported, with no
  fetch attempted

#### Scenario: Nothing is selected

- **WHEN** `enter` is pressed with an empty list
- **THEN** nothing happens and nothing is reported as an error

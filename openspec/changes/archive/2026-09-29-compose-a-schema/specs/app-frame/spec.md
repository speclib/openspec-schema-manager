# app-frame

## ADDED Requirements

### Requirement: A schema can be sent to the composer from any list

`p` on a listed schema SHALL add it to the composer's palette as a source,
without leaving the tab it was pressed on.

#### Scenario: A registry schema is added as a source

- **WHEN** `p` is pressed on a registry schema
- **THEN** it is fetched if needed and its artifacts appear in the composer's
  palette, and the Registry tab is still selected

#### Scenario: A local schema is added as a source

- **WHEN** `p` is pressed on a local schema
- **THEN** its artifacts appear in the palette

#### Scenario: The same schema is added twice

- **WHEN** `p` is pressed twice on the same schema
- **THEN** it appears once in the palette

#### Scenario: A schema that cannot be read

- **WHEN** `p` is pressed on a schema that cannot be read
- **THEN** the reason is reported and the palette is unchanged

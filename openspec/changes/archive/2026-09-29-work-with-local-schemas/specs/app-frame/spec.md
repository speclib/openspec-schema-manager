# app-frame

## ADDED Requirements

### Requirement: The path prompt is reachable from anywhere

`:` and `ctrl+o` SHALL open the path prompt from any tab. While it is open the
frame SHALL take `ctrl+c` and nothing else, and opening a schema SHALL move to
the Local tab.

#### Scenario: The prompt is opened from another tab

- **WHEN** `:` is pressed on the Registry tab
- **THEN** the path prompt opens

#### Scenario: A schema is opened from the prompt

- **WHEN** a path holding a schema is entered from any tab
- **THEN** the Local tab is selected and the schema is shown there

#### Scenario: A digit is typed into the prompt

- **WHEN** a digit is typed while the prompt is open
- **THEN** it goes into the path and no tab is selected

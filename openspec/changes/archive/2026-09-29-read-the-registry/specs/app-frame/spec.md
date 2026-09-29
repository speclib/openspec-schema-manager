# app-frame

## MODIFIED Requirements

### Requirement: Every tab says what it will hold

Until a tab's screen is built, it SHALL draw a placeholder naming what will live
there and which milestone builds it. A tab whose screen is built SHALL draw that
screen instead.

#### Scenario: An unbuilt tab is selected

- **WHEN** a tab whose screen is not yet built is selected
- **THEN** it names what will live there and the milestone that builds it,
  rather than showing an empty pane

#### Scenario: The Registry tab is selected

- **WHEN** the Registry tab is selected
- **THEN** it draws the list of schemas, and no placeholder

## ADDED Requirements

### Requirement: A screen may take a key the frame does not define

A screen SHALL receive every key the frame does not take for itself, and SHALL
report those keys through the help overlay. While a screen is capturing text,
the frame SHALL still take `ctrl+c`.

#### Scenario: The Registry tab takes a key

- **WHEN** `r` is pressed on the Registry tab
- **THEN** the Registry screen receives it and refreshes, and no other tab
  changes behaviour

#### Scenario: A screen is capturing text

- **WHEN** the filter is open and a letter is typed
- **THEN** it goes into the filter rather than selecting a tab or quitting

#### Scenario: Quitting while capturing text

- **WHEN** the filter is open and `ctrl+c` is pressed
- **THEN** the application exits

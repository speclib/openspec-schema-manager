# registry-search Specification

## Purpose
TBD - created by archiving change read-the-registry. Update Purpose after archive.

## Requirements

### Requirement: The filter matches id, name and description

`/` SHALL open a filter that narrows the list to entries whose `id`, `name` or
`description` matches what is typed. Matching SHALL ignore case.

#### Scenario: Matching a name

- **WHEN** the filter holds text appearing in an entry's name
- **THEN** that entry is listed

#### Scenario: Matching a description

- **WHEN** the filter holds text appearing only in an entry's description
- **THEN** that entry is listed, because a user searching for what a schema does
  does not know its name yet

#### Scenario: Matching regardless of case

- **WHEN** the filter holds text differing only in case from an entry's name
- **THEN** that entry is listed

#### Scenario: Matching nothing

- **WHEN** the filter matches no entry
- **THEN** the list says so, and says which text matched nothing

### Requirement: The filter is dismissed without losing the list

`esc` SHALL clear the filter and restore the full list. Closing the filter SHALL
never leave the list narrowed with no visible reason.

#### Scenario: The filter is cleared

- **WHEN** `esc` is pressed while filtering
- **THEN** the filter is emptied and every entry is listed again

#### Scenario: The filter is visible while it applies

- **WHEN** a filter is applied
- **THEN** the text being filtered on is drawn, so a narrowed list is never
  unexplained

### Requirement: The selection survives filtering where it can

When the filter changes, the selected entry SHALL stay selected if it still
matches, and SHALL move to the first match if it does not.

#### Scenario: The selection still matches

- **WHEN** the filter is narrowed and the selected entry still matches
- **THEN** it stays selected

#### Scenario: The selection no longer matches

- **WHEN** the filter is narrowed past the selected entry
- **THEN** the first matching entry is selected, and the list is never left with
  nothing selected while it holds rows

# project-view

## ADDED Requirements

### Requirement: Inside a project the tab shows what the project has

Inside an OpenSpec project the Project tab SHALL show the project root, the
default schema, every schema available with where each resolves from, and every
change with the schema it uses.

#### Scenario: A project with schemas and changes

- **WHEN** the Project tab is opened inside a project
- **THEN** it shows the root, the default schema, each available schema with its
  origin, and each change with its schema

#### Scenario: A schema shadows another

- **WHEN** a project-local schema shadows one of the same name from elsewhere
- **THEN** the view says which one wins and that it shadows another

#### Scenario: A project with no changes

- **WHEN** a project holds no changes
- **THEN** the view says so rather than showing an empty area

### Requirement: Outside a project the tab says what would make it work

Outside an OpenSpec project the Project tab SHALL say so and say what would
change that, rather than showing an empty view.

#### Scenario: Started outside a project

- **WHEN** the Project tab is opened outside any project
- **THEN** it says there is no project here, that browsing works anyway, and
  that installing needs one

### Requirement: The view refreshes from OpenSpec

`r` SHALL re-read the project from OpenSpec and the project's files. The view
SHALL also refresh after an install.

#### Scenario: A refresh is asked for

- **WHEN** `r` is pressed on the Project tab
- **THEN** the project is read again and the view redrawn

#### Scenario: A schema has just been installed

- **WHEN** an install finishes
- **THEN** the project view shows the new schema without the user asking

### Requirement: A schema in the project can be opened

`enter` on a schema in the project view SHALL open its detail view, read from
the path OpenSpec reported.

#### Scenario: A project schema is opened

- **WHEN** `enter` is pressed on a schema in the project view
- **THEN** its detail opens, read from where OpenSpec says it resolves, with no
  fetch attempted

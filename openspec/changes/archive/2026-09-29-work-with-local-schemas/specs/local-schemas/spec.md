# local-schemas

## ADDED Requirements

### Requirement: A local schema is a folder holding a schema.yaml

A directory named in `schemas_dirs` SHALL be scanned for folders holding a
`schema.yaml`, and each such folder SHALL be listed as a local schema. Scanning
SHALL NOT descend past the folders directly inside a configured directory.

#### Scenario: A configured directory holding schemas

- **WHEN** a configured directory holds three folders, two of which hold a
  `schema.yaml`
- **THEN** those two are listed and the third is not

#### Scenario: A configured directory that does not exist

- **WHEN** a configured directory is absent
- **THEN** it is skipped and reported, and the other directories are still
  scanned

#### Scenario: A configured directory holding no schemas

- **WHEN** a configured directory holds no schema folders
- **THEN** it is listed as holding none rather than being left out

#### Scenario: No directories are configured

- **WHEN** `schemas_dirs` is empty
- **THEN** the tab says how to configure one and offers the path prompt

### Requirement: A local schema is listed with what it declares

Each local schema SHALL be listed with its declared name, its artifact count and
the directory it was found in. A folder whose `schema.yaml` cannot be parsed
SHALL be listed as unreadable rather than hidden.

#### Scenario: A readable schema

- **WHEN** a local schema parses
- **THEN** it is listed with its declared name and artifact count

#### Scenario: An unreadable schema

- **WHEN** a folder holds a `schema.yaml` that does not parse
- **THEN** it is listed with its folder name and marked unreadable, so the user
  can open it and fix it

### Requirement: A local schema behaves like any other

A local schema SHALL open the same detail view, with the same artifact table,
figures, findings and diagram, as a registry or built-in schema.

#### Scenario: A local schema is opened

- **WHEN** `enter` is pressed on a local schema
- **THEN** its detail opens, read from its directory, with no fetch attempted

# install

## ADDED Requirements

### Requirement: A schema installs to the destination its name gives

Installing SHALL write the schema folder to `openspec/schemas/<name>/` inside
the project, where `<name>` is the schema's own declared name. The destination
SHALL NOT be taken from the registry entry, which carries no destination field.

#### Scenario: A schema is installed

- **WHEN** a schema named `minimalist` is installed
- **THEN** its folder lands at `openspec/schemas/minimalist/` in the project

#### Scenario: A name that differs from the source path

- **WHEN** a schema stored at `superspec/` declares `name: SuperSpec`
- **THEN** it installs to `openspec/schemas/SuperSpec/`, following the name and
  not the path

### Requirement: Nothing is written before the user agrees

Installing SHALL first show the destination, what would be written there, and
whether that directory already holds something. No file SHALL be written until
the user confirms.

#### Scenario: An install is started

- **WHEN** install is pressed
- **THEN** the destination and the files that would arrive are shown, and
  nothing has been written

#### Scenario: The user declines

- **WHEN** the user declines at the confirmation
- **THEN** nothing has been written and the project is exactly as it was

#### Scenario: The user agrees

- **WHEN** the user confirms
- **THEN** the files are written and the result reported

### Requirement: An occupied destination is refused unless overwriting is chosen

When the destination directory already exists, installing SHALL refuse by
default, say what is there, and require a separate explicit confirmation to
overwrite.

#### Scenario: The destination is occupied

- **WHEN** the destination already holds a schema
- **THEN** the install is refused, the collision is named, and overwriting is
  offered as its own choice

#### Scenario: Overwriting is confirmed

- **WHEN** the user explicitly confirms overwriting
- **THEN** the existing directory is replaced and the replacement is reported

#### Scenario: Overwriting is not confirmed

- **WHEN** the user does not confirm overwriting
- **THEN** the existing directory is untouched

### Requirement: OpenSpec validates what was installed

After writing, ossm SHALL run OpenSpec's own schema validation and report what
it says. A schema OpenSpec rejects SHALL be reported as installed and invalid.

#### Scenario: OpenSpec accepts the schema

- **WHEN** validation passes
- **THEN** the install is reported as done and the schema as valid

#### Scenario: OpenSpec rejects the schema

- **WHEN** validation fails
- **THEN** the install is reported as done and the problems OpenSpec reported
  are shown, rather than the install being reported as clean

#### Scenario: Validation cannot be run

- **WHEN** the `openspec` CLI cannot be run
- **THEN** the install is reported as done and unverified, saying why

### Requirement: Installing never changes the project default

Installing SHALL NOT write to `openspec/config.yaml`. Setting the project
default SHALL be a separate action with its own confirmation.

#### Scenario: A schema is installed

- **WHEN** an install finishes
- **THEN** `openspec/config.yaml` is byte for byte unchanged

#### Scenario: The default is set

- **WHEN** the user chooses to set the project default and confirms
- **THEN** `openspec/config.yaml` is updated and nothing else in it changes

### Requirement: Installing is refused outside a project

Installing SHALL be refused when the working directory is not inside an OpenSpec
project, with a message saying so.

#### Scenario: Install outside a project

- **WHEN** install is pressed outside any OpenSpec project
- **THEN** it is refused with a message saying installing needs a project, and
  browsing keeps working

#### Scenario: Install on a schema that is already available

- **WHEN** install is pressed on a schema OpenSpec already resolves, whether
  built in or already in the project
- **THEN** it is refused with a message saying the schema is already available

### Requirement: A failed install leaves the project as it was

When any step of an install fails, the project SHALL be left as it was before
the install started.

#### Scenario: The fetch fails

- **WHEN** the source cannot be fetched
- **THEN** nothing has been written into the project

#### Scenario: The copy fails partway

- **WHEN** writing fails partway through
- **THEN** the destination is left as it was, and no half-written schema is left
  behind

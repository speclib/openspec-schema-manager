# app-frame Specification

## Purpose
TBD - created by archiving change build-the-app-frame-and-e2e-harness. Update Purpose after archive.

## Requirements

### Requirement: Four tabs in a fixed order

The application SHALL present four tabs named Project, Registry, Local and
Composer, in that order. The order SHALL NOT depend on what the application
finds on the machine.

#### Scenario: The tab bar is drawn

- **WHEN** the application starts
- **THEN** the tab bar lists Project, Registry, Local and Composer in that order

#### Scenario: There is no OpenSpec project

- **WHEN** the application starts outside an OpenSpec project
- **THEN** all four tabs are still listed, and the Project tab is still reachable

### Requirement: The opening tab reflects where the user is standing

The application SHALL open on the Project tab when the working directory is
inside an OpenSpec project, and on the Registry tab when it is not.

#### Scenario: Started inside a project

- **WHEN** the application starts in a directory inside an OpenSpec project
- **THEN** the Project tab is selected

#### Scenario: Started outside a project

- **WHEN** the application starts in a directory that is not inside an OpenSpec
  project
- **THEN** the Registry tab is selected, because it is the one that works there

### Requirement: Tabs are reachable by cycling and by number

`tab` SHALL move to the next tab and `shift+tab` to the previous, both wrapping
around. The digits `1` to `4` SHALL select the tab at that position.

#### Scenario: Cycling forward past the last tab

- **WHEN** the last tab is selected and `tab` is pressed
- **THEN** the first tab is selected

#### Scenario: Cycling backward past the first tab

- **WHEN** the first tab is selected and `shift+tab` is pressed
- **THEN** the last tab is selected

#### Scenario: Selecting by number

- **WHEN** `3` is pressed
- **THEN** the third tab is selected, whatever was selected before

#### Scenario: A digit with no tab

- **WHEN** a digit above the number of tabs is pressed
- **THEN** the selection does not change and nothing is reported as an error

### Requirement: Help lists the keys that work where the user is standing

`?` SHALL open a help overlay listing the global keys and the keys the selected
tab adds, kept apart. `?`, `esc` and `q` SHALL close it.

#### Scenario: Help is opened

- **WHEN** `?` is pressed
- **THEN** an overlay appears listing the global keys and the selected tab's
  keys under separate headings

#### Scenario: Help follows the tab

- **WHEN** help is open and a different tab is selected
- **THEN** the overlay lists that tab's keys

#### Scenario: Help is closed

- **WHEN** help is open and `?`, `esc` or `q` is pressed
- **THEN** the overlay closes and the application is still running

### Requirement: Quitting never surprises a user dismissing something

`ctrl+c` SHALL quit from anywhere. `q` SHALL quit only when no overlay is open;
when one is open it SHALL close the overlay instead.

#### Scenario: q with nothing open

- **WHEN** `q` is pressed and no overlay is open
- **THEN** the application exits

#### Scenario: q with help open

- **WHEN** `q` is pressed and the help overlay is open
- **THEN** the overlay closes and the application keeps running

#### Scenario: ctrl+c with help open

- **WHEN** `ctrl+c` is pressed and the help overlay is open
- **THEN** the application exits

### Requirement: A status line carries the name, the version and a message

The application SHALL draw a status line carrying its name and version and a
message area that a screen can write into. A message SHALL be replaceable and
clearable by the screen that set it.

#### Scenario: The status line is drawn

- **WHEN** the application starts
- **THEN** the status line shows the application name and the running version

#### Scenario: A screen sets a message

- **WHEN** a screen sets a status message
- **THEN** it appears in the status line, and setting another replaces it

### Requirement: A run with no terminal reports it plainly

When the application is asked to start its interface with no terminal attached,
it SHALL write a plain message naming the problem and exit non-zero, rather than
failing with a terminal error.

#### Scenario: No terminal is attached

- **WHEN** ossm is started with its output redirected and no terminal attached
- **THEN** it writes a message saying it needs a terminal and exits non-zero,
  and writes no escape sequence

#### Scenario: A flag that needs no terminal

- **WHEN** `--version` or `--help` is given with no terminal attached
- **THEN** it answers as before and exits zero, because neither starts the
  interface

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

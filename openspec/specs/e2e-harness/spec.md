# e2e-harness Specification

## Purpose
TBD - created by archiving change build-the-app-frame-and-e2e-harness. Update Purpose after archive.

## Requirements

### Requirement: The harness drives the built binary

End to end tests SHALL exercise the compiled `ossm` binary as a user would,
rather than calling its packages. The binary SHALL be built once per test run
and shared by every case in it.

#### Scenario: A test run starts

- **WHEN** an end to end test run begins
- **THEN** the binary is built once, and every case in the run uses that build

#### Scenario: The build fails

- **WHEN** the binary cannot be built
- **THEN** the run fails with the compiler's output, rather than each case
  failing separately with the same message

### Requirement: The harness drives a terminal

The harness SHALL run the binary attached to a pseudo terminal, send key presses
to it, and read what it draws, so that a screen can be asserted on as drawn.

#### Scenario: A key is sent

- **WHEN** the harness sends a key press
- **THEN** the binary receives it as a terminal would deliver it

#### Scenario: Output is read

- **WHEN** the harness reads what the binary drew
- **THEN** it can wait for text to appear and fail with the screen contents when
  it does not

### Requirement: An end to end run cannot touch the machine's real state

The harness SHALL point `XDG_CONFIG_HOME`, `XDG_CACHE_HOME` and
`XDG_STATE_HOME` inside a temporary directory belonging to the test, and SHALL
run the binary with a working directory the test controls.

#### Scenario: A run writes state

- **WHEN** the binary under test writes a cache, a recents list or a draft
- **THEN** it lands inside the test's temporary directory and the user's real
  ossm directories are untouched

#### Scenario: A run reads configuration

- **WHEN** the binary under test reads its configuration
- **THEN** it reads the test's configuration, and a file in the user's home has
  no effect on the result

### Requirement: End to end tests are separable from the unit suite

The end to end tests SHALL be runnable on their own and SHALL be skippable, so
that the unit suite stays fast and an environment without a pseudo terminal can
still run the gate.

#### Scenario: A pseudo terminal is unavailable

- **WHEN** the environment cannot allocate a pseudo terminal
- **THEN** the end to end cases skip with a message saying why, and the rest of
  the suite passes

#### Scenario: Only the end to end tests are wanted

- **WHEN** a contributor runs the end to end package alone
- **THEN** it builds the binary and runs those cases without running the unit
  suite

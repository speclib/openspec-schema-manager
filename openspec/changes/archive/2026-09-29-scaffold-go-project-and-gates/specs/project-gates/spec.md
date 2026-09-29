# project-gates

## ADDED Requirements

### Requirement: One command decides whether the project passes

The project SHALL expose a single gate script that runs `go vet`, the whole test
suite and the coverage floor check. `nix flake check` and the CI workflow SHALL
both invoke that script rather than reimplementing its steps, so that a pass in
one place means the same thing as a pass in the other.

#### Scenario: The gate runs locally

- **WHEN** a contributor runs the gate script in the dev shell
- **THEN** it runs `go vet`, the full test suite and the coverage floor check,
  and exits non-zero if any of the three fails

#### Scenario: The gate runs under nix flake check

- **WHEN** `nix flake check` runs
- **THEN** it builds the package and invokes the same gate script, without
  restating the vet command, the test command or the coverage floors

#### Scenario: A floor is changed in one place

- **WHEN** a coverage floor is raised
- **THEN** it is edited in the gate script alone, and both `nix flake check` and
  CI enforce the new value without any other file changing

### Requirement: The test suite makes no network request

The test suite SHALL run to completion with no network access, because
`nix flake check` builds in a sandbox that has none. Any check that needs the
network SHALL live outside the suite the gate runs.

#### Scenario: The suite runs offline

- **WHEN** the test suite runs with no network available
- **THEN** every test either passes or fails on its own merits, and none fails
  or skips because a host could not be reached

#### Scenario: A check needs the network

- **WHEN** a check can only be answered by reaching a remote host
- **THEN** it is placed in a separate CI workflow rather than in the gate's
  suite, and its absence does not weaken the gate

### Requirement: The coverage floor ratchets and never drops silently

The gate SHALL fail when coverage falls below the recorded floor. The floor
SHALL be raised as tested code lands, and lowering it SHALL be a deliberate edit
to the gate script rather than a side effect of adding untested code.

#### Scenario: Coverage falls below the floor

- **WHEN** a change lowers coverage under the recorded floor
- **THEN** the gate fails and names the package and the measured percentage

#### Scenario: Coverage rises above the floor

- **WHEN** a change raises coverage well above the recorded floor
- **THEN** the gate passes, and the floor is a value a contributor can raise in
  the same change

### Requirement: The package boundaries are enforced by a test

The TUI package SHALL depend on the other packages and SHALL NOT be depended on
by them. A test SHALL assert this rather than leaving it to review, because an
import added in passing is invisible in a diff that is otherwise about
something else.

#### Scenario: A core package imports the TUI

- **WHEN** a package outside the TUI and the command entry point imports the TUI
  package
- **THEN** the boundary test fails and names the offending package and import

#### Scenario: The TUI imports a core package

- **WHEN** the TUI package imports the graph, schema, registry or config package
- **THEN** the boundary test passes, because that is the permitted direction

### Requirement: The binary answers without a terminal

`ossm --version` and `ossm --help` SHALL write to standard output and exit zero
without starting the TUI, so that the binary can be exercised from a script and
from an end to end test that has no terminal attached.

#### Scenario: Version is asked for

- **WHEN** `ossm --version` runs with no terminal attached
- **THEN** it prints the version and exits zero, without clearing the screen or
  entering raw mode

#### Scenario: Help is asked for

- **WHEN** `ossm --help` runs with no terminal attached
- **THEN** it prints the available flags and exits zero

### Requirement: The flake enumerates supported systems explicitly

The flake SHALL build the supported systems from an explicit list using
`nixpkgs.lib.genAttrs`, and SHALL NOT take flake-utils as an input.

#### Scenario: A system is added

- **WHEN** a system is added to the supported list
- **THEN** the packages, checks and dev shell outputs all gain that system
  without any other edit

#### Scenario: The flake inputs are read

- **WHEN** the flake's inputs are inspected
- **THEN** nixpkgs is the only one, and flake-utils is absent

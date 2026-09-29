# artifact-graph

## ADDED Requirements

### Requirement: The graph is built from artifacts and their requires edges

A graph SHALL be built with one node per artifact and one edge per entry in an
artifact's `requires`, directed from the requirement to the artifact that needs
it, so that an edge points the way work flows.

#### Scenario: A straight chain

- **WHEN** a schema declares specs, then tasks requiring specs
- **THEN** the graph holds two nodes and one edge from specs to tasks

#### Scenario: A branch

- **WHEN** two artifacts both require the same one
- **THEN** the graph holds two edges out of that one

#### Scenario: A join

- **WHEN** one artifact requires two others
- **THEN** the graph holds two edges into it

### Requirement: The order is stable

The graph SHALL expose a topological order that is the same every time for the
same schema, so that two runs never draw the same schema differently. Where the
topological order leaves a choice, it SHALL be settled by the order the
artifacts were declared in.

#### Scenario: The same schema is ordered twice

- **WHEN** the same schema is ordered twice
- **THEN** the order is identical

#### Scenario: Two artifacts could come next

- **WHEN** two artifacts are both ready at the same point in the order
- **THEN** the one declared first in the schema comes first

#### Scenario: A cyclic schema

- **WHEN** the graph holds a cycle
- **THEN** ordering reports that rather than returning a partial order

### Requirement: The figures that make two schemas comparable

The graph SHALL report the artifact count, the length of the longest dependency
chain, how many artifacts gate apply, and how many artifacts nothing requires.

#### Scenario: A two-artifact chain

- **WHEN** a schema declares specs and tasks with tasks requiring specs, and
  apply gated on tasks
- **THEN** the artifact count is 2, the longest chain is 2, the gate count is 1

#### Scenario: A branchy schema

- **WHEN** a schema branches and rejoins
- **THEN** the longest chain counts the longest path through it, not the number
  of artifacts

#### Scenario: A single artifact

- **WHEN** a schema declares one artifact
- **THEN** the longest chain is 1, not 0

#### Scenario: Artifacts nothing requires

- **WHEN** three artifacts are required by nothing
- **THEN** the count of them is 3, which is what says how wide a workflow's
  leaves are

### Requirement: The graph reports what gates apply

The graph SHALL report which nodes are apply gates, so that a renderer can mark
them without reading the schema again.

#### Scenario: A node gates apply

- **WHEN** an artifact is named in `apply.requires`
- **THEN** the graph reports that node as a gate

#### Scenario: A node does not gate apply

- **WHEN** an artifact is not named in `apply.requires`
- **THEN** the graph does not report it as a gate

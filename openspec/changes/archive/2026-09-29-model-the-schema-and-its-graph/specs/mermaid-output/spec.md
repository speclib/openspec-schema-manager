# mermaid-output

## ADDED Requirements

### Requirement: The graph emits a Mermaid flowchart

The graph SHALL emit a Mermaid `graph TD` flowchart with one node per artifact
and one edge per `requires` entry, directed the way work flows.

#### Scenario: A chain is emitted

- **WHEN** a schema of specs then tasks is emitted
- **THEN** the output opens with `graph TD`, declares both nodes and draws one
  edge from specs to tasks

#### Scenario: An artifact with no requirements

- **WHEN** an artifact requires nothing
- **THEN** it appears as a node with no incoming edge, rather than being left
  out

### Requirement: An apply gate is marked distinctly

A node that gates apply SHALL be drawn with a different node shape from one that
does not, so a reader can tell them apart without a legend.

#### Scenario: A gate node

- **WHEN** an artifact gates apply
- **THEN** its node uses the gate shape

#### Scenario: A plain node

- **WHEN** an artifact does not gate apply
- **THEN** its node uses the plain shape

### Requirement: An artifact id is safe to put in a diagram

An artifact id that is not a valid Mermaid identifier SHALL be given a generated
node identifier, and its own text SHALL be used as the node's label. A label
SHALL be quoted and escaped so that no id can break the diagram.

#### Scenario: An ordinary id

- **WHEN** an id is alphanumeric with hyphens
- **THEN** it is used as the node identifier and as the label

#### Scenario: An id with characters Mermaid cannot take

- **WHEN** an id holds a space, a quotation mark, a bracket or a semicolon
- **THEN** the node gets a generated identifier, the label carries the original
  text escaped, and the output is still a parseable flowchart

#### Scenario: Two ids differing only in what is stripped

- **WHEN** two ids reduce to the same generated identifier
- **THEN** they still get distinct node identifiers

### Requirement: The same schema emits the same diagram

Emission SHALL be deterministic: the same schema SHALL produce byte-identical
output every time, in the graph's stable order.

#### Scenario: The same schema is emitted twice

- **WHEN** a schema is emitted twice
- **THEN** the two outputs are byte-identical

#### Scenario: Edges are ordered

- **WHEN** a schema has several edges
- **THEN** they are written in the graph's stable order rather than in map order

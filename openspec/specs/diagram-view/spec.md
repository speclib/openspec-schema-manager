# diagram-view Specification

## Purpose
TBD - created by archiving change show-a-schema-in-detail. Update Purpose after archive.

## Requirements

### Requirement: The diagram is drawn from the artifact graph

`d` SHALL toggle a flow diagram drawn from the schema's artifact graph, with one
box per artifact, arrows along the `requires` edges, and apply gates drawn in a
different shape from the rest.

#### Scenario: The diagram is opened

- **WHEN** `d` is pressed on a schema with artifacts
- **THEN** a diagram appears with a box per artifact and an arrow per
  requirement

#### Scenario: A gate is drawn distinctly

- **WHEN** a schema has an apply gate
- **THEN** that artifact's box is drawn in a different shape from the others

#### Scenario: The diagram is closed

- **WHEN** `d` is pressed again
- **THEN** the diagram closes and the detail returns

### Requirement: A schema that cannot be drawn says why

A schema whose graph holds a cycle cannot be laid out. The view SHALL say so
rather than showing an empty pane or a partial diagram.

#### Scenario: A cyclic schema

- **WHEN** the diagram is opened on a schema with a cycle
- **THEN** it says the schema cannot be drawn because its artifacts require each
  other, and names them

#### Scenario: A schema with no artifacts

- **WHEN** the diagram is opened on a schema declaring no artifacts
- **THEN** it says there is nothing to draw

### Requirement: The diagram scrolls in both directions

A diagram larger than its pane SHALL scroll horizontally and vertically from the
keyboard, and SHALL be rendered once and then scrolled rather than re-rendered
on every key.

#### Scenario: A diagram taller than the pane

- **WHEN** the diagram is taller than the pane and the down key is pressed
- **THEN** the view moves down and the top of the diagram scrolls out of sight

#### Scenario: A diagram wider than the pane

- **WHEN** the diagram is wider than the pane and the right key is pressed
- **THEN** the view moves right, and no line is wrapped to fake the width

#### Scenario: Scrolling stops at the edges

- **WHEN** scrolling continues past the top, bottom, left or right edge
- **THEN** the view stops at the edge rather than scrolling into empty space

#### Scenario: A diagram smaller than the pane

- **WHEN** the diagram fits its pane
- **THEN** scrolling changes nothing and is not reported as an error

### Requirement: The renderer falls back to ASCII

Drawing SHALL use Unicode box-drawing characters, and SHALL fall back to ASCII
when the terminal cannot be trusted with them.

#### Scenario: A terminal that can draw Unicode

- **WHEN** the terminal reports a UTF-8 locale
- **THEN** the diagram uses box-drawing characters

#### Scenario: A terminal that cannot

- **WHEN** the terminal reports no UTF-8 locale
- **THEN** the diagram uses ASCII and is still readable

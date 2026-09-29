# Design

## Two packages, one direction

`internal/schema` parses and validates. `internal/graph` answers graph
questions and emits Mermaid. `graph` takes what `schema` produces and `schema`
knows nothing about `graph`.

The split is not tidiness. The composer in milestone 07 builds a graph that no
`schema.yaml` describes yet, and the web front wants the Mermaid without the
YAML. Both need `graph` without `schema`.

## The model

```go
type Schema struct {
    Name        string
    Version     int
    Description string
    Artifacts   []Artifact
    Apply       Apply
}

type Artifact struct {
    ID          string
    Generates   string
    Template    string
    Description string
    Instruction string
    Requires    []string
}

type Apply struct {
    Requires    []string
    Tracks      string
    Instruction string
}
```

`Requires` absent and `requires: []` both give a nil slice, so nothing
downstream has to tell them apart. `Generates` keeps its glob as written: it
describes a set of files, and expanding it needs a directory that may not exist.

Decoding is strict, with `KnownFields(true)`, unlike the registry. A schema is
either written by the user in front of you or fetched from a repository they
chose to install. A key ossm silently ignores in a schema is a workflow step
that quietly does not happen, which is worse than a failed parse.

`Version` is an int because every schema in the wild declares `version: 1`. If
one ever declares something else, the parse error names the field, which is a
better way to find out than a silent zero.

## Reporting every problem

`Validate` returns a slice of findings rather than an error:

```go
type Finding struct {
    Severity Severity   // Fatal or Warning
    Artifact string     // empty when it is about the schema
    Message  string
}
```

Returning the first error means someone fixing a schema runs validation once per
mistake. The composer needs the whole list anyway, because it validates
continuously and shows the findings in a pane.

Fatal and Warning are separated because the two have different consequences.
A cycle means the workflow cannot run. Two artifacts writing the same path means
one clobbers the other, which is usually a mistake and occasionally deliberate.
Refusing to open a schema over the second would make ossm less useful than a
text editor.

## The template check is conditional

`Validate(s)` checks everything that can be answered from the model.
`ValidateDir(s, dir)` also checks that every template exists under
`templates/`. A registry entry has no files until milestone 04 fetches them, and
a validator that reports every template as missing teaches people to ignore it.

## Cycles

Cycle detection is a depth-first walk with a colour per node, which gives the
cycle's members in the order they connect rather than just the fact of one.
"tasks requires design requires tasks" is a fixable message; "this schema has a
cycle" is not.

A self-reference is reported as its own kind rather than as a one-node cycle,
because that is what the reader is looking at.

## Stable order

Kahn's algorithm with the ready set kept in declaration order. A plain map-based
topological sort gives a different order per run, and the diagram would then
move between runs for no reason a user could see. Declaration order is the
tiebreaker because it is the order the schema's author chose.

## Longest chain

The longest path in a DAG, computed over the topological order in one pass. It
counts nodes, not edges, so a single artifact is a chain of 1. A chain of 0
would mean an empty schema, which validation already rejects.

## Mermaid identifiers

An artifact id is a string from a third party. Mermaid node identifiers cannot
hold spaces, quotation marks or brackets, and a label can break the diagram if
it is not quoted.

So: if an id matches `^[A-Za-z][A-Za-z0-9_-]*$` it is used directly. Otherwise
the node gets `n<index>`, and the original text becomes the quoted label with
quotation marks escaped. Two ids that would collide get distinct identifiers
because the index is the artifact's position, which is unique by construction.

Gate nodes use `id{{"label"}}`, the hexagon, and plain nodes use `id["label"]`.
Two shapes rather than a class definition, because the ASCII renderer in
milestone 04 reads shapes and ignores styling.

## What is not here

Nothing renders. `graph.Mermaid` returns a string. Turning it into boxes is
milestone 04's job and needs a renderer this package must not depend on, or the
web front cannot reuse it.

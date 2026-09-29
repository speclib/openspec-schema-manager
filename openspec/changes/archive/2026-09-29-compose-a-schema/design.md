# Design

## The model

```go
type Artifact struct {
    ID          string
    Generates   string
    Description string
    Template    string   // path inside the source's templates/
    Instruction string
    Requires    []string

    SourceDir  string    // where to copy the template from
    SourceName string    // the schema it came from
    SourceRef  string
    SourceID   string    // the id it had there
}

type Composition struct {
    Artifacts []Artifact
    Gates     []string
    Tracks    string
}
```

The artifact carries its provenance rather than pointing at a source object.
A draft is then a plain JSON round trip, and resuming needs nothing but the
file. The cost is four repeated strings per artifact, which is nothing against
not having to reconstruct a graph of pointers from a file.

`Requires` is filtered on add to what is already on the canvas. An artifact that
required `proposal` in its own schema does not silently pull `proposal` in, and
does not sit there unresolved either. The user chooses what to link, which is
the whole point of a composer; the alternative is a tool that keeps adding
things you did not ask for.

An artifact added later does not restore an edge that was dropped, for the same
reason. Guessing an edge back is worse than a missing one, because a missing
edge is visible on the canvas.

## Collisions

`Add` returns a collision rather than renaming for you. Two schemas both
declaring `tasks` is common, and which one keeps the name is a decision about
the workflow being built. Renaming to `tasks-2` automatically is the kind of
help that has to be undone.

## Validation

`internal/compose` does not reimplement `internal/schema`'s validation. It
builds the `schema.Schema` the composition describes and runs `schema.Validate`
on it, then adds what only a composition can be wrong about: nothing selected at
all.

That keeps one definition of a valid schema. A composer that agreed with itself
and disagreed with the validator would let a user build something the detail view
then calls broken.

## Template reference scanning

For each artifact on the canvas, its template text and instruction are searched
for:

- the id of any artifact on any source schema that is not on the canvas
- the `generates` path of any such artifact

Matching is on word boundaries, so `design` does not match `designer` and
`tasks.md` does not match `subtasks.md`. Only ids and paths from the sources the
user actually added are looked for; scanning for arbitrary words would produce
noise no one reads.

The warnings are advisory. A template mentioning an artifact the user
deliberately left out is their decision, and a composer that refuses to write
over it is one people work around by editing the template to lie.

## Writing

`schema.yaml` is generated, not stitched. A composed schema has no single source
document to preserve comments from, and generating it is the only way the
artifact order and the requirements can be exactly what the canvas holds. The
provenance comment block is therefore the only comment in the file, which makes
it easy to find.

Templates are copied byte for byte. Two artifacts from different schemas can
declare the same template path (`tasks.md` is not rare), so on collision the
second is written under a path prefixed with its source schema's name and the
artifact's `template` names that. A schema whose two artifacts share a template
would otherwise silently lose one.

Writing stages into a temporary directory beside the destination and renames, the
way install and duplicate already do.

## Drafts

One JSON file per draft under `$XDG_STATE_HOME/ossm/drafts/`, named after the
draft. Listing sorts by modification time, most recent first. A file that does
not parse is reported and skipped, because one bad draft must not hide the rest.

Resuming checks each artifact's `SourceDir` still exists and reports the ones
that have gone. The composition is still editable: the user can remove those
artifacts and write the rest, which is better than refusing to open it.

## The screen

Two panes side by side, palette left and canvas right, the layout the briefing
sketches. Below them, the validation findings and the template warnings.

Key handling is modal, because linking, renaming, naming a tracked file, naming
a draft and naming the written schema all need text. Each is a small state with
its own prompt, in the same shape the install flow already uses, and
`Capturing()` reports true in every one of them so the frame keeps its hands off.

The diagram is the same `graph.Draw` the detail view uses, on a graph built from
the canvas. A composition with a cycle cannot be drawn, and says so, which is
the same message the detail view gives for the same reason.

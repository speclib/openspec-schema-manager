# Design

## Why ossm duplicates rather than calling openspec schema fork

`openspec schema fork` was tested before this was decided. Two things rule it
out.

It takes a schema name OpenSpec can already resolve, not a path. A registry
schema that has not been installed and an arbitrary folder opened by path are
both invisible to it, and those are the two cases that matter here.

And it fails outright when the source is read-only:

```
EACCES: permission denied, open '.../openspec/schemas/.fork-staging-WFwIPY/schema.yaml'
```

It copies the source preserving its mode, then reopens the copy for writing to
change the name. A schema in the Nix store is mode 444, so forking any built-in
schema fails on every Nix install. Forking a schema already in the project works,
because that copy is writable.

So ossm copies the tree itself and writes each file mode 644. `NOTES.md` records
the finding.

## Rewriting the name

The copied `schema.yaml` is edited as text, replacing only the `name:` line,
rather than being parsed and re-serialised. A schema is full of instruction
blocks, comments and a deliberate artifact order, and round-tripping through a
YAML marshaller loses comments and can reorder keys. The one line that has to
change is the one line that changes.

## Scanning

A configured directory is scanned one level deep: a folder directly inside it
holding a `schema.yaml` is a schema. Not recursive, because `schemas_dirs`
points at a directory of schemas, and walking a whole home directory looking for
`schema.yaml` is how a tool earns a reputation for being slow.

A folder whose `schema.yaml` does not parse is listed and marked unreadable. It
is exactly the folder the user wants to open and fix, and hiding it makes ossm
useless at the moment it would be most useful.

## The recents file

```json
{"paths": ["/home/t/work/schemas/team-review", "/tmp/experiments/quick"]}
```

Most recent first, capped at `recents_cap`, no duplicates. Written atomically the
way the registry cache is: temporary file in the same directory, then rename.

A remembered path that no longer holds a schema is shown as missing rather than
dropped. Dropping it silently means the user cannot tell whether they
misremembered the path or the folder moved.

## The path prompt lives in the frame

`:` and `ctrl+o` open it from any tab, so the frame owns it, not the Local
screen. It is the first thing in ossm that captures text at the frame level, and
it uses the same rule the screens do: while it is open the frame takes `ctrl+c`
and passes everything else to the prompt.

Completion is `filepath.Glob(typed + "*")` filtered to directories, filling in
the longest common prefix. Not a full readline. A path prompt used a few times a
session does not need history, kill rings or word motion, and every one of those
is a key that has to stop meaning what it means everywhere else.

## Editing

Bubble Tea's `tea.ExecProcess` releases the terminal, runs the editor and returns
a message when it exits. The editor comes from `$VISUAL`, `$EDITOR` and then
`vi`. The value is split on spaces so `EDITOR="code --wait"` works, which is
common enough that ignoring it would be a bug report.

No shell is involved. Passing the editor through `sh -c` would let a crafted
`$EDITOR` do more than edit, and it buys only quoting that a split on spaces
already covers for every editor anyone actually sets.

On return the schema is re-read from disk and re-validated with `ValidateDir`,
so a template that was reported missing stops being reported once it exists.

## The Local screen's shape

Three sections in one list: schemas from the configured directories, grouped by
directory; the recents; and whatever was opened by `--path`. One selection moves
through all of them, because two independent cursors in one pane is the kind of
thing that reads fine in a design and is miserable to use.

`enter` opens the detail, `c` duplicates, `t` shows the file tree, `e` edits the
selected file. The file tree is a mode of the Local screen rather than a fourth
screen, because it is always about the schema the cursor is on.

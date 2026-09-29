# Work with local schemas

Beans epics: `openspec-schema-manager-pkws` (Local schema directories),
`openspec-schema-manager-2on6` (Path prompt and recents),
`openspec-schema-manager-t4t5` (Duplicate a schema),
`openspec-schema-manager-myh3` (File tree and editor round trip).

## Why

Finding a schema that nearly fits is the common case. Nobody's way of working
matches a stranger's exactly, and the gap between "this is close" and "this is
mine" is a copy and a few edits. Without that, ossm can only show a user
somebody else's workflows.

A schema is plain YAML and Markdown, so the editing is not the hard part. What
is missing is somewhere for a schema that is not in a registry and not in a
project to live, a way to reach it, and a way to copy something into it.

## What Changes

- The Local tab lists schemas found in the directories `schemas_dirs` names: any
  folder holding a `schema.yaml`.
- `:` and `ctrl+o` open a path prompt that takes any directory, with completion,
  so a schema in an awkward place costs one typing rather than a configuration
  change.
- A path that is opened is remembered. The recents list lives in the state
  directory, is capped by `recents_cap`, and is shown in its own section.
- `--path <dir>` opens a folder on launch, and that also counts as a recent.
- `c` duplicates the selected schema into a chosen local directory under a new
  name, rewriting `name` in the copied `schema.yaml`.
- Duplication is done by ossm rather than by `openspec schema fork`. Fork takes
  a schema OpenSpec already resolves, not a path, so it cannot copy a registry
  schema or an arbitrary folder; and it fails outright when the source is
  read-only, which is every schema on a Nix install.
- A local schema's files are browsable as a tree: `schema.yaml` and everything
  under `templates/`.
- `e` opens the selected file in `$VISUAL`, then `$EDITOR`, then `vi`,
  suspending the TUI and resuming when the editor exits.
- On return the schema is re-read and re-validated, and the findings are shown
  against the file tree.

## Capabilities

### New Capabilities

- `local-schemas`: where local schemas are found and what a local schema is.
- `path-prompt`: opening a folder by path, and what is remembered.
- `duplicate`: copying a schema under a new name, and what is rewritten.
- `file-tree`: browsing a schema's files, editing one, and what happens on
  return.

### Modified Capabilities

- `app-frame`: the Local tab stops being a placeholder, and the path prompt is
  reachable from anywhere.

## Impact

- New: the Local screen, the path prompt, the recents store and the file tree in
  `internal/tui`; duplication and the recents file in packages below it.
- ossm writes outside a project for the first time: into a local schemas
  directory, and into the recents file. Duplication shows its destination and
  refuses to overwrite.
- Suspending for an editor is Bubble Tea's `ExecProcess`. Nothing is built that
  edits text in the terminal, following D11.
- The recents file is ossm's own state, which is not the same as tracking
  install state. It records paths a user opened, not what a project has.

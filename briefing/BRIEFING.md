# Briefing: ossm — OpenSpec Schema Manager

## 1. Why this exists

OpenSpec supports custom schemas: a schema defines the artifacts a change
produces, their templates and the order they depend on each other. A schema is
effectively *a way of working* with OpenSpec and an AI agent. A light schema
suits a solo side project; a heavier one with review gates suits a team.

Most OpenSpec users never find this out, and those who do have no easy way to
compare schemas or try one. **ossm lowers the barrier for OpenSpec users to
discover, try and compare schemas, so they can find the workflow that fits a
project.** Trying a schema in a demo project, or carefully in a real one, should
take seconds. Git makes experimenting safe.

ossm is the power-user tool and the main reason the registry exists. A public
web store front on the same registry is planned later and is out of scope here.

## 2. Context

- **Registry**: `github.com/speclib/openspec-schema-registry`. A single JSON
  index (`openspec-schemas.json`) of schemas that live in their own GitHub repos.
  Community-extendable by PR. The built-in `spec-driven` schema is deliberately
  not listed, because it is present without installing; a tool that shows every
  schema available must merge the built-in list with the registry.
  Field reference: `docs/registry-entry.md`. Tickets: hidden `.beans/` directory.
- **Sibling tool**: `github.com/speclib/specgetty` (`spg`), a Bubble Tea TUI for
  local OpenSpec projects. Match its conventions.
- **Organisation**: SpecLib (speclib.eu), promoting spec-driven development and
  OpenSpec tooling.

## 3. OpenSpec schema anatomy (verify against current docs)

A schema is a folder:

```
<schema-name>/
├── schema.yaml
└── templates/
    ├── proposal.md
    └── ...
```

`schema.yaml` declares artifacts. Each artifact has an `id`, what it
`generates` (a path or glob inside the change folder), a `template`, an optional
`description` and `instruction`, and `requires`: the artifact ids that must
exist first. An `apply` block names the artifacts that gate implementation
(`apply.requires`) and the file that tracks progress (`apply.tracks`).

```yaml
name: minimalist
version: 1
description: Lightweight schema for well-scoped, low-risk changes
artifacts:
  - id: specs
    generates: specs/**/*.md
    template: specs/spec.md
    requires: []
  - id: tasks
    generates: tasks.md
    template: tasks.md
    requires: [specs]
apply:
  requires: [tasks]
  tracks: tasks.md
```

Schema resolution order in OpenSpec (first match wins): project
(`openspec/schemas/<name>/`), user (`~/.local/share/openspec/schemas/<name>/`),
then built-in package schemas. The project's `openspec/config.yaml` sets the
default schema; each change can use a different one, so a project can have
several schemas active at once.

**Consequence for ossm:** the artifact list plus `requires` edges is a directed
acyclic graph. That graph is the core data model for the detail view, the
diagram, and the composer.

## 4. Scope of this POC

### In scope

1. **Registry browse** — fetch, cache and search the registry.
2. **Project view** — the OpenSpec project you are standing in: its default
   schema, installed schemas, and which schema each change uses.
3. **Schema detail** — metadata, artifact list, apply gate, and a flow diagram.
4. **Install** — install a registry schema into the current project by handing
   off to OpenSpec.
5. **Local schemas** — schemas from a configured directory, plus any folder
   opened by path, with a recents list.
6. **Authoring** — duplicate a schema, browse its files, press `e` to edit in
   `$EDITOR`.
7. **Composer** — build a new schema from artifacts taken from several schemas,
   by editing the dependency graph with the keyboard.

### Out of scope

- The web store front, download counting, telemetry of any kind.
- Runtime composition (several schemas cooperating on one live change).
  Investigated: OpenSpec binds one schema per change; nothing supports this.
- ossm's own lockfile or install tracking. That is OpenSpec's job.
- A search index service. The registry is one JSON file.
- Sharing things other than schemas (specs, changes). See §9.

## 5. Features in detail

### 5.1 Registry browse

- On start, load the cached registry. Fetch fresh in the background if the cache
  is older than a configurable TTL (default 24h); `r` forces a refresh.
- Work fully offline from cache; show cache age in the status bar.
- Fuzzy filter over id, name and description (`/`).
- Merge in the built-in schemas reported by OpenSpec, marked as built-in.
- The registry entry is thin by design. Richer detail (the schema's own
  `schema.yaml`, templates, README) is fetched lazily from the schema's source
  repo when the user opens it, and cached under `$XDG_CACHE_HOME/ossm/schemas/`
  keyed by source and ref.
- Entries may pin a ref (tag or commit) or track a branch such as main. That is
  the schema author's call. Show which it is.

### 5.2 Project view

- Detect the OpenSpec project from the working directory (walk up to the
  `openspec/` folder), as specgetty does. No project: the view says so and the
  rest of ossm still works.
- Show the project default schema, every schema available to the project and
  where it resolves from (project / user / built-in), and a list of changes with
  the schema each uses.
- Source all of this from OpenSpec (CLI output, `config.yaml`, change folders).
  Do not keep a parallel record.
- Where possible, flag an installed schema whose registry entry points at a
  newer ref and offer to update it. How reliably this can work depends on what
  OpenSpec records at install time; see open questions.

### 5.3 Schema detail

- Header: name, version, description, source repo and ref, local path if any.
- Artifact table: id, generates, template, requires.
- Apply gate: required artifacts and tracked file.
- Summary figures useful for comparing: artifact count, longest dependency
  chain, number of gates before apply.
- Flow diagram (see 5.4), toggled with `d` or shown in a split pane.
- `i` install, `c` duplicate to local, `o` open source repo URL.

### 5.4 Flow diagram

- Convert the artifact graph to a Mermaid flowchart (`graph TD`; node per
  artifact, edge per `requires` entry, mark apply-gate nodes distinctly) and
  render it with mermaid-ascii in Unicode mode, ASCII fallback.
- Schemas may be branchy; rely on the renderer's layout rather than
  hand-placing boxes.
- The diagram must scroll horizontally and vertically when larger than the
  pane. Render into a string once, then viewport it.
- Keep Mermaid generation in its own package so the web front can reuse it.

### 5.5 Install

- Refuse outside an OpenSpec project, with a clear message.
- Hand off to OpenSpec. If the CLI has a command that installs a schema from a
  source, use it. If not, the fallback is the method OpenSpec's own docs
  describe: place the schema folder in the project's `openspec/schemas/<name>/`,
  then run OpenSpec's validation. Confirm which applies before building.
- Before writing anything: show what will be written where, and whether it
  would overwrite an existing schema of the same name. Require confirmation.
- Offer to set the new schema as the project default as a separate, explicit
  step. Never change `config.yaml` silently.
- After install, refresh the project view from OpenSpec.

### 5.6 Local schemas

- Config key `schemas_dirs`: directories scanned for schema folders (a folder
  containing `schema.yaml`).
- `:` or `ctrl+o` opens a path prompt (with completion) to open any schema
  folder, however awkward its location. Opened paths go into a recents list
  in `$XDG_STATE_HOME/ossm/recents.json`, capped (default 20).
- A `--path <dir>` CLI flag does the same on launch.
- Local schemas appear in their own section and behave like registry schemas
  in the detail view, diagram and composer.

### 5.7 Authoring

- `c` on any schema duplicates it into a chosen local schemas directory under
  a new name, rewriting `name` in `schema.yaml`. Prefer OpenSpec's own fork
  command if it exists for the case.
- A file tree for a local schema: `schema.yaml` and `templates/`.
- `e` opens the selected file in `$VISUAL`, else `$EDITOR`, else `vi`, using
  Bubble Tea's exec support to suspend and resume the TUI.
- On return, re-parse and re-validate; show errors inline (unknown artifact in
  `requires`, cycles, missing template file, unknown `apply.requires`).

### 5.8 Composer (the interesting part)

Build-time composition: assemble a brand-new, standalone schema from artifacts
drawn from several existing schemas.

- Start from an empty canvas or an existing schema.
- **Palette**: artifacts from any number of chosen source schemas, grouped by
  schema, each showing its id and what it generates.
- **Canvas**: the artifacts picked so far and their edges, rendered as the same
  diagram as 5.4, plus a list view that is the actual editing surface.
- Keyboard flow:
  - `a` add the palette selection to the canvas
  - `x` remove selected artifact (and its edges)
  - `l` start a link from the selected artifact, pick the target from a list,
    `enter` to confirm: adds a `requires` edge
  - `u` remove an edge
  - `g` toggle the selected artifact as an apply gate
  - `t` set the tracked file
  - `R` rename an artifact id (updates every edge that references it)
- Id collisions between sources are resolved on add by prompting for a new id.
- Continuous validation: no cycles, every `requires` resolves, at least one
  apply gate, tracked file set, no two artifacts generating the same path.
- **Template caveat**: templates and instructions written for one schema often
  mention sibling artifacts that may not exist in the new mix. After each add,
  scan the template and instruction text for references to artifact ids or
  generated file names that are not on the canvas, and list them as warnings
  with a jump to edit (`e`).
- `w` writes the result: a new schema folder with a generated `schema.yaml` and
  each template copied from its source, into a local schemas directory. Add a
  comment block at the top of `schema.yaml` recording which source schema and
  ref each artifact came from.
- Composition state can be saved as a draft (`$XDG_STATE_HOME/ossm/drafts/`)
  and resumed.
- Mouse is not required anywhere. Optional mouse selection is a later nicety.

## 6. Architecture

```
cmd/ossm/            main, flags, wiring
internal/config      XDG paths, config.yml, recents
internal/registry    fetch, cache, parse, search, merge built-ins
internal/source      fetch a schema folder from a source+ref (git or HTTP), cache
internal/schema      schema.yaml model, parse, validate, write
internal/graph       artifact DAG: build, cycle check, metrics, Mermaid output
internal/openspec    adapter over the openspec CLI (project detect, list, install)
internal/compose     composer model: canvas, edges, id remap, template scan, write
internal/tui         Bubble Tea models per screen; depends on everything above
```

- The `openspec` adapter is an interface with a fake for tests.
- Fetching a schema: prefer shallow git clone of the source at the ref into the
  cache; fall back to fetching raw files if a source is not a git repo. Respect
  the registry's own description of where the schema lives inside the repo.
- Everything under `internal/graph` and `internal/schema` is pure and fully unit
  tested.

## 7. Milestones

Each milestone must build, pass tests, and be committed.

1. **Skeleton** — repo, Nix flake, CI (`go test`, `go vet`, `nix flake check`),
   config and XDG paths, empty Bubble Tea app with tab bar and help overlay.
2. **Registry** — fetch, cache, TTL, offline mode, search, built-in merge.
3. **Schema model and graph** — parse, validate, DAG metrics, Mermaid output.
4. **Detail view and diagram** — detail screen, lazy fetch, mermaid-ascii
   rendering in a scrollable viewport.
5. **Project view and install** — project detection, OpenSpec adapter,
   installed and per-change schemas, guarded install flow.
6. **Local schemas and authoring** — schemas dirs, path prompt, recents,
   duplicate, file tree, `$EDITOR` round-trip, re-validation.
7. **Composer** — palette, canvas, keyboard edge editing, validation, template
   reference warnings, write, drafts.
8. **Polish** — update detection if feasible, README with screenshots
   (VHS tape files), help text for every screen.

## 8. Acceptance criteria

- `ossm` launched outside any project shows the registry and local schemas;
  pressing install explains that you must be inside an OpenSpec project.
- `ossm` launched inside an OpenSpec project shows its default schema, available
  schemas with their origin, and each change with its schema.
- Opening a registry schema shows its artifacts and a readable diagram, working
  offline once cached.
- Installing a registry schema into a demo project results in OpenSpec seeing
  it (`openspec schema which <name>` or equivalent) and a change can be created
  with it.
- Duplicating a schema, editing a template via `e`, and returning shows the
  updated content with validation results.
- The composer can build a schema from artifacts of two fixture schemas, catch
  a cycle, warn about a dangling template reference, and write a schema that
  OpenSpec accepts and can create a change with.
- All unit tests run offline.

## 9. Later, not now

- **Web store front** on speclib.eu, generated statically from the registry,
  reusing `internal/graph` for diagrams. Real download counts preferred over
  stars, via a small opt-in, anonymous counter (serverless).
- **Generic registry**: a `type` field on entries so the same machinery could
  later index other shareable OpenSpec pieces. Don't build it; avoid designs
  that make it impossible.
- Runtime composition, if OpenSpec ever supports it.

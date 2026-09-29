# Screens and key bindings

Global: `tab`/`shift+tab` switch tabs, `/` filter, `?` help, `r` refresh
registry, `:` or `ctrl+o` open path, `q` back/quit.

## Tabs

```
 Project │ Registry │ Local │ Composer
```

## Project tab (inside an OpenSpec project)

```
 ~/src/demo-app  ·  default schema: spec-driven
 ─────────────────────────────────────────────────────────
 Schemas available              origin
 ▸ spec-driven                  built-in   (default)
   minimalist                   project
   research-first               user

 Changes                        schema
   add-auth                     spec-driven
   fix-export                   minimalist
 ─────────────────────────────────────────────────────────
 enter detail · s set default · u update · ? help
```

Outside a project: "Not in an OpenSpec project — browse Registry or Local.
Installing needs a project."

## Registry tab

```
 Registry · 14 schemas · cached 3h ago
 ─────────────────────────────────────────────────────────
 ▸ minimalist        2 artifacts   specs → tasks          main
   event-driven      6 artifacts   discovery → … → tasks  v1.2
   research-first    3 artifacts   research → … → tasks   a1b2c3d
   spec-driven       4 artifacts   built-in
 ─────────────────────────────────────────────────────────
 enter detail · i install · c duplicate · / filter
```

## Schema detail

```
 research-first  v1  ·  github.com/example/schemas @ main
 Research before proposing
 ─────────────────────────────────────────────────────────
 Artifacts (3) · chain length 3 · apply gate: tasks
   id         generates      requires
   research   research.md    —
   proposal   proposal.md    research
   tasks      tasks.md       proposal
 ─────────────────────────────────────────────────────────
 ┌──────────┐     ┌──────────┐     ┌────────┐
 │ research │────▶│ proposal │────▶│ tasks  │◆
 └──────────┘     └──────────┘     └────────┘
 ─────────────────────────────────────────────────────────
 d diagram · i install · c duplicate · o open repo · e edit (local)
```

## Local tab

```
 Local schemas
 ─────────────────────────────────────────────────────────
 ~/src/my-schemas
 ▸ team-review
     schema.yaml
     templates/proposal.md
     templates/review.md
 Recent
   ~/tmp/experiments/odd-place/quick
 ─────────────────────────────────────────────────────────
 e edit file · c duplicate · enter detail · : open path
```

## Composer

```
 Composer · draft: my-mix                       ⚠ 1 warning
 ─────────────────────────────┬───────────────────────────
 Palette                      │ Canvas
  research-first              │ ▸ research
   ▸ research                 │   proposal   ← research
     proposal                 │   review     ← proposal
  team-review                 │   tasks ◆    ← review
     review                   │
  minimalist                  │ ⚠ review.md mentions design.md
     tasks                    │   (not on canvas)
 ─────────────────────────────┴───────────────────────────
 a add · x remove · l link · u unlink · g gate · t tracks
 R rename · e edit template · d diagram · w write · s save draft
```

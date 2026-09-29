# Stand in a project and install

Beans epics: `openspec-schema-manager-qjtn` (OpenSpec CLI adapter),
`openspec-schema-manager-wl35` (Project view),
`openspec-schema-manager-0ay4` (Guarded install flow).

## Why

Everything so far is a browser. A user can read schemas and compare them and
still has no way to try one, which was the point: trying a schema in a demo
project, or carefully in a real one, should take seconds.

The project view comes first because installing without it is guesswork. A user
about to install needs to know what the project already has, where each schema
resolves from, and which schema each change is using, or they cannot tell
whether an install will shadow something or collide with it.

## What Changes

- Inside an OpenSpec project, the Project tab shows the project root, its
  default schema, every schema available to it with where it resolves from, and
  every change with the schema it uses and how far along it is.
- All of that comes from OpenSpec and the project's own files. ossm keeps no
  record of what is installed.
- The default schema is read from the project's `openspec/config.yaml`, because
  no OpenSpec command reports it for a project that has no changes yet.
- Outside a project the tab says so and says what would make it work.
- `i` on a registry schema installs it into `openspec/schemas/<name>/`.
- Before anything is written, ossm shows the destination, the files that would
  arrive, and whether that directory already holds something. Nothing is written
  until the user agrees.
- An install into an occupied directory is refused by default and needs a
  separate, explicit confirmation to overwrite.
- After copying, ossm runs OpenSpec's own validation and reports what it says.
  A schema OpenSpec rejects is reported as installed and invalid, not quietly
  accepted.
- Installing never changes the project default. Setting it is a separate action
  with its own confirmation.
- Installing is refused outside a project, saying so rather than failing
  obscurely.
- After an install, the project view and the schema list are refreshed from
  OpenSpec.

## Capabilities

### New Capabilities

- `openspec-adapter`: what ossm asks the `openspec` CLI, what it reads from the
  project's files, and what it does when the CLI is missing or answers badly.
- `project-view`: what the Project tab shows inside a project and outside one.
- `install`: what installing does, what it shows first, what it refuses, and
  what it never does on its own.

### Modified Capabilities

- `schema-listing`: a row can now be installed, so the listing has to say which
  rows can be and which cannot.

## Impact

- `internal/openspec` grows the commands the project view and install need, all
  behind the existing interface with its fake.
- New: the install flow and the project screen in `internal/tui`.
- Install is implemented as fetch, copy and validate, because OpenSpec 1.10.0
  has no install-from-source command. `NOTES.md` records this and the route was
  confirmed end to end before being built on.
- ossm writes into a user's project for the first time. Every write is behind a
  confirmation that names the destination, and an occupied destination needs a
  second one.
- Nothing is done about pinning. The registry is a catalogue and an entry
  tracking a branch installs whatever is there today; recording what was
  installed would be ossm keeping install state, which it does not do.

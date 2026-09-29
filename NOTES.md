# Notes

Where the real OpenSpec and the real registry contradict `briefing/BRIEFING.md`,
this file records what was found and what ossm does about it. It answers
`briefing/docs/open-questions.md` in the same order.

Verified against OpenSpec 1.10.0 on 2026-09-29.

## 1. Registry format

The published registry is reachable and serves an entry shape matching
`docs/registry-entry.md` in `openspec-schema-registry`:

```
https://registry.speclib.org/api/v1/openspec-schemas.json
https://registry.speclib.org/api/v1/schema.json
```

Required fields are `id`, `name`, `description`, `artifacts`, `source.repo` and
`source.path`. Optional: `source.ref`, `requires.openspec`, `status`,
`superseded_by`, `license`, `language`. `source.ref` absent means the
repository's default branch, which is not always `main`; one seeded repository
uses `master`, so ossm asks the repository rather than assuming.

`source.path` cannot be derived from `name`. Seven of the eight schemas checked
upstream sit at `openspec/schemas/<name>/` and the eighth does not, so the path
is read from the entry.

The install destination is not entry data. Every entry installs to
`openspec/schemas/<name>/`, derived from `name`, and a consumer that finds that
directory occupied reports the conflict rather than overwriting it.

The fixtures under `briefing/fixtures/registry/` are illustrative. Tests use
fixtures shaped like the real file.

## 2. Registry tickets

Two open beans in the registry repo cover ground ossm also covers:

- `openspec-schema-registry-ozna`, fetch and cache behaviour
- `openspec-schema-registry-ftt2`, install semantics

Neither is settled there. ossm implements both, and what it learns goes back
into those beans afterwards. Half of install semantics is already settled in
`docs/registry-entry.md` (destination, conflict, pinning); the unsettled half is
upgrading an existing install, which item 5 below constrains.

## 3. OpenSpec schema commands

`openspec schema` offers exactly four subcommands, all marked experimental:

| Command                          | `--json` | Notes                                      |
| -------------------------------- | -------- | ------------------------------------------ |
| `schema which [name] [--all]`    | yes      | resolution source per schema               |
| `schema validate [name]`         | yes      | also `--verbose`                           |
| `schema fork <source> [name]`    | yes      | copies an existing schema into the project |
| `schema init <name>`             | yes      | creates a new project-local schema         |

There is no install-from-source command, and `fork` takes a schema OpenSpec can
already resolve rather than a repository URL.

## 4. Install mechanism

Because no install-from-source command exists, ossm fetches the source folder
itself, writes it to `openspec/schemas/<name>/`, and then runs
`openspec schema validate <name> --json`. All of this lives behind the adapter
in `internal/openspec`, so it can be replaced by a single command if OpenSpec
grows one.

## 5. Update detection

OpenSpec records no provenance for an installed schema. `schema fork` and
`schema init` write a schema folder and nothing that says where it came from or
at which ref, and `schema which` reports only the resolution source (project,
user or built-in). So ossm cannot ask OpenSpec whether an installed schema is
behind its registry entry.

That leaves comparing content against the source, which is honest only when the
entry pins a ref. Against a tracked branch a difference means the schema moved,
was edited locally, or both, and ossm cannot tell which. Milestone 08 decides
what to claim on that basis.

## 6. mermaid-ascii

`github.com/pgavlin/mermaid-ascii` resolves as a Go module at
`v0.0.0-20260322123205-ab8074a98bef`, with no tagged release. Whether it exports
an importable rendering package rather than only a `main` is settled in
milestone 04, before anything depends on it. In process is preferred; shelling
out to a binary shipped in the flake is the fallback.

## 7. Built-in schemas

`openspec schema which --all --json` lists every schema with its resolution
source, which is where the built-in list comes from. Its output shape:

```json
[
  {
    "name": "spec-driven",
    "source": "package",
    "path": "/nix/store/.../lib/openspec/schemas/spec-driven",
    "shadows": []
  }
]
```

`source` is `package` for a built-in schema, `project` for one in
`openspec/schemas/`, and `user` for one in the user's data directory. `shadows`
names the sources this schema hides.

The JSON goes to standard output and the line
`Note: Schema commands are experimental and may change.` goes to standard
error, so the adapter reads standard output alone and does not have to strip a
preamble.

`openspec init` sets `schema: spec-driven` in `openspec/config.yaml`, and
`spec-driven` is deliberately absent from the registry because it is present
without being installed.

## 8. Stores

`openspec store` exists and registers standalone OpenSpec repositories on the
machine. Whether project detection and schema resolution change for a project
pointing at a store is checked in milestone 05, when the adapter is built. If
handling it is cheap it is handled; otherwise the limitation is recorded here.

## 9. Where ossm's registry reading differs from the registry's own validation

ossm validates what it needs to read an entry safely and no more: the six
required fields, `id` shaped `<owner>/<name>`, a non-empty `artifacts`, a known
`status`, and `superseded_by` only on a deprecated entry.

It does not enforce the 100-character cap on `description`, and it ignores a
field it does not know.

Both are deliberate. The cap is the registry's gate to apply before an entry is
merged; a consumer that also refuses it means one bad entry breaks every user
rather than one pull request. And a registry may add a field before the ossm
reading it knows about it, so refusing the whole document over an unknown key
would strand every user until they upgrade.

This is the opposite of what `internal/config` does with `config.yml`, where an
unknown key is an error. A config file is written by the user in front of you,
so a typo is worth stopping for. A registry file is written by someone else and
may be newer than the reader.

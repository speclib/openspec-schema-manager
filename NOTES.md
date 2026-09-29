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

Settled, and confirmed end to end before anything was built on it.

Because no install-from-source command exists, ossm fetches the source folder
itself, writes it to `openspec/schemas/<name>/`, and then runs
`openspec schema validate <name> --json`. All of this lives behind the adapter
in `internal/openspec`, so it can be replaced by a single command if OpenSpec
grows one.

The route was run by hand first: copy a schema into
`openspec/schemas/minimalist/`, then

- `openspec schema validate minimalist --json` answers `"valid": true`
- `openspec schema which minimalist --json` answers `"source": "project"`
- `openspec new change x --schema minimalist` succeeds

An end to end test now asserts the same three things against a schema ossm
installed itself.

Two details the adapter has to know:

`openspec schema validate` exits non-zero when a schema is invalid while still
writing its report to standard output. The report is the answer; the exit status
alone is not, so the adapter reads the output first and only treats the exit
status as an error when nothing parseable came back.

No command reports a project's default schema. `openspec status --change <name>
--json` carries `planningHome.defaultSchema`, but it needs a change to exist,
and a project with no changes is exactly where someone is deciding which schema
to adopt. ossm reads `schema:` from `openspec/config.yaml`.

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

Settled, in process. `github.com/pgavlin/mermaid-ascii` is a Go module with no
tagged release, and it exports what is needed:

```go
render.Render(src string, cfg *diagram.Config) (string, error)
```

`pkg/render` and `pkg/diagram` are importable packages, not a `main`. The
licence is MIT. No binary is shipped in the flake and nothing is shelled out to.

`diagram.DefaultConfig()` returns a `*Config`, not a `Config`, which is easy to
get wrong at the call site. `UseAscii` switches between box-drawing characters
and ASCII, and it is set from whether the environment's locale says UTF-8. That
is the strongest signal a Bubble Tea model can read without querying the
terminal mid-render.

The renderer already draws Mermaid's `{{"x"}}` as a hexagon, so the apply-gate
shape chosen in milestone 03 arrives on screen with nothing further to do.

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

## 10. schema.yaml as the real schemas declare it

The briefing's account in section 3 is accurate as far as it goes. Checked
against the `spec-driven` schema OpenSpec 1.10.0 ships and the two fixtures in
the briefing package, one field is missing from it:

`apply` carries an `instruction` block of its own, alongside `requires` and
`tracks`. `spec-driven` uses it.

Everything else matches: an artifact carries `id`, `generates`, `template`, an
optional `description`, an optional `instruction` and `requires`; a schema
carries `name`, `version`, `description`, `artifacts` and `apply`.

Two details worth recording because they shape the model:

`generates` is sometimes a glob (`specs/**/*.md`) and sometimes one path
(`tasks.md`). It is kept as written, because expanding it needs a directory that
may not exist yet.

`version` is `1` in every schema seen. It is decoded as an integer, so a schema
declaring something else fails with an error naming the field rather than
silently reading as zero.

Schema parsing is strict, unlike registry parsing: an unknown key is an error.
A schema is written by the user or fetched from a repository they chose. A key
ossm silently ignores there is a workflow step that quietly does not happen,
which is worse than a failed parse.

## 11. Fetching a schema needs git, and nothing else is built

`internal/source` fetches with `git init`, `git remote add`, `git fetch --depth
1` and `git checkout --detach FETCH_HEAD`, not with `git clone --branch`. Clone
with a branch fails on a commit SHA, and an entry may pin one.

With no ref, the fetch is `git fetch --depth 1 origin HEAD`, which makes the
repository resolve its own default branch. This is the case the registry
already warns about: one seeded repository uses `master`, so computing `main`
would fail on it.

The briefing allows for fetching raw files when a source is not a git
repository. Nothing is built for it. Every entry in the registry today is a git
repository, and a raw-file fetch cannot list a directory without a
host-specific API, so the fallback would be a GitHub adapter wearing a general
name. `source.Fetcher` is one interface, so adding one later changes nothing
above it.

Tests fetch from git repositories built inside the test, including one whose
default branch is `master`. That exercises the real git commands offline, so the
suite still makes no network request and the case that breaks a naive
implementation is covered rather than assumed.

## 12. OpenSpec requires an artifact description; the briefing says it is optional

OpenSpec 1.10.0 rejects a schema whose artifact declares no `description`:

```
Invalid schema: artifacts.0.description: Invalid input: expected string, received undefined
```

The briefing's section 3 lists `description` as optional, and its own fixture
`research-first` omits it on every artifact. That fixture cannot be installed.

Found by an end to end test, not by reading: the install wrote the schema,
OpenSpec validated it, and the rejection appeared on screen. The install flow
reporting it correctly is the system working, but it is better known before the
install than after.

So ossm follows the CLI. A missing artifact description is a fatal finding in
`internal/schema`, with a message naming OpenSpec as the source of the rule, and
every fixture in the repository now declares one.

## 13. openspec schema fork is not used, and cannot be

`openspec schema fork <source> [name]` copies a schema into the project for
customisation, which looks like exactly what duplication needs. It is not used,
for two reasons found by running it.

It takes a schema name OpenSpec can already resolve, not a path:

```
$ openspec schema fork /path/to/a/schema copied --json
{"forked": false, "error": "Schema '/path/to/a/schema' not found", "available": [...]}
```

A registry schema that has not been installed, and a folder opened by path, are
both invisible to it. Those are the two cases duplication exists for.

And it fails outright when the source is read-only:

```
$ openspec schema fork spec-driven my-fork --json
{"forked": false, "error": "EACCES: permission denied, open '.../openspec/schemas/.fork-staging-WFwIPY/schema.yaml'"}
```

It copies the source preserving its mode and then reopens the copy for writing
to change the name. A schema in the Nix store is mode 444, so forking any
built-in schema fails on every Nix install. Forking a schema already in the
project works, because that copy is writable.

So ossm copies the tree itself and writes every file mode 644. A test duplicates
a deliberately read-only source and asserts the copy is writable.

The copied `schema.yaml` has only its `name:` line rewritten, as text. A schema
carries comments, instruction blocks and a deliberate artifact order, and a YAML
round trip loses the first and can reorder the rest.

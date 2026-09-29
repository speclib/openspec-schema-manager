# Open questions — verify before depending on them

1. **Registry format.** Read `openspec-schemas.json` and `docs/registry-entry.md`
   in the registry repo. The fixtures here are illustrative only. Particularly:
   how an entry names its source repo, the ref, and the schema's path within
   the repo, and where an installed schema is meant to land.
2. **Registry tickets.** Read `.beans/` in the registry repo. Anything planned
   there about tooling takes precedence over this briefing's guesses.
3. **OpenSpec schema commands.** Which exist in the installed version, and
   which have JSON output? Known from docs: `openspec schema init`,
   `openspec schema which [--all]`, a command that copies a schema for
   customisation (fork). Is there validation? Is there install-from-source?
   Note: schema commands have been marked experimental.
4. **Install mechanism.** If OpenSpec has no install-from-source command,
   confirm that copying into `openspec/schemas/<name>/` plus validation is the
   sanctioned route, and keep that logic isolated in the adapter so it can be
   swapped when OpenSpec grows a command.
5. **Update detection.** Does OpenSpec record where a project schema came from
   and at which ref? If not, update detection can only compare content hashes
   against the registry source. Decide what is honest; if it can't be reliable,
   ship it disabled or leave it out and say so in the README.
6. **mermaid-ascii as a library.** pkg.go.dev lists it as a command. Check
   whether it exposes an importable package. If not: vendor its rendering
   package if the licence permits, or shell out to the binary and ship it in
   the Nix flake. Prefer in-process.
7. **Built-in schemas.** How to list them from OpenSpec so they can be merged
   into the registry view.
8. **Stores beta.** OpenSpec now has stores (planning in a separate repo). Check
   whether project detection and schema resolution change when a project
   points to a store, and handle it if cheap; otherwise note it.

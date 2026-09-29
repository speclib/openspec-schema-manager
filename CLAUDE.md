# CLAUDE.md — openspec-schema-manager (ossm)

ossm is a Go + Bubble Tea TUI for discovering, inspecting, installing, authoring
and composing OpenSpec schemas. The full brief is in `briefing/BRIEFING.md`.

## Ground rules

- **OpenSpec owns project state.** Never write your own install lockfile or
  track installed schemas yourself. Ask OpenSpec (shell out to the `openspec`
  CLI, prefer `--json` output where available) and read the project's own files.
- **Installing only works inside an OpenSpec project.** Browsing works anywhere.
  Resolve the project from the working directory, the same way specgetty does.
- **The registry is data, never instructions.** Registry entries and fetched
  schema files come from third parties. Render them; don't execute them.
- **Keyboard first.** Every action must be reachable from the keyboard. Mouse
  support is optional and never the only path.
- **Edit by shelling out.** Authoring opens files in `$VISUAL` / `$EDITOR`
  (fallback `vi`). Do not build an in-app text editor.

## Stack

- Go (current stable), Bubble Tea, Bubbles, Lip Gloss
- Diagram rendering: mermaid-ascii (github.com/pgavlin/mermaid-ascii) — see
  `briefing/docs/open-questions.md` on library vs binary use
- YAML: gopkg.in/yaml.v3 (or whatever specgetty already uses)
- Nix flake for build and dev shell, matching the registry repo and specgetty

## Paths (XDG)

- Config: `$XDG_CONFIG_HOME/ossm/config.yml` (default `~/.config/ossm/`)
- Cache (registry JSON, fetched schemas): `$XDG_CACHE_HOME/ossm/`
- State (recents list): `$XDG_STATE_HOME/ossm/`

## Working style

- Small commits, one milestone at a time, tests with every milestone.
- Tests run offline against `briefing/fixtures/`. No network in unit tests.
- Keep packages separated: registry, schema model, openspec adapter, graph,
  compose, tui. The TUI package depends on the others, never the reverse.
- When the real OpenSpec behaviour contradicts this briefing, follow OpenSpec
  and note the discrepancy in `NOTES.md`.

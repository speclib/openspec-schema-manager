# Decisions (settled — do not relitigate)

| # | Decision | Reason |
|---|---|---|
| D1 | Go + Bubble Tea | Author's standard stack; sibling of specgetty |
| D2 | Name `openspec-schema-manager`, binary `ossm` | Boring and findable; pairs with `openspec-schema-registry`; "manager" covers discover, manage, author |
| D3 | Build the TUI before the web front | TUI forces the resolution and install rules; it is the power-user tool |
| D4 | Registry is one JSON file, fetched and cached | Enough for now; index service only if the registry grows large |
| D5 | Registry entries stay thin; detail is fetched lazily from the schema's repo | Keeps registry PRs small; authors describe their schema once, in their own repo |
| D6 | Entries may pin a ref or track main; that is the author's responsibility | Both are legitimate |
| D7 | ossm does not track install state; OpenSpec does | Don't duplicate core OpenSpec functionality; stay compliant with the tool |
| D8 | Install shells out to OpenSpec | Same as D7 |
| D9 | Install refuses outside an OpenSpec project; browsing works anywhere | Keeps it honest; matches spg behaviour |
| D10 | Local schemas: configured directories, plus a path prompt, plus recents | Regular schemas are found automatically; awkward paths cost you once |
| D11 | Authoring = duplicate, file tree, `e` opens `$EDITOR` | It's plain markdown and YAML; an in-app editor is not worth it |
| D12 | Diagrams via mermaid-ascii | Handles layered layout of branchy flows; Mermaid output is reusable on the web |
| D13 | Composition is build-time, producing a standalone schema | Runtime composition does not exist in OpenSpec |
| D14 | Composer is keyboard-driven; mouse optional later | Dragging in a terminal breaks over SSH/tmux; list-picking is faster once learned |

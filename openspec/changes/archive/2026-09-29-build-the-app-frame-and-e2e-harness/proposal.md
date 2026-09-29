# Build the app frame and the end to end harness

Beans epics: `openspec-schema-manager-zl37` (Bubble Tea shell with tab bar and
help overlay), `openspec-schema-manager-1qv0` (End to end test harness).

## Why

Every screen in milestones 02 to 07 mounts into the same frame: a tab bar, a
status line, a help overlay and the key routing between them. Building that once
now means each later screen is a model with `Update` and `View` and nothing
else. Building it five times, a little differently each time, is the alternative.

The harness has the same argument behind it. The briefing's acceptance criteria
are statements about what a user sees when they run `ossm`, and every one of
them is either asserted by a test or demonstrated by hand once and then assumed.
The harness has to exist before the screens it will check, or the screens get
written without anything holding them to those statements.

## What Changes

- `ossm` opens a full screen application with four tabs, Project, Registry,
  Local and Composer, in that order, starting on Registry when there is no
  OpenSpec project and on Project when there is.
- `tab` and `shift+tab` move between tabs and wrap at both ends. A tab is also
  reachable by its number, `1` to `4`.
- `?` opens a help overlay listing the keys that work where the user is
  standing, global keys separated from the ones the current tab adds. `?`,
  `esc` and `q` close it.
- `q` and `ctrl+c` quit. `q` closes an overlay first when one is open, so it
  never quits out from under a user who meant to dismiss something.
- A status line carries the application name, the version, and a message area
  that later screens write into.
- Each tab shows a placeholder naming what will live there and which milestone
  builds it, so a run of ossm today is honest about what it is.
- The terminal is only taken over when ossm actually starts the interface.
  `--version` and `--help` already avoid it; a run with no terminal attached now
  reports that plainly instead of failing obscurely.
- An end to end harness builds the binary once per run and drives it in a pseudo
  terminal, sending keys and asserting on what is drawn.

## Capabilities

### New Capabilities

- `app-frame`: the tabs, the key routing, the help overlay, the status line and
  what happens with no terminal.
- `e2e-harness`: what the harness guarantees about the binary it drives and the
  environment it drives it in.

### Modified Capabilities

None.

## Impact

- New: `internal/tui` with the frame and its tests, `test/e2e` with the harness
  and its first cases, and a dependency on Bubble Tea, Bubbles and Lip Gloss.
- The harness needs a pseudo terminal, which adds `creack/pty`. It is a test
  dependency only and the flake's vendorHash changes with it.
- The harness drives an isolated environment: `XDG_CONFIG_HOME`,
  `XDG_CACHE_HOME` and `XDG_STATE_HOME` all point inside a temporary directory,
  so an end to end run cannot read or damage the machine's real ossm state.
- Placeholder tabs are a deliberate cost. They will be replaced one at a time,
  and each replacement is a change that removes its placeholder.

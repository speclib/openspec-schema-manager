# Design

## Bubble Tea version

`charm.land/bubbletea/v2`, with `bubbles/v2` and `lipgloss/v2`, matching
specgetty. v2 declares `go 1.25.0`, which is why the flake pins its own Go
toolchain rather than taking the nixpkgs default.

## The frame and its screens

```go
type screen interface {
    Update(tea.Msg) (screen, tea.Cmd)
    View(width, height int) string
    Title() string
    Keys() []keyHelp
}
```

The frame owns the tab bar, the status line, the overlay and the key routing. A
screen owns its pane and nothing else, and reports its own keys so the help
overlay does not carry a second list that can drift.

`Keys()` returning the help entries rather than the frame holding a table of
them is the point. A screen that adds a key and forgets to document it is a
missing method call, not a silently stale overlay.

## Key routing order

1. `ctrl+c` quits, always, before anything else sees it.
2. An open overlay takes the key next. This is what makes `q` close help rather
   than quit.
3. The frame takes the tab keys and `?`.
4. Everything else goes to the selected screen.

Global keys are taken before the screen so that `tab` cannot be captured by a
list widget, which is the usual way a tab bar stops working once a screen has a
text input in it.

## Detecting a terminal

`main` checks whether standard output is a character device before starting
Bubble Tea. Without that, a redirected run fails inside the terminal library
with a message about an ioctl, which tells a user nothing. The check is one
`Stat` call and turns it into a sentence.

`--version` and `--help` are handled before the check, because neither starts
the interface.

## Project detection, for now

The opening tab depends on whether the working directory is inside an OpenSpec
project. Full detection belongs to `internal/openspec` in milestone 05. Here it
is one function that walks up looking for an `openspec/` directory, living in
`internal/openspec` from the start so that milestone 05 extends it rather than
replacing a copy that grew in the TUI.

## The harness

`test/e2e` is its own package, not part of any other. It builds the binary once
with `TestMain` and shares it. Each case gets:

- a temporary directory with `config`, `cache` and `state` inside it, exported
  as the three XDG variables
- a working directory the case chooses, so a case can run inside or outside an
  OpenSpec project
- a pseudo terminal from `creack/pty`, sized to a fixed 100 by 30 so that
  assertions on drawn text do not depend on the machine

Reading is a background goroutine appending into a buffer under a mutex.
`waitFor(text)` polls that buffer until the text appears or a timeout passes,
and on timeout fails with everything drawn so far. Polling beats a single read
because Bubble Tea redraws in several writes and a test that reads once sees
half a frame.

Assertions strip ANSI escape sequences before matching. Matching against styled
output would tie every test to the colour scheme.

## Skipping

A pseudo terminal is not available everywhere. The harness skips the case with a
message naming the reason rather than failing, so the gate stays runnable in a
sandbox. The unit suite carries the frame's own coverage through `Update` and
`View`, which need no terminal, so skipping the end to end cases does not leave
the frame untested.

## What the placeholders say

Each unbuilt tab names what will live there and the milestone that builds it.
A user running ossm today should be able to tell that it is unfinished on
purpose, and a contributor should be able to tell where to start.

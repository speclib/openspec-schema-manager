---
# openspec-schema-manager-1qv0
title: End to end test harness
status: completed
type: epic
priority: normal
created_at: 2026-09-29T18:01:15Z
updated_at: 2026-09-29T18:45:06Z
parent: openspec-schema-manager-6ycu
---

A harness that drives the built binary against fixture registries and throwaway OpenSpec projects, so the acceptance criteria can be asserted rather than demonstrated by hand.

## Summary of Changes

`test/e2e` builds the binary once in `TestMain` and drives it on a pseudo
terminal fixed at 100 by 30. Each case gets its own XDG config, cache and state
roots plus its own `HOME`, and one case asserts that a full run leaves that home
directory empty.

The harness reached its final shape through two corrections worth recording.

A raw byte accumulator is not a screen. Bubble Tea repaints changed cells, so
`Built in milestone 06` is never written as those bytes in that order; the
assertion has to run against a terminal's state, not its input. The pty output
now feeds `github.com/charmbracelet/x/vt`, and `String()` on that emulator is
what a case reads.

The emulator answers a query by writing into an internal pipe, and that write
blocks until something reads it. Bubble Tea sends those queries at startup, so
the first version deadlocked before drawing anything: the read loop stopped
mid-write while holding the lock. A second goroutine copies the emulator's
replies back to the pty, which is what a real terminal does anyway.

Matching is done on whitespace-collapsed text, because the screen pads every
line to its width and wraps long text, so a phrase a user reads as one is
several fragments in the buffer.

`internal/ansi` strips escape sequences and is used by the frame's own tests.
Tying an assertion to styled output ties it to the colour scheme, and a change
of emphasis then breaks tests that care about nothing of the sort.

The cases cover the opening tab in and out of a project, cycling and wrapping,
selecting by number, help opening and closing without quitting, `q` exiting
zero, a redirected run saying it needs a terminal, and `--version` and `--help`
answering without one.

They run inside the nix sandbox with a real pseudo terminal, so `nix flake
check` covers them too. They skip with a reason where one cannot be allocated.

openspec-link: openspec/changes/archive/2026-09-29-build-the-app-frame-and-e2e-harness

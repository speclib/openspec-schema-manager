---
# openspec-schema-manager-zl37
title: Bubble Tea shell with tab bar and help overlay
status: completed
type: epic
priority: normal
created_at: 2026-09-29T18:01:15Z
updated_at: 2026-09-29T18:45:06Z
parent: openspec-schema-manager-6ycu
---

The app frame every screen mounts into: tab bar, status bar, help overlay, and the key routing that keeps every action reachable from the keyboard.

## Summary of Changes

The frame owns the tab bar, the status line, the help overlay and the key
routing. A screen implements `Title`, `Keys`, `Update` and `View` and owns its
pane only. `Keys()` is what the help overlay reads, so a screen that adds a key
and forgets to document it is a missing method call rather than a stale overlay
nobody notices.

Routing order is `ctrl+c`, then an open overlay, then the frame's tab keys, then
the screen. Taking the tab keys before the screen is what stops a list or a text
input from swallowing `tab` later, which is the usual way a tab bar quietly
stops working once a screen grows an input.

Tab navigation keeps working while help is open, so the overlay can be read for
each tab without closing it. `q` closes an overlay before it quits, and
`ctrl+c` quits from anywhere.

Two corrections came out of building it:

`View()` in bubbletea v2.0.10 returns a `tea.View`, not a string, and the
alternate screen is a field on that value rather than a program option. The
model's string rendering is a separate `Render` method, which is also what the
tests drive.

lipgloss v2 renders `Underline(true)` one character at a time, each wrapped in
its own escape pair. The active tab is `Bold` and `Reverse` instead: one escape
pair, and readable output.

The status line is truncated to the window width. It was overflowing at 80
columns and wrapping into the next frame, which is only visible once something
reads the screen the way a terminal does.

The four placeholders each name what will live there and the milestone that
builds it, and list the keys they will handle, so help is honest about what does
nothing yet.

openspec-link: openspec/changes/archive/2026-09-29-build-the-app-frame-and-e2e-harness

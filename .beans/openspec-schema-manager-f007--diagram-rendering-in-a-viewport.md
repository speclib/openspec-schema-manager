---
# openspec-schema-manager-f007
title: Diagram rendering in a viewport
status: completed
type: epic
priority: normal
created_at: 2026-09-29T18:01:15Z
updated_at: 2026-09-29T19:34:44Z
parent: openspec-schema-manager-8ff6
---

Render the Mermaid flowchart to Unicode box drawing with an ASCII fallback, into a string once, then scroll it horizontally and vertically inside its pane.

## Summary of Changes

`d` toggles a flow diagram drawn from the artifact graph through mermaid-ascii,
in process. Open question 6 is closed: `pkg/render` and `pkg/diagram` are
importable, under MIT, so no binary is shipped and nothing is shelled out to.
The renderer already draws Mermaid's `{{"x"}}` as a hexagon, so the gate shape
chosen in milestone 03 arrived on screen for free.

Unicode or ASCII is decided by whether the environment's locale says UTF-8,
which is the strongest signal a Bubble Tea model can read without querying the
terminal mid-render.

The diagram is rendered once into a string and then viewported. Re-rendering per
keystroke would be visible on a branchy graph.

The viewport slices by rune, not by byte. Slicing by byte cuts a box-drawing
character into mojibake, and a test asserts no replacement character survives a
horizontal scroll. Scrolling clamps at all four edges, and a diagram that fits
its pane never moves.

A cyclic schema says it cannot be drawn and quotes the finding naming the
artifacts on the cycle. A schema with no artifacts says there is nothing to
draw. Neither shows an empty pane.

openspec-link: openspec/changes/archive/2026-09-29-show-a-schema-in-detail

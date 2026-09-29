---
# openspec-schema-manager-gypt
title: 08 Polish
status: completed
type: milestone
priority: normal
created_at: 2026-09-29T18:00:33Z
updated_at: 2026-09-29T21:00:10Z
---

What makes it presentable as an alpha base: update detection if it can be honest, README with recorded demos, and help text on every screen.

## Summary of Changes

Milestone 08 is one OpenSpec change, `finish-the-proof-of-concept`, adding
`update-detection` and `acceptance` and extending `project-view`.

The open question the briefing left is answered: update detection ships as a
comparison that says what it cannot tell, because OpenSpec records no
provenance. Shipping it as "update available" would have been a lie and leaving
it out would have wasted what ossm does know.

Every acceptance criterion is now a case driving the built binary rather than a
claim, including the offline one, which removes the source repository and the
registry file before reopening ossm.

The README, the recorded demos and the generated key reference make the
repository readable by someone who has not followed the eight milestones.

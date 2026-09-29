# Fixtures

Illustrative only. The registry fixture's field names are a guess; replace it
with a trimmed copy of the real `openspec-schemas.json` once read, and adjust
the parser to the real format. The schema fixtures follow OpenSpec's documented
`schema.yaml` shape and are good for testing the graph, diagram and composer.
`research-first/templates/proposal.md` deliberately mentions `design.md`, which
that schema does not generate, to exercise the composer's dangling-reference
warning.

# Tasks

## 1. The model

- [x] 1.1 `Artifact` carrying what it declares plus its source directory, schema
      name, ref and original id
- [x] 1.2 `Composition` holding artifacts, gates and the tracked file
- [x] 1.3 `FromSchema` starts a composition from an existing schema
- [x] 1.4 `Add` filters requirements to what is on the canvas and reports a
      collision rather than renaming
- [x] 1.5 Adding an artifact later does not restore a dropped edge
- [x] 1.6 Sources are recorded once however many times they are added
- [x] 1.7 Table test over each
- [x] 1.8 `go test ./...` passes

## 2. Editing

- [x] 2.1 `Remove` takes the artifact, every edge naming it and its gate
- [x] 2.2 `Link` adds a requirement; a duplicate changes nothing; a self-link is
      refused
- [x] 2.3 `Unlink` removes a requirement
- [x] 2.4 `ToggleGate` and `SetTracks`
- [x] 2.5 `Rename` updates every requirement and the gate; a collision is
      refused; renaming to the same id is not a collision
- [x] 2.6 Table test over each
- [x] 2.7 `go test ./...` passes

## 3. Validation

- [x] 3.1 Build the `schema.Schema` the composition describes and run
      `schema.Validate` on it
- [x] 3.2 Add the one thing only a composition can be wrong about: nothing
      selected
- [x] 3.3 Test that a cycle, an unresolved requirement, a missing gate, a
      missing tracked file and a path collision are all reported
- [x] 3.4 Test that the composer and the schema validator agree
- [x] 3.5 `go test ./...` passes

## 4. Template scanning

- [x] 4.1 Scan each canvas artifact's template text and instruction for ids and
      generated paths of source artifacts not on the canvas
- [x] 4.2 Match on word boundaries, so `design` does not match `designer`
- [x] 4.3 Each warning names the file and the reference
- [x] 4.4 A template that cannot be read is reported once
- [x] 4.5 Adding the missing artifact clears the warning
- [x] 4.6 Table test over each
- [x] 4.7 `go test ./...` passes

## 5. Writing

- [x] 5.1 Generate `schema.yaml` from the canvas: name, artifacts in canvas
      order with their requirements, apply gate and tracked file
- [x] 5.2 Open with a comment block recording each artifact's source schema, ref
      and original id
- [x] 5.3 Copy each template byte for byte; two sources sharing a template path
      get distinct destinations and each artifact's `template` names its own
- [x] 5.4 Stage and rename; a failure leaves nothing partial
- [x] 5.5 Refuse an occupied destination, an invalid composition, a bad name and
      no configured directory
- [x] 5.6 Test that what is written parses and validates with `internal/schema`
- [x] 5.7 `go test ./...` passes

## 6. Drafts

- [x] 6.1 Save a composition as JSON under the drafts directory
- [x] 6.2 List drafts most recently saved first
- [x] 6.3 Resume restores artifacts, edges, gates and the tracked file
- [x] 6.4 A corrupt draft is reported and the others still list
- [x] 6.5 Resuming reports artifacts whose source directory has gone, and the
      rest stays editable
- [x] 6.6 Table test over each
- [x] 6.7 `go test ./...` passes

## 7. The screen

- [x] 7.1 Palette and canvas side by side, with the findings below
- [x] 7.2 Movement between and within the panes
- [x] 7.3 `a` add, `x` remove, `l` link, `u` unlink, `g` gate, `t` tracks,
      `R` rename, `d` diagram, `w` write, `s` save draft
- [x] 7.4 Each prompt is its own state and `Capturing()` reports true in all of
      them
- [x] 7.5 The diagram uses the same renderer as the detail view and says so when
      a cycle stops it
- [x] 7.6 `Keys()` lists what it handles in each state
- [x] 7.7 Test each through `Update` and `View`
- [x] 7.8 `go test ./...` passes

## 8. Sending a schema to the composer

- [x] 8.1 `p` on a registry, project or local schema adds it as a source
- [x] 8.2 The tab it was pressed on stays selected
- [x] 8.3 A schema that cannot be read reports the reason
- [x] 8.4 Test from each tab
- [x] 8.5 `go test ./...` passes

## 9. End to end

- [x] 9.1 Two fixture schemas added as sources from the Local tab
- [x] 9.2 Artifacts from both added to the canvas, linked, gated and tracked
- [x] 9.3 A cycle is reported as it is made
- [x] 9.4 A dangling template reference is warned about
- [x] 9.5 The composition is written and OpenSpec validates the result
- [x] 9.6 A change can be created with the written schema

## 10. Close out

- [x] 10.1 Raise the coverage floors to the measured values, with
      `internal/compose` among the highest
- [x] 10.2 `bash scripts/coverage-gate.sh` and `nix flake check` pass
- [x] 10.3 Fill in the four epics' summaries and mark them completed; close
      milestone 07
- [x] 10.4 Archive the change, commit and push

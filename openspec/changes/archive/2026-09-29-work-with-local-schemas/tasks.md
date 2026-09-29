# Tasks

## 1. Scanning

- [x] 1.1 Scan each configured directory one level deep for folders holding a
      `schema.yaml`
- [x] 1.2 A folder whose schema does not parse is listed and marked unreadable
- [x] 1.3 A configured directory that does not exist is reported and skipped
- [x] 1.4 A configured directory holding no schemas is listed as holding none
- [x] 1.5 Table test over all four cases
- [x] 1.6 `go test ./...` passes

## 2. Recents

- [x] 2.1 Read and write the recents file in the state directory, atomically
- [x] 2.2 Most recent first, no duplicates, capped at `recents_cap`
- [x] 2.3 A missing or corrupt file reads as empty and does not stop a write
- [x] 2.4 A remembered path that no longer holds a schema is reported as missing
- [x] 2.5 Table test over adding, re-adding, the cap, corruption and a cap of
      zero
- [x] 2.6 `go test ./...` passes

## 3. Duplication

- [x] 3.1 Copy a schema folder into a destination, writing every file mode 644
- [x] 3.2 Rewrite only the `name:` line in the copied `schema.yaml`
- [x] 3.3 Refuse an empty name, a name with a separator, `.` and `..`
- [x] 3.4 Refuse an occupied destination without writing anything
- [x] 3.5 Test that comments, instructions and artifact order survive, and that
      a read-only source can be duplicated
- [x] 3.6 `go test ./...` passes

## 4. The path prompt

- [x] 4.1 A prompt model with the typed path, an error line and completion
- [x] 4.2 `tab` completes against existing directories, filling in the longest
      common prefix and showing the matches
- [x] 4.3 A leading `~/` expands
- [x] 4.4 A path with no schema, and a path that does not exist, each report and
      leave the prompt open
- [x] 4.5 `esc` closes without opening anything
- [x] 4.6 Test each, including a digit typed into the prompt
- [x] 4.7 `go test ./...` passes

## 5. The frame

- [x] 5.1 `:` and `ctrl+o` open the prompt from any tab
- [x] 5.2 While it is open the frame takes `ctrl+c` and nothing else
- [x] 5.3 Opening a schema from the prompt selects the Local tab
- [x] 5.4 Test from a tab other than Local
- [x] 5.5 `go test ./...` passes

## 6. The Local screen

- [x] 6.1 One list over the configured directories, the recents and `--path`
- [x] 6.2 Each schema shows its declared name, artifact count and directory
- [x] 6.3 No directories configured says how to configure one and offers the
      prompt
- [x] 6.4 `enter` opens the detail from the folder, with no fetch
- [x] 6.5 `c` duplicates: prompt for a name, show the destination, refuse a
      collision
- [x] 6.6 Movement and scrolling
- [x] 6.7 `Keys()` lists what it handles
- [x] 6.8 Test each through `Update` and `View`
- [x] 6.9 `go test ./...` passes

## 7. The file tree and editing

- [x] 7.1 `t` shows the schema's files as a tree in a stable order
- [x] 7.2 A schema with no `templates/` lists `schema.yaml` alone
- [x] 7.3 `e` runs `$VISUAL`, `$EDITOR` or `vi` through `tea.ExecProcess`,
      splitting the value on spaces and using no shell
- [x] 7.4 An editor that fails is reported and the interface resumes
- [x] 7.5 On return the schema is re-read and re-validated and the findings are
      shown
- [x] 7.6 Test the editor resolution as a pure function; test the round trip
      with a stub editor that writes to the file
- [x] 7.7 `go test ./...` passes

## 8. End to end

- [x] 8.1 A configured schemas directory holding two schemas is listed
- [x] 8.2 `:` opens a path prompt and a typed path opens a schema
- [x] 8.3 `c` duplicates a schema and the copy declares the new name
- [x] 8.4 `t` shows the file tree
- [x] 8.5 `e` with a stub editor that appends to a template returns to a
      re-validated schema

## 9. Close out

- [x] 9.1 Record in `NOTES.md` why `openspec schema fork` is not used, with the
      EACCES failure on a read-only source
- [x] 9.2 Raise the coverage floors to the measured values
- [x] 9.3 `bash scripts/coverage-gate.sh` and `nix flake check` pass
- [x] 9.4 Fill in the four epics' summaries and mark them completed; close
      milestone 06
- [x] 9.5 Archive the change, commit and push

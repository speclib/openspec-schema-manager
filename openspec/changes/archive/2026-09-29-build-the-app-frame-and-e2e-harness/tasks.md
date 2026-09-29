# Tasks

## 1. Dependencies

- [x] 1.1 Add `charm.land/bubbletea/v2`, `charm.land/bubbles/v2` and
      `charm.land/lipgloss/v2`
- [x] 1.2 Add `github.com/creack/pty` for the harness
- [x] 1.3 `go mod tidy`, update the vendorHash in `flake.nix` and `package.nix`
- [x] 1.4 `nix build` passes

## 2. Project detection

- [x] 2.1 `internal/openspec.FindProjectRoot(dir)` walks up looking for an
      `openspec/` directory and returns the root or reports that there is none
- [x] 2.2 Stop at the filesystem root, and do not follow a symlink out
- [x] 2.3 Table test: at the root, in a subdirectory, outside any project, and
      with a file named `openspec` rather than a directory
- [x] 2.4 `go test ./...` passes

## 3. The frame

- [x] 3.1 Define the `screen` interface: `Update`, `View`, `Title`, `Keys`
- [x] 3.2 The frame model holds the four screens, the selection, the overlay
      state, the status message and the window size
- [x] 3.3 Draw the tab bar, the pane and the status line with the name and
      version
- [x] 3.4 Open on Project inside a project and on Registry outside one
- [x] 3.5 `go test ./...` passes

## 4. Key routing

- [x] 4.1 `ctrl+c` quits first, from anywhere
- [x] 4.2 An open overlay takes the key next, so `q` closes help
- [x] 4.3 `tab`, `shift+tab` and `1` to `4` are taken by the frame before the
      screen sees them
- [x] 4.4 Cycling wraps at both ends; a digit with no tab changes nothing
- [x] 4.5 Everything else goes to the selected screen
- [x] 4.6 Table test over every routing case above, driving `Update` directly
- [x] 4.7 `go test ./...` passes

## 5. Help overlay

- [x] 5.1 `?` opens it, listing global keys and the selected tab's keys under
      separate headings
- [x] 5.2 The tab's keys come from that screen's `Keys()`, so the overlay cannot
      drift from what the screen handles
- [x] 5.3 `?`, `esc` and `q` close it
- [x] 5.4 Switching tabs with help open changes what it lists
- [x] 5.5 Test each of those through `Update` and `View`
- [x] 5.6 `go test ./...` passes

## 6. Placeholders

- [x] 6.1 Each of the four screens names what will live there and the milestone
      that builds it
- [x] 6.2 Each returns the keys it will handle, so the help overlay is honest
      about what does nothing yet
- [x] 6.3 `go test ./...` passes

## 7. No terminal

- [x] 7.1 `main` checks standard output is a character device before starting
      Bubble Tea
- [x] 7.2 Without one, write a message naming the problem and exit non-zero,
      emitting no escape sequence
- [x] 7.3 `--version` and `--help` are handled before the check
- [x] 7.4 Test through `run` with a non-terminal writer
- [x] 7.5 `go test ./...` passes

## 8. The harness

- [x] 8.1 `test/e2e` builds the binary once in `TestMain` and shares it
- [x] 8.2 A failing build fails the run with the compiler output, once
- [x] 8.3 Each case gets a temporary XDG config, cache and state root and a
      working directory it chooses
- [x] 8.4 Start the binary on a pseudo terminal fixed at 100 by 30
- [x] 8.5 Read into a buffer in the background; `waitFor` polls it and fails
      with everything drawn so far
- [x] 8.6 Strip ANSI escape sequences before matching
- [x] 8.7 Skip with a reason when a pseudo terminal cannot be allocated
- [x] 8.8 Assert that a run writes nothing outside its temporary roots

## 9. First end to end cases

- [x] 9.1 Starting outside a project opens on Registry and draws all four tabs
- [x] 9.2 Starting inside a project opens on Project
- [x] 9.3 `tab` moves to the next tab and wraps at the end
- [x] 9.4 `?` opens help and `q` closes it without quitting
- [x] 9.5 `q` with nothing open quits, and the process exits zero
- [x] 9.6 A redirected run says it needs a terminal and exits non-zero

## 10. Close out

- [x] 10.1 Raise the coverage floors for `internal/tui` and `internal/openspec`
      to their measured values
- [x] 10.2 `bash scripts/coverage-gate.sh` and `nix flake check` pass
- [x] 10.3 Fill in both epics' summaries and mark them completed; close
      milestone 01
- [x] 10.4 Archive the change and commit

// Package tui holds the Bubble Tea models, one per screen, and the frame they
// mount into.
//
// This package depends on the others and none of them depends on it. A test in
// internal/arch enforces that.
//
// Every action is reachable from the keyboard. Editing a file suspends the TUI
// and shells out to $VISUAL, $EDITOR or vi rather than being handled in an
// in-app editor.
package tui

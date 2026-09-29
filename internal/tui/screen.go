package tui

import tea "charm.land/bubbletea/v2"

type KeyHelp struct {
	Key         string
	Description string
}

type Screen interface {
	Title() string
	Keys() []KeyHelp
	// Capturing reports that the screen is taking text, so the frame must
	// stop claiming tab keys and digits for itself. Without it, typing 4 into
	// a filter jumps to the Composer.
	Capturing() bool
	Update(tea.Msg) (Screen, tea.Cmd)
	View(width, height int) string
}

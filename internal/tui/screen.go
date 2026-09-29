package tui

import tea "charm.land/bubbletea/v2"

type KeyHelp struct {
	Key         string
	Description string
}

type Screen interface {
	Title() string
	Keys() []KeyHelp
	Update(tea.Msg) (Screen, tea.Cmd)
	View(width, height int) string
}

package tui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/speclib/openspec-schema-manager/internal/registry"
	"github.com/speclib/openspec-schema-manager/internal/source"
)

// sendToComposerCmd resolves a listed schema and hands it to the composer as a
// source. It resolves the same way opening the detail does, so a registry
// schema is fetched and a local one is read from where it is.
func sendToComposerCmd(r *Resolver, row registry.Row) tea.Cmd {
	if r == nil {
		return func() tea.Msg {
			return SourceAdded{Err: errNoResolver}
		}
	}

	return func() tea.Msg {
		resolved, err := r.resolve(row, false)
		if err != nil {
			return SourceAdded{Err: err}
		}

		return SourceAdded{
			Schema: resolved.Schema,
			Dir:    resolved.Dir,
			Ref:    refOf(resolved),
		}
	}
}

func refOf(resolved source.Resolved) string {
	if resolved.Source.Ref != "" {
		return resolved.Source.Ref
	}
	if resolved.Origin == source.OriginRegistry {
		return "default branch"
	}
	return ""
}

var errNoResolver = contextError("no way to read a schema is configured")

type contextError string

func (e contextError) Error() string { return string(e) }

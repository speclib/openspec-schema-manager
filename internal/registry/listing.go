package registry

import (
	"sort"
	"strings"
)

const (
	OriginRegistry = "registry"
	OriginBuiltIn  = "built-in"
	OriginProject  = "project"
	OriginUser     = "user"
)

type Row struct {
	Name         string
	Description  string
	Origin       string
	Artifacts    []string
	Entry        *Entry
	Path         string
	Installable  bool
	Deprecated   bool
	SupersededBy string
	Ref          string
	Pinned       bool
	sortKey      string
	originRank   int
	tiebreak     string
}

func (r Row) ArtifactCount() int { return len(r.Artifacts) }

func (r Row) Shape() string {
	switch len(r.Artifacts) {
	case 0:
		return ""
	case 1, 2, 3:
		return strings.Join(r.Artifacts, " → ")
	default:
		return r.Artifacts[0] + " → … → " + r.Artifacts[len(r.Artifacts)-1]
	}
}

func (r Row) RefLabel() string {
	switch {
	case r.Origin == OriginBuiltIn:
		return ""
	case r.Pinned:
		return r.Ref
	default:
		return "default branch"
	}
}

func RowFromEntry(e Entry) Row {
	entry := e

	return Row{
		Name:         e.Name,
		Description:  e.Description,
		Origin:       OriginRegistry,
		Artifacts:    e.Artifacts,
		Entry:        &entry,
		Installable:  true,
		Deprecated:   e.Deprecated(),
		SupersededBy: e.SupersededBy,
		Ref:          e.Source.Ref,
		Pinned:       e.Pinned(),
		sortKey:      strings.ToLower(e.Name),
		originRank:   1,
		tiebreak:     e.ID,
	}
}

func RowFromInstalled(name, source, path string, artifacts []string) Row {
	origin, rank := OriginBuiltIn, 0

	switch source {
	case "project":
		origin, rank = OriginProject, 2
	case "user":
		origin, rank = OriginUser, 3
	}

	return Row{
		Name:        name,
		Origin:      origin,
		Artifacts:   artifacts,
		Path:        path,
		Installable: false,
		sortKey:     strings.ToLower(name),
		originRank:  rank,
		tiebreak:    path,
	}
}

func Merge(entries []Entry, installed []Row) []Row {
	rows := make([]Row, 0, len(entries)+len(installed))

	for _, e := range entries {
		rows = append(rows, RowFromEntry(e))
	}
	rows = append(rows, installed...)

	sort.SliceStable(rows, func(i, j int) bool {
		a, b := rows[i], rows[j]
		if a.sortKey != b.sortKey {
			return a.sortKey < b.sortKey
		}
		if a.originRank != b.originRank {
			return a.originRank < b.originRank
		}
		return a.tiebreak < b.tiebreak
	})

	return rows
}

func Filter(rows []Row, query string) []Row {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return rows
	}

	matched := make([]Row, 0, len(rows))
	for _, r := range rows {
		if r.Matches(query) {
			matched = append(matched, r)
		}
	}

	return matched
}

func (r Row) Matches(lowered string) bool {
	if strings.Contains(strings.ToLower(r.Name), lowered) {
		return true
	}
	if strings.Contains(strings.ToLower(r.Description), lowered) {
		return true
	}
	if r.Entry != nil && strings.Contains(strings.ToLower(r.Entry.ID), lowered) {
		return true
	}

	return false
}

package registry

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

type Document struct {
	Schema  string  `json:"$schema,omitempty"`
	Schemas []Entry `json:"schemas"`
}

var idPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]*/[a-z0-9][a-z0-9._-]*$`)

func Parse(source string, raw []byte) (Document, error) {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(raw, &top); err != nil {
		return Document{}, fmt.Errorf("reading %s: %w", source, err)
	}

	schemas, ok := top["schemas"]
	if !ok {
		return Document{}, fmt.Errorf("reading %s: no schemas key; this is not a registry", source)
	}

	var doc Document
	if err := json.Unmarshal(raw, &doc); err != nil {
		return Document{}, fmt.Errorf("reading %s: %w", source, err)
	}

	var probe []json.RawMessage
	if err := json.Unmarshal(schemas, &probe); err != nil {
		return Document{}, fmt.Errorf("reading %s: schemas is not an array", source)
	}

	for i, entry := range doc.Schemas {
		if err := validate(entry, i); err != nil {
			return Document{}, fmt.Errorf("reading %s: %w", source, err)
		}
	}

	return doc, nil
}

func validate(e Entry, position int) error {
	where := fmt.Sprintf("entry %s", e.ID)
	if e.ID == "" {
		return fmt.Errorf("entry at position %d has no id", position)
	}

	if !idPattern.MatchString(e.ID) {
		return fmt.Errorf("%s: id must be lowercase and shaped <owner>/<name>", where)
	}

	for _, missing := range []struct {
		field string
		empty bool
	}{
		{field: "name", empty: strings.TrimSpace(e.Name) == ""},
		{field: "description", empty: strings.TrimSpace(e.Description) == ""},
		{field: "artifacts", empty: len(e.Artifacts) == 0},
		{field: "source.repo", empty: strings.TrimSpace(e.Source.Repo) == ""},
		{field: "source.path", empty: strings.TrimSpace(e.Source.Path) == ""},
	} {
		if missing.empty {
			return fmt.Errorf("%s: %s is required and is missing or empty", where, missing.field)
		}
	}

	if e.Status != "" && e.Status != StatusActive && e.Status != StatusDeprecated {
		return fmt.Errorf("%s: status is %q; it is active or deprecated", where, e.Status)
	}

	if e.SupersededBy != "" && !e.Deprecated() {
		return fmt.Errorf("%s: superseded_by is set on an entry that is not deprecated", where)
	}

	return nil
}

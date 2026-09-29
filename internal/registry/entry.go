package registry

import "strings"

const (
	StatusActive     = "active"
	StatusDeprecated = "deprecated"

	DefaultLanguage = "en"
)

type Source struct {
	Repo string `json:"repo"`
	Path string `json:"path"`
	Ref  string `json:"ref,omitempty"`
}

type Requires struct {
	OpenSpec string `json:"openspec,omitempty"`
}

type Entry struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Artifacts    []string `json:"artifacts"`
	Source       Source   `json:"source"`
	Requires     Requires `json:"requires,omitzero"`
	Status       string   `json:"status,omitempty"`
	SupersededBy string   `json:"superseded_by,omitempty"`
	License      string   `json:"license,omitempty"`
	Language     string   `json:"language,omitempty"`
}

func (e Entry) EffectiveStatus() string {
	if e.Status == "" {
		return StatusActive
	}
	return e.Status
}

func (e Entry) Deprecated() bool {
	return e.EffectiveStatus() == StatusDeprecated
}

func (e Entry) EffectiveLanguage() string {
	if e.Language == "" {
		return DefaultLanguage
	}
	return e.Language
}

func (e Entry) LicenseKnown() bool {
	return e.License != ""
}

func (e Entry) Pinned() bool {
	return e.Source.Ref != ""
}

func (e Entry) RefLabel() string {
	if e.Pinned() {
		return e.Source.Ref
	}
	return "default branch"
}

func (e Entry) Owner() string {
	owner, _, found := strings.Cut(e.ID, "/")
	if !found {
		return ""
	}
	return owner
}

func (e Entry) MinimumOpenSpec() (string, bool) {
	return e.Requires.OpenSpec, e.Requires.OpenSpec != ""
}

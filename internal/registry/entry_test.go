package registry

import "testing"

func TestAnAbsentOptionalFieldHasADefinedMeaning(t *testing.T) {
	t.Parallel()

	bare := Entry{
		ID:          "speclib/tinychange",
		Name:        "tinychange",
		Description: "Lean specs to tasks workflow for very small changes.",
		Artifacts:   []string{"specs", "tasks"},
		Source: Source{
			Repo: "https://github.com/speclib/openspec-tinychange-schema",
			Path: "openspec/schemas/tinychange",
		},
	}

	if got := bare.EffectiveStatus(); got != StatusActive {
		t.Errorf("EffectiveStatus = %q, want %q", got, StatusActive)
	}
	if bare.Deprecated() {
		t.Error("an entry with no status is deprecated")
	}
	if got := bare.EffectiveLanguage(); got != DefaultLanguage {
		t.Errorf("EffectiveLanguage = %q, want %q", got, DefaultLanguage)
	}
	if bare.LicenseKnown() {
		t.Error("an entry with no licence reports the licence as known")
	}
	if bare.Pinned() {
		t.Error("an entry with no ref reports itself as pinned")
	}
	if got := bare.RefLabel(); got != "default branch" {
		t.Errorf("RefLabel = %q, want %q", got, "default branch")
	}
	if got := bare.RefLabel(); got == "main" {
		t.Error("an absent ref was reported as main; the default branch is not always main")
	}
	if _, ok := bare.MinimumOpenSpec(); ok {
		t.Error("an entry with no requires states a minimum OpenSpec version")
	}
	if got := bare.Owner(); got != "speclib" {
		t.Errorf("Owner = %q, want %q", got, "speclib")
	}
}

func TestAPresentOptionalFieldIsUsed(t *testing.T) {
	t.Parallel()

	full := Entry{
		ID:       "lukk17/e2e-runbooks",
		Source:   Source{Ref: "v0.2.0"},
		Status:   StatusDeprecated,
		License:  "MIT",
		Language: "nl",
		Requires: Requires{OpenSpec: ">=1.10"},
	}

	if !full.Deprecated() {
		t.Error("a deprecated entry does not report itself as deprecated")
	}
	if got := full.EffectiveLanguage(); got != "nl" {
		t.Errorf("EffectiveLanguage = %q, want nl", got)
	}
	if !full.LicenseKnown() {
		t.Error("an entry with a licence reports it as unknown")
	}
	if !full.Pinned() {
		t.Error("an entry with a ref does not report itself as pinned")
	}
	if got := full.RefLabel(); got != "v0.2.0" {
		t.Errorf("RefLabel = %q, want v0.2.0", got)
	}

	min, ok := full.MinimumOpenSpec()
	if !ok || min != ">=1.10" {
		t.Errorf("MinimumOpenSpec = (%q, %v), want (\">=1.10\", true)", min, ok)
	}
}

func TestOwnerOfAMalformedID(t *testing.T) {
	t.Parallel()

	if got := (Entry{ID: "noslash"}).Owner(); got != "" {
		t.Errorf("Owner of an id with no slash = %q, want empty", got)
	}
}

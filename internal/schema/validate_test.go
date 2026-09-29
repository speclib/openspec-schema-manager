package schema

import (
	"path/filepath"
	"strings"
	"testing"
)

func messages(f Findings) string {
	var b strings.Builder
	for _, finding := range f {
		b.WriteString(finding.String())
		b.WriteString("\n")
	}
	return b.String()
}

func TestAValidSchemaReportsNothing(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"chain.yaml", "branchy.yaml", "single.yaml", "wide.yaml", "spec-driven.yaml"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			findings := Validate(load(t, name))
			if !findings.Valid() {
				t.Errorf("%s does not validate:\n%s", name, messages(findings))
			}
			if len(findings.Warnings()) != 0 {
				t.Errorf("%s produced warnings:\n%s", name, messages(findings.Warnings()))
			}
		})
	}
}

func TestFatalProblems(t *testing.T) {
	t.Parallel()

	cases := []struct {
		fixture string
		want    []string
	}{
		{fixture: "no-artifacts.yaml", want: []string{"declares no artifacts"}},
		{fixture: "duplicate.yaml", want: []string{"tasks", "more than once"}},
		{fixture: "unknown-requirement.yaml", want: []string{"tasks", `requires "specs"`, "no artifact declares"}},
		{fixture: "self.yaml", want: []string{"tasks", "requires itself"}},
		{fixture: "cycle.yaml", want: []string{"cycle", "design", "tasks"}},
		{fixture: "unknown-gate.yaml", want: []string{"apply.requires", `"verify"`}},
		{fixture: "no-gate.yaml", want: []string{"apply.requires names no artifact", "apply.tracks names no file"}},
	}

	for _, tc := range cases {
		t.Run(tc.fixture, func(t *testing.T) {
			t.Parallel()

			findings := Validate(load(t, tc.fixture))
			if findings.Valid() {
				t.Fatalf("%s validates but should not:\n%s", tc.fixture, messages(findings))
			}

			got := messages(findings.Fatal())
			for _, want := range tc.want {
				if !strings.Contains(got, want) {
					t.Errorf("the findings do not mention %q:\n%s", want, got)
				}
			}
		})
	}
}

func TestASelfReferenceIsNotReportedAsACycle(t *testing.T) {
	t.Parallel()

	got := messages(Validate(load(t, "self.yaml")).Fatal())

	if !strings.Contains(got, "requires itself") {
		t.Errorf("a self-reference is not named as one:\n%s", got)
	}
	if strings.Contains(got, "cycle") {
		t.Errorf("a self-reference was also reported as a cycle:\n%s", got)
	}
}

func TestACycleNamesItsMembersInOrder(t *testing.T) {
	t.Parallel()

	got := messages(Validate(load(t, "cycle.yaml")).Fatal())

	if !strings.Contains(got, "→") {
		t.Errorf("the cycle does not show how its members connect:\n%s", got)
	}
	if strings.Contains(got, "proposal") {
		t.Errorf("an artifact outside the cycle was named as part of it:\n%s", got)
	}
}

func TestACycleIsReportedOnce(t *testing.T) {
	t.Parallel()

	var cycles int
	for _, f := range Validate(load(t, "cycle.yaml")).Fatal() {
		if strings.Contains(f.Message, "cycle") {
			cycles++
		}
	}

	if cycles != 1 {
		t.Errorf("the same cycle was reported %d times", cycles)
	}
}

func TestEveryProblemIsReportedInOnePass(t *testing.T) {
	t.Parallel()

	body := `name: many
version: 1
description: several problems at once
artifacts:
  - id: specs
    generates: notes.md
    template: spec.md
  - id: specs
    generates: notes.md
    template: spec.md
  - id: tasks
    generates: tasks.md
    template: tasks.md
    requires: [design]
apply:
  requires: [verify]
  tracks: tasks.md
`

	s, err := Parse("many.yaml", []byte(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := messages(Validate(s))

	for _, want := range []string{"more than once", `requires "design"`, `apply.requires names "verify"`, "generate"} {
		if !strings.Contains(got, want) {
			t.Errorf("the findings do not mention %q:\n%s", want, got)
		}
	}
}

func TestWarnings(t *testing.T) {
	t.Parallel()

	t.Run("two artifacts generating the same path", func(t *testing.T) {
		t.Parallel()

		findings := Validate(load(t, "collides.yaml"))
		if !findings.Valid() {
			t.Errorf("a collision made the schema invalid:\n%s", messages(findings.Fatal()))
		}

		got := messages(findings.Warnings())
		if !strings.Contains(got, "notes.md") || !strings.Contains(got, "overwrites") {
			t.Errorf("the collision is not reported:\n%s", got)
		}
	})

	t.Run("an artifact nothing reaches", func(t *testing.T) {
		t.Parallel()

		findings := Validate(load(t, "orphan.yaml"))
		if !findings.Valid() {
			t.Errorf("an orphan made the schema invalid:\n%s", messages(findings.Fatal()))
		}

		got := messages(findings.Warnings())
		if !strings.Contains(got, "sketch") {
			t.Errorf("the orphan is not reported:\n%s", got)
		}
		if strings.Contains(got, "specs") {
			t.Errorf("an artifact something requires was reported as an orphan:\n%s", got)
		}
	})

	t.Run("an artifact with no template", func(t *testing.T) {
		t.Parallel()

		findings := Validate(load(t, "no-template.yaml"))
		if !findings.Valid() {
			t.Errorf("a missing template declaration made the schema invalid:\n%s", messages(findings.Fatal()))
		}
		if got := messages(findings.Warnings()); !strings.Contains(got, "no template") {
			t.Errorf("the missing template declaration is not reported:\n%s", got)
		}
	})
}

func TestASchemaWithNoNameIsFatal(t *testing.T) {
	t.Parallel()

	s := load(t, "chain.yaml")
	s.Name = "  "

	if got := messages(Validate(s).Fatal()); !strings.Contains(got, "declares no name") {
		t.Errorf("a schema with no name validates:\n%s", got)
	}
}

func TestAnArtifactWithNoIDIsFatal(t *testing.T) {
	t.Parallel()

	s := load(t, "chain.yaml")
	s.Artifacts[0].ID = ""

	got := messages(Validate(s).Fatal())
	if !strings.Contains(got, "position 0") {
		t.Errorf("an artifact with no id is not reported by position:\n%s", got)
	}
}

func TestTemplateFilesAreCheckedOnlyWithADirectory(t *testing.T) {
	t.Parallel()

	complete := filepath.Join("testdata", "folder-complete")
	missing := filepath.Join("testdata", "folder-missing")

	s, err := Load(missing)
	if err != nil {
		t.Fatalf("loading the fixture: %v", err)
	}

	if findings := Validate(s); !findings.Valid() {
		t.Errorf("validating without a directory reported a missing template:\n%s", messages(findings.Fatal()))
	}

	findings := ValidateDir(s, missing)
	if findings.Valid() {
		t.Fatal("a missing template file was not reported")
	}

	got := messages(findings.Fatal())
	if !strings.Contains(got, "specs") {
		t.Errorf("the finding does not name the artifact:\n%s", got)
	}
	if !strings.Contains(got, filepath.Join("templates", "specs", "spec.md")) {
		t.Errorf("the finding does not name the path it looked for:\n%s", got)
	}

	whole, err := Load(complete)
	if err != nil {
		t.Fatalf("loading the fixture: %v", err)
	}
	if findings := ValidateDir(whole, complete); !findings.Valid() {
		t.Errorf("a complete folder does not validate:\n%s", messages(findings.Fatal()))
	}
}

func TestAnArtifactWithNoTemplateIsNotCheckedOnDisk(t *testing.T) {
	t.Parallel()

	s := load(t, "no-template.yaml")

	for _, f := range ValidateDir(s, t.TempDir()).Fatal() {
		if strings.Contains(f.Message, "template is missing") {
			t.Errorf("an artifact declaring no template was checked on disk: %s", f)
		}
	}
}

func TestEveryFindingNamesWhereItIs(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"duplicate.yaml", "unknown-requirement.yaml", "self.yaml", "cycle.yaml", "unknown-gate.yaml", "collides.yaml", "orphan.yaml"} {
		for _, f := range Validate(load(t, name)) {
			if f.Message == "" {
				t.Errorf("%s produced a finding with no message", name)
			}
			if !strings.Contains(f.String(), f.Severity.String()) {
				t.Errorf("%s: a finding does not carry its severity: %s", name, f)
			}
		}
	}
}

func TestSeverityString(t *testing.T) {
	t.Parallel()

	if got := Fatal.String(); got != "fatal" {
		t.Errorf("Fatal = %q", got)
	}
	if got := Warning.String(); got != "warning" {
		t.Errorf("Warning = %q", got)
	}
}

func TestNoFindingsIsValid(t *testing.T) {
	t.Parallel()

	var findings Findings

	if !findings.Valid() {
		t.Error("no findings is not valid")
	}
	if len(findings.Fatal()) != 0 || len(findings.Warnings()) != 0 {
		t.Error("no findings produced findings")
	}
}

func TestAnArtifactWithNoDescriptionIsFatal(t *testing.T) {
	t.Parallel()

	body := `name: n
version: 1
description: a schema
artifacts:
  - id: tasks
    generates: tasks.md
    template: tasks.md
apply:
  requires: [tasks]
  tracks: tasks.md
`

	s, err := Parse("schema.yaml", []byte(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	findings := Validate(s)
	if findings.Valid() {
		t.Fatal("an artifact with no description validates; OpenSpec rejects it")
	}

	got := messages(findings.Fatal())
	if !strings.Contains(got, "tasks") {
		t.Errorf("the finding does not name the artifact:\n%s", got)
	}
	if !strings.Contains(got, "OpenSpec rejects") {
		t.Errorf("the finding does not say whose rule it is:\n%s", got)
	}
}

package schema

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

type Severity int

const (
	Warning Severity = iota
	Fatal
)

func (s Severity) String() string {
	if s == Fatal {
		return "fatal"
	}
	return "warning"
}

type Finding struct {
	Severity Severity
	Artifact string
	Message  string
}

func (f Finding) String() string {
	if f.Artifact == "" {
		return fmt.Sprintf("%s: %s", f.Severity, f.Message)
	}
	return fmt.Sprintf("%s: %s: %s", f.Severity, f.Artifact, f.Message)
}

type Findings []Finding

func (f Findings) Valid() bool {
	return len(f.Fatal()) == 0
}

func (f Findings) Fatal() Findings {
	return f.bySeverity(Fatal)
}

func (f Findings) Warnings() Findings {
	return f.bySeverity(Warning)
}

func (f Findings) bySeverity(want Severity) Findings {
	var out Findings
	for _, finding := range f {
		if finding.Severity == want {
			out = append(out, finding)
		}
	}
	return out
}

func Validate(s Schema) Findings {
	return validate(s, "")
}

func ValidateDir(s Schema, dir string) Findings {
	return validate(s, dir)
}

func validate(s Schema, dir string) Findings {
	var findings Findings

	fatal := func(artifact, format string, args ...any) {
		findings = append(findings, Finding{Severity: Fatal, Artifact: artifact, Message: fmt.Sprintf(format, args...)})
	}
	warn := func(artifact, format string, args ...any) {
		findings = append(findings, Finding{Severity: Warning, Artifact: artifact, Message: fmt.Sprintf(format, args...)})
	}

	if strings.TrimSpace(s.Name) == "" {
		fatal("", "the schema declares no name; the name decides where it installs and what --schema takes")
	}

	if len(s.Artifacts) == 0 {
		fatal("", "the schema declares no artifacts")
	}

	known := make(map[string]bool, len(s.Artifacts))
	seen := make(map[string]bool, len(s.Artifacts))

	for i, a := range s.Artifacts {
		switch {
		case strings.TrimSpace(a.ID) == "":
			fatal("", "the artifact at position %d declares no id", i)
			continue
		case seen[a.ID]:
			fatal(a.ID, "the id is declared more than once")
		}

		seen[a.ID] = true
		known[a.ID] = true
	}

	for _, a := range s.Artifacts {
		if a.ID == "" {
			continue
		}

		for _, req := range a.Requires {
			switch {
			case req == a.ID:
				fatal(a.ID, "it requires itself")
			case !known[req]:
				fatal(a.ID, "it requires %q, which no artifact declares", req)
			}
		}

		if strings.TrimSpace(a.Template) == "" {
			warn(a.ID, "it declares no template, so nothing seeds the file it generates")
		}

		// OpenSpec 1.10.0 rejects an artifact with no description. The
		// briefing calls the field optional; the CLI is the authority on what
		// it will accept, and a schema it rejects cannot be installed.
		if strings.TrimSpace(a.Description) == "" {
			fatal(a.ID, "it declares no description; OpenSpec rejects a schema whose artifact has none")
		}
	}

	findings = append(findings, cycleFindings(s, known)...)

	if len(s.Apply.Requires) == 0 {
		fatal("", "apply.requires names no artifact, so nothing gates implementation")
	}
	for _, id := range s.Apply.Requires {
		if !known[id] {
			fatal("", "apply.requires names %q, which no artifact declares", id)
		}
	}
	if strings.TrimSpace(s.Apply.Tracks) == "" {
		fatal("", "apply.tracks names no file, so nothing records progress")
	}

	findings = append(findings, collisionFindings(s)...)
	findings = append(findings, orphanFindings(s)...)

	if dir != "" {
		findings = append(findings, templateFindings(s, dir)...)
	}

	return findings
}

func cycleFindings(s Schema, known map[string]bool) Findings {
	const (
		white = 0
		grey  = 1
		black = 2
	)

	colour := make(map[string]int, len(s.Artifacts))
	requires := make(map[string][]string, len(s.Artifacts))

	for _, a := range s.Artifacts {
		reqs := make([]string, 0, len(a.Requires))
		for _, req := range a.Requires {
			if known[req] && req != a.ID {
				reqs = append(reqs, req)
			}
		}
		requires[a.ID] = reqs
	}

	var findings Findings
	var stack []string
	reported := make(map[string]bool)

	var walk func(id string)
	walk = func(id string) {
		colour[id] = grey
		stack = append(stack, id)

		for _, req := range requires[id] {
			switch colour[req] {
			case white:
				walk(req)
			case grey:
				cycle := cycleFrom(stack, req)
				key := strings.Join(sorted(cycle), ",")
				if !reported[key] {
					reported[key] = true
					findings = append(findings, Finding{
						Severity: Fatal,
						Message:  "these artifacts require each other in a cycle: " + strings.Join(cycle, " → ") + " → " + cycle[0],
					})
				}
			}
		}

		stack = stack[:len(stack)-1]
		colour[id] = black
	}

	for _, a := range s.Artifacts {
		if a.ID != "" && colour[a.ID] == white {
			walk(a.ID)
		}
	}

	return findings
}

func cycleFrom(stack []string, start string) []string {
	for i, id := range stack {
		if id == start {
			return append([]string{}, stack[i:]...)
		}
	}
	return append([]string{}, stack...)
}

func sorted(in []string) []string {
	out := append([]string{}, in...)
	sort.Strings(out)
	return out
}

func collisionFindings(s Schema) Findings {
	byPath := make(map[string][]string)
	var order []string

	for _, a := range s.Artifacts {
		if a.Generates == "" || a.ID == "" {
			continue
		}
		if len(byPath[a.Generates]) == 0 {
			order = append(order, a.Generates)
		}
		byPath[a.Generates] = append(byPath[a.Generates], a.ID)
	}

	var findings Findings
	for _, path := range order {
		if ids := byPath[path]; len(ids) > 1 {
			findings = append(findings, Finding{
				Severity: Warning,
				Message:  fmt.Sprintf("%s all generate %q, so one overwrites the others", strings.Join(ids, ", "), path),
			})
		}
	}

	return findings
}

func orphanFindings(s Schema) Findings {
	required := make(map[string]bool)
	for _, a := range s.Artifacts {
		for _, req := range a.Requires {
			required[req] = true
		}
	}

	gates := s.Gates()

	var findings Findings
	for _, a := range s.Artifacts {
		if a.ID == "" || required[a.ID] || gates[a.ID] {
			continue
		}
		findings = append(findings, Finding{
			Severity: Warning,
			Artifact: a.ID,
			Message:  "nothing requires it and it does not gate apply, so the workflow never reaches it",
		})
	}

	return findings
}

func templateFindings(s Schema, dir string) Findings {
	var findings Findings

	for _, a := range s.Artifacts {
		if a.ID == "" || strings.TrimSpace(a.Template) == "" {
			continue
		}

		path := a.TemplatePath(dir)
		if _, err := os.Stat(path); err != nil {
			findings = append(findings, Finding{
				Severity: Fatal,
				Artifact: a.ID,
				Message:  fmt.Sprintf("its template is missing; nothing at %s", path),
			})
		}
	}

	return findings
}

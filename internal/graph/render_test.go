package graph

import (
	"errors"
	"strings"
	"testing"

	"github.com/speclib/openspec-schema-manager/internal/schema"
)

func TestDrawAChain(t *testing.T) {
	t.Parallel()

	got, err := build(t, "chain.yaml").Draw(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, want := range []string{"specs", "tasks", "│", "▼"} {
		if !strings.Contains(got, want) {
			t.Errorf("the drawing does not carry %q:\n%s", want, got)
		}
	}
}

func TestDrawMarksAGateDistinctly(t *testing.T) {
	t.Parallel()

	got, err := build(t, "chain.yaml").Draw(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, "╱") && !strings.Contains(got, "╲") {
		t.Errorf("the gate is not drawn in a different shape:\n%s", got)
	}
}

func TestDrawABranchyGraph(t *testing.T) {
	t.Parallel()

	got, err := build(t, "branchy.yaml").Draw(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, want := range []string{"proposal", "specs", "design", "tasks"} {
		if !strings.Contains(got, want) {
			t.Errorf("%s is missing from the drawing:\n%s", want, got)
		}
	}

	if lines := strings.Split(got, "\n"); len(lines) < 5 {
		t.Errorf("a four-artifact graph drew %d lines:\n%s", len(lines), got)
	}
}

func TestDrawInASCII(t *testing.T) {
	t.Parallel()

	got, err := build(t, "chain.yaml").Draw(true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, want := range []string{"specs", "tasks"} {
		if !strings.Contains(got, want) {
			t.Errorf("the ASCII drawing does not carry %q:\n%s", want, got)
		}
	}
	for _, r := range got {
		if r > 127 {
			t.Errorf("the ASCII drawing holds %q:\n%s", r, got)
			break
		}
	}
}

func TestDrawIndependentArtifacts(t *testing.T) {
	t.Parallel()

	got, err := build(t, "wide.yaml").Draw(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, want := range []string{"research", "risks", "budget"} {
		if !strings.Contains(got, want) {
			t.Errorf("%s is missing from the drawing:\n%s", want, got)
		}
	}
}

func TestDrawReportsACycle(t *testing.T) {
	t.Parallel()

	got, err := build(t, "cycle.yaml").Draw(false)
	if !errors.Is(err, ErrCycle) {
		t.Fatalf("error = %v, want ErrCycle", err)
	}
	if got != "" {
		t.Errorf("a drawing was produced for a cyclic schema:\n%s", got)
	}
}

func TestDrawReportsAnEmptyGraph(t *testing.T) {
	t.Parallel()

	got, err := Build(schema.Schema{}).Draw(false)
	if !errors.Is(err, ErrNothingToDraw) {
		t.Fatalf("error = %v, want ErrNothingToDraw", err)
	}
	if got != "" {
		t.Errorf("a drawing was produced for an empty schema:\n%s", got)
	}
}

func TestDrawIsDeterministic(t *testing.T) {
	t.Parallel()

	g := build(t, "branchy.yaml")

	first, err := g.Draw(false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for range 5 {
		again, err := g.Draw(false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if again != first {
			t.Fatalf("the drawing changed between runs:\n%s\nthen\n%s", first, again)
		}
	}
}

func TestDrawHandlesAwkwardIDs(t *testing.T) {
	t.Parallel()

	got, err := build(t, "awkward-ids.yaml").Draw(false)
	if err != nil {
		t.Fatalf("an id Mermaid cannot take broke the drawing: %v", err)
	}
	if !strings.Contains(got, "first step") {
		t.Errorf("the label was lost:\n%s", got)
	}
}

func TestUnicodeAvailable(t *testing.T) {
	cases := []struct {
		name string
		vars map[string]string
		want bool
	}{
		{name: "LANG is UTF-8", vars: map[string]string{"LANG": "en_GB.UTF-8"}, want: true},
		{name: "LANG is utf8", vars: map[string]string{"LANG": "en_US.utf8"}, want: true},
		{name: "LANG is C", vars: map[string]string{"LANG": "C"}, want: false},
		{name: "LC_ALL wins", vars: map[string]string{"LC_ALL": "C", "LANG": "en_GB.UTF-8"}, want: false},
		{name: "LC_CTYPE is consulted", vars: map[string]string{"LC_CTYPE": "nl_NL.UTF-8"}, want: true},
		{name: "nothing is set", vars: nil, want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, key := range []string{"LC_ALL", "LC_CTYPE", "LANG"} {
				t.Setenv(key, "")
			}
			for key, value := range tc.vars {
				t.Setenv(key, value)
			}

			if got := UnicodeAvailable(); got != tc.want {
				t.Errorf("UnicodeAvailable = %v, want %v", got, tc.want)
			}
		})
	}
}

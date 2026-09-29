package graph

import (
	"errors"
	"strings"
	"testing"
)

func mermaid(t *testing.T, fixture string) string {
	t.Helper()

	got, err := build(t, fixture).Mermaid()
	if err != nil {
		t.Fatalf("emitting %s: %v", fixture, err)
	}

	return got
}

func TestMermaidEmitsAChain(t *testing.T) {
	t.Parallel()

	got := mermaid(t, "chain.yaml")

	want := `graph TD
    specs["specs"]
    tasks{{"tasks"}}
    specs --> tasks
`

	if got != want {
		t.Errorf("output:\n%s\nwant:\n%s", got, want)
	}
}

func TestMermaidMarksGatesDistinctly(t *testing.T) {
	t.Parallel()

	got := mermaid(t, "branchy.yaml")

	if !strings.Contains(got, `tasks{{"tasks"}}`) {
		t.Errorf("the gate node does not use the gate shape:\n%s", got)
	}
	for _, plain := range []string{"proposal", "specs", "design"} {
		if !strings.Contains(got, plain+`["`+plain+`"]`) {
			t.Errorf("%s does not use the plain shape:\n%s", plain, got)
		}
	}
}

func TestMermaidDrawsEveryEdge(t *testing.T) {
	t.Parallel()

	got := mermaid(t, "branchy.yaml")

	for _, edge := range []string{
		"proposal --> specs",
		"proposal --> design",
		"specs --> tasks",
		"design --> tasks",
	} {
		if !strings.Contains(got, edge) {
			t.Errorf("the edge %q is missing:\n%s", edge, got)
		}
	}
}

func TestAnArtifactWithNoRequirementsStillAppears(t *testing.T) {
	t.Parallel()

	got := mermaid(t, "wide.yaml")

	for _, id := range []string{"research", "risks", "budget"} {
		if !strings.Contains(got, id+`{{"`+id+`"}}`) {
			t.Errorf("%s is missing from the diagram:\n%s", id, got)
		}
	}
	if strings.Contains(got, "-->") {
		t.Errorf("edges were drawn for a schema with no requirements:\n%s", got)
	}
}

func TestMermaidIsDeterministic(t *testing.T) {
	t.Parallel()

	g := build(t, "branchy.yaml")

	first, err := g.Mermaid()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for range 20 {
		again, err := g.Mermaid()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if again != first {
			t.Fatalf("the output changed between runs:\n%s\nthen\n%s", first, again)
		}
	}
}

func TestMermaidFollowsTheStableOrder(t *testing.T) {
	t.Parallel()

	got := mermaid(t, "branchy.yaml")

	positions := make([]int, 0, 4)
	for _, id := range []string{"proposal", "specs", "design", "tasks"} {
		at := strings.Index(got, "    "+id)
		if at < 0 {
			t.Fatalf("%s is missing:\n%s", id, got)
		}
		positions = append(positions, at)
	}

	for i := 1; i < len(positions); i++ {
		if positions[i] < positions[i-1] {
			t.Errorf("the nodes are not declared in the graph's order:\n%s", got)
		}
	}
}

func TestAnIDThatIsNotASafeIdentifier(t *testing.T) {
	t.Parallel()

	got := mermaid(t, "awkward-ids.yaml")

	if !strings.HasPrefix(got, "graph TD\n") {
		t.Errorf("the output does not open with graph TD:\n%s", got)
	}

	for _, want := range []string{
		`n0["first step"]`,
		`n1["say #quot;hi#quot;"]`,
		`n2{{"a[b];c"}}`,
		"n0 --> n1",
		"n1 --> n2",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("the output does not carry %q:\n%s", want, got)
		}
	}

	for _, unsafe := range []string{"    first step[", `    say "hi"`} {
		if strings.Contains(got, unsafe) {
			t.Errorf("an unsafe id reached the identifier position:\n%s", got)
		}
	}
}

func TestIDsThatWouldCollideStayDistinct(t *testing.T) {
	t.Parallel()

	g := Graph{
		nodes: []Node{
			{ID: "a b", index: 0},
			{ID: "a!b", index: 1, Requires: []string{"a b"}},
		},
		byID: map[string]int{"a b": 0, "a!b": 1},
	}

	got, err := g.Mermaid()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(got, `n0["a b"]`) || !strings.Contains(got, `n1["a!b"]`) {
		t.Errorf("two awkward ids did not get distinct identifiers:\n%s", got)
	}
	if !strings.Contains(got, "n0 --> n1") {
		t.Errorf("the edge between them is missing:\n%s", got)
	}
}

func TestANewlineInAnIDDoesNotBreakTheDiagram(t *testing.T) {
	t.Parallel()

	g := Graph{
		nodes: []Node{{ID: "one\ntwo", index: 0}},
		byID:  map[string]int{"one\ntwo": 0},
	}

	got, err := g.Mermaid()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Count(got, "\n") != 2 {
		t.Errorf("a newline in an id added a line to the diagram:\n%q", got)
	}
	if !strings.Contains(got, `n0["one two"]`) {
		t.Errorf("the label was not flattened:\n%s", got)
	}
}

func TestMermaidReportsACycle(t *testing.T) {
	t.Parallel()

	got, err := build(t, "cycle.yaml").Mermaid()
	if !errors.Is(err, ErrCycle) {
		t.Fatalf("error = %v, want ErrCycle", err)
	}
	if got != "" {
		t.Errorf("a diagram was emitted for a cyclic schema:\n%s", got)
	}
}

func TestAnEmptyGraphEmitsAValidHeader(t *testing.T) {
	t.Parallel()

	got, err := Graph{byID: map[string]int{}}.Mermaid()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "graph TD\n" {
		t.Errorf("output = %q", got)
	}
}

func TestAnUnknownRequirementDrawsNoEdge(t *testing.T) {
	t.Parallel()

	got := mermaid(t, "unknown-requirement.yaml")

	if strings.Contains(got, "-->") {
		t.Errorf("an edge was drawn to an artifact that does not exist:\n%s", got)
	}
}

package graph

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/speclib/openspec-schema-manager/internal/schema"
)

func build(t *testing.T, name string) Graph {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join("..", "schema", "testdata", name))
	if err != nil {
		t.Fatalf("reading the fixture: %v", err)
	}

	s, err := schema.Parse(name, raw)
	if err != nil {
		t.Fatalf("parsing %s: %v", name, err)
	}

	return Build(s)
}

func ids(nodes []Node) []string {
	out := make([]string, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, n.ID)
	}
	return out
}

func TestBuildFromAChain(t *testing.T) {
	t.Parallel()

	g := build(t, "chain.yaml")

	if g.Len() != 2 {
		t.Fatalf("got %d nodes, want 2", g.Len())
	}
	if got := g.Edges(); !reflect.DeepEqual(got, []Edge{{From: "specs", To: "tasks"}}) {
		t.Errorf("edges = %v, want one from specs to tasks", got)
	}
	if g.Track != "tasks.md" {
		t.Errorf("Track = %q", g.Track)
	}

	tasks, ok := g.Node("tasks")
	if !ok {
		t.Fatal("tasks is missing from the graph")
	}
	if !tasks.Gate {
		t.Error("tasks gates apply but the node does not say so")
	}
	if tasks.Generates != "tasks.md" || tasks.Template != "tasks.md" {
		t.Errorf("the node lost what the artifact declared: %+v", tasks)
	}

	specs, _ := g.Node("specs")
	if specs.Gate {
		t.Error("specs does not gate apply but the node says it does")
	}

	if _, ok := g.Node("absent"); ok {
		t.Error("Node found an artifact that is not declared")
	}
}

func TestBuildFromABranchAndAJoin(t *testing.T) {
	t.Parallel()

	g := build(t, "branchy.yaml")

	want := []Edge{
		{From: "proposal", To: "specs"},
		{From: "proposal", To: "design"},
		{From: "specs", To: "tasks"},
		{From: "design", To: "tasks"},
	}

	if got := g.Edges(); !reflect.DeepEqual(got, want) {
		t.Errorf("edges = %v, want %v", got, want)
	}
}

func TestOrderIsStableAndFollowsDeclarationOrder(t *testing.T) {
	t.Parallel()

	g := build(t, "branchy.yaml")

	first, err := g.Order()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"proposal", "specs", "design", "tasks"}
	if got := ids(first); !reflect.DeepEqual(got, want) {
		t.Errorf("order = %v, want %v", got, want)
	}

	for range 20 {
		again, err := g.Order()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(ids(again), want) {
			t.Fatalf("the order changed between runs: %v", ids(again))
		}
	}
}

func TestOrderOfIndependentArtifactsFollowsDeclaration(t *testing.T) {
	t.Parallel()

	g := build(t, "wide.yaml")

	got, err := g.Order()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"research", "risks", "budget"}
	if !reflect.DeepEqual(ids(got), want) {
		t.Errorf("order = %v, want %v", ids(got), want)
	}
}

func TestOrderReportsACycle(t *testing.T) {
	t.Parallel()

	g := build(t, "cycle.yaml")

	got, err := g.Order()
	if !errors.Is(err, ErrCycle) {
		t.Fatalf("error = %v, want ErrCycle", err)
	}
	if got != nil {
		t.Errorf("a partial order was returned: %v", ids(got))
	}
}

func TestASelfReferenceDoesNotBlockOrdering(t *testing.T) {
	t.Parallel()

	g := build(t, "self.yaml")

	got, err := g.Order()
	if err != nil {
		t.Fatalf("a self-reference blocked ordering: %v", err)
	}
	if !reflect.DeepEqual(ids(got), []string{"tasks"}) {
		t.Errorf("order = %v", ids(got))
	}
}

func TestAnUnknownRequirementIsIgnoredByTheGraph(t *testing.T) {
	t.Parallel()

	g := build(t, "unknown-requirement.yaml")

	if len(g.Edges()) != 0 {
		t.Errorf("an edge was built to an artifact that does not exist: %v", g.Edges())
	}

	got, err := g.Order()
	if err != nil {
		t.Fatalf("an unknown requirement blocked ordering: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("order = %v", ids(got))
	}
}

func TestMetrics(t *testing.T) {
	t.Parallel()

	cases := []struct {
		fixture string
		want    Metrics
	}{
		{fixture: "chain.yaml", want: Metrics{Artifacts: 2, LongestChain: 2, Gates: 1, Leaves: 1}},
		{fixture: "single.yaml", want: Metrics{Artifacts: 1, LongestChain: 1, Gates: 1, Leaves: 1}},
		{fixture: "wide.yaml", want: Metrics{Artifacts: 3, LongestChain: 1, Gates: 3, Leaves: 3}},
		{fixture: "branchy.yaml", want: Metrics{Artifacts: 4, LongestChain: 3, Gates: 1, Leaves: 1}},
		{fixture: "spec-driven.yaml", want: Metrics{Artifacts: 4, LongestChain: 3, Gates: 1, Leaves: 1}},
		{fixture: "orphan.yaml", want: Metrics{Artifacts: 3, LongestChain: 2, Gates: 1, Leaves: 2}},
	}

	for _, tc := range cases {
		t.Run(tc.fixture, func(t *testing.T) {
			t.Parallel()

			got, err := build(t, tc.fixture).Metrics()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("metrics = %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestMetricsReportsACycle(t *testing.T) {
	t.Parallel()

	if _, err := build(t, "cycle.yaml").Metrics(); !errors.Is(err, ErrCycle) {
		t.Fatalf("error = %v, want ErrCycle", err)
	}
}

func TestGatesAreReported(t *testing.T) {
	t.Parallel()

	if got := ids(build(t, "wide.yaml").Gates()); !reflect.DeepEqual(got, []string{"research", "risks", "budget"}) {
		t.Errorf("gates = %v", got)
	}
	if got := ids(build(t, "chain.yaml").Gates()); !reflect.DeepEqual(got, []string{"tasks"}) {
		t.Errorf("gates = %v", got)
	}
}

func TestNodesAreCopied(t *testing.T) {
	t.Parallel()

	g := build(t, "chain.yaml")

	nodes := g.Nodes()
	nodes[0].ID = "tampered"

	if again := g.Nodes(); again[0].ID != "specs" {
		t.Error("Nodes handed out the graph's own slice")
	}
}

func TestIDsAndSummary(t *testing.T) {
	t.Parallel()

	g := build(t, "branchy.yaml")

	if got := g.IDs(); !reflect.DeepEqual(got, []string{"proposal", "specs", "design", "tasks"}) {
		t.Errorf("IDs = %v", got)
	}
	if got := g.Summary(); !strings.Contains(got, "proposal") || !strings.Contains(got, "tasks") {
		t.Errorf("Summary = %q", got)
	}
}

func TestAnEmptyGraph(t *testing.T) {
	t.Parallel()

	g := Build(schema.Schema{})

	if g.Len() != 0 {
		t.Errorf("an empty schema gave %d nodes", g.Len())
	}

	got, err := g.Order()
	if err != nil {
		t.Fatalf("ordering an empty graph failed: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("order = %v", ids(got))
	}

	metrics, err := g.Metrics()
	if err != nil {
		t.Fatalf("measuring an empty graph failed: %v", err)
	}
	if metrics != (Metrics{}) {
		t.Errorf("metrics = %+v, want all zero", metrics)
	}
}

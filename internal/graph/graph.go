package graph

import (
	"errors"
	"strings"

	"github.com/speclib/openspec-schema-manager/internal/schema"
)

var ErrCycle = errors.New("the artifacts require each other in a cycle")

type Node struct {
	ID        string
	Generates string
	Template  string
	Gate      bool
	Requires  []string
	index     int
}

type Graph struct {
	nodes []Node
	byID  map[string]int
	Track string
}

func Build(s schema.Schema) Graph {
	g := Graph{
		nodes: make([]Node, 0, len(s.Artifacts)),
		byID:  make(map[string]int, len(s.Artifacts)),
		Track: s.Apply.Tracks,
	}

	gates := s.Gates()

	for i, a := range s.Artifacts {
		g.byID[a.ID] = i
		g.nodes = append(g.nodes, Node{
			ID:        a.ID,
			Generates: a.Generates,
			Template:  a.Template,
			Gate:      gates[a.ID],
			Requires:  append([]string{}, a.Requires...),
			index:     i,
		})
	}

	return g
}

func (g Graph) Nodes() []Node {
	return append([]Node{}, g.nodes...)
}

func (g Graph) Len() int { return len(g.nodes) }

func (g Graph) Node(id string) (Node, bool) {
	i, ok := g.byID[id]
	if !ok {
		return Node{}, false
	}
	return g.nodes[i], true
}

type Edge struct {
	From string
	To   string
}

func (g Graph) Edges() []Edge {
	var edges []Edge

	for _, n := range g.nodes {
		for _, req := range n.Requires {
			if _, ok := g.byID[req]; ok {
				edges = append(edges, Edge{From: req, To: n.ID})
			}
		}
	}

	return edges
}

func (g Graph) Order() ([]Node, error) {
	indegree := make([]int, len(g.nodes))
	dependents := make([][]int, len(g.nodes))

	for _, n := range g.nodes {
		for _, req := range n.Requires {
			from, ok := g.byID[req]
			if !ok || from == n.index {
				continue
			}
			indegree[n.index]++
			dependents[from] = append(dependents[from], n.index)
		}
	}

	var ready []int
	for i := range g.nodes {
		if indegree[i] == 0 {
			ready = append(ready, i)
		}
	}

	ordered := make([]Node, 0, len(g.nodes))

	for len(ready) > 0 {
		next := 0
		for i, candidate := range ready {
			if candidate < ready[next] {
				next = i
			}
		}

		current := ready[next]
		ready = append(ready[:next], ready[next+1:]...)
		ordered = append(ordered, g.nodes[current])

		for _, dependent := range dependents[current] {
			indegree[dependent]--
			if indegree[dependent] == 0 {
				ready = append(ready, dependent)
			}
		}
	}

	if len(ordered) != len(g.nodes) {
		return nil, ErrCycle
	}

	return ordered, nil
}

func (g Graph) Gates() []Node {
	var gates []Node
	for _, n := range g.nodes {
		if n.Gate {
			gates = append(gates, n)
		}
	}
	return gates
}

type Metrics struct {
	Artifacts    int
	LongestChain int
	Gates        int
	Leaves       int
}

func (g Graph) Metrics() (Metrics, error) {
	ordered, err := g.Order()
	if err != nil {
		return Metrics{}, err
	}

	depth := make(map[string]int, len(ordered))
	longest := 0

	for _, n := range ordered {
		best := 1
		for _, req := range n.Requires {
			if d, ok := depth[req]; ok && d+1 > best {
				best = d + 1
			}
		}
		depth[n.ID] = best
		if best > longest {
			longest = best
		}
	}

	required := make(map[string]bool, len(g.nodes))
	for _, e := range g.Edges() {
		required[e.From] = true
	}

	leaves := 0
	for _, n := range g.nodes {
		if !required[n.ID] {
			leaves++
		}
	}

	return Metrics{
		Artifacts:    len(g.nodes),
		LongestChain: longest,
		Gates:        len(g.Gates()),
		Leaves:       leaves,
	}, nil
}

func (g Graph) IDs() []string {
	ids := make([]string, 0, len(g.nodes))
	for _, n := range g.nodes {
		ids = append(ids, n.ID)
	}
	return ids
}

func (g Graph) Summary() string {
	return strings.Join(g.IDs(), " · ")
}

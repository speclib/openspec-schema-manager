package graph

import (
	"fmt"
	"regexp"
	"strings"
)

var safeIdentifier = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]*$`)

func (g Graph) Mermaid() (string, error) {
	ordered, err := g.Order()
	if err != nil {
		return "", err
	}

	identifiers := make(map[string]string, len(g.nodes))
	for _, n := range g.nodes {
		identifiers[n.ID] = identifier(n)
	}

	var b strings.Builder
	b.WriteString("graph TD\n")

	for _, n := range ordered {
		b.WriteString("    ")
		b.WriteString(identifiers[n.ID])
		if n.Gate {
			b.WriteString(`{{"` + escape(n.ID) + `"}}`)
		} else {
			b.WriteString(`["` + escape(n.ID) + `"]`)
		}
		b.WriteString("\n")
	}

	for _, n := range ordered {
		for _, req := range n.Requires {
			if _, ok := g.byID[req]; !ok {
				continue
			}
			b.WriteString(fmt.Sprintf("    %s --> %s\n", identifiers[req], identifiers[n.ID]))
		}
	}

	return b.String(), nil
}

func identifier(n Node) string {
	if safeIdentifier.MatchString(n.ID) {
		return n.ID
	}
	return fmt.Sprintf("n%d", n.index)
}

func escape(s string) string {
	replacer := strings.NewReplacer(
		`"`, `#quot;`,
		"\n", " ",
		"\r", " ",
	)
	return replacer.Replace(s)
}
